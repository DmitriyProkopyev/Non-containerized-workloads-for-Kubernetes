package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	nomad "github.com/hashicorp/nomad/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/dynamic"
)

var nomadGVR = schema.GroupVersionResource{
	Group:    "nomad.hashicorp.com",
	Version:  "v1alpha1",
	Resource: "nomadstatefulworkloads",
}

type StatusUpdate struct {
	Status struct {
		JobStatus string `json:"jobStatus"`
	} `json:"status"`
}

func ReconcileNomadStatus(dynClient dynamic.Interface, jobName string, namespace string) {
	// ПОдключение к Nomad
	nomadClient, err := nomad.NewClient(nomad.DefaultConfig())
	if err != nil {
		log.Printf("❌ Failed to connect to Nomad: %v", err)
		return
	}

	// Получаем allocations для job
	allocs, _, err := nomadClient.Jobs().Allocations(jobName, false, nil)
	if err != nil {
		log.Printf("❌ Failed to get allocations for job %s: %v", jobName, err)
		return
	}

	running := 0
	failed := 0
	for _, alloc := range allocs {
		switch alloc.ClientStatus {
		case "running":
			running++
		case "failed":
			failed++
		}
	}

	jobStatus := fmt.Sprintf("Running: %d, Failed: %d", running, failed)
	log.Printf("📊 Job %s status: %s", jobName, jobStatus)

	// Чтение существующего CRD-объект
	res, err := dynClient.Resource(nomadGVR).Namespace(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		log.Printf("❌ Failed to get CRD %s: %v", jobName, err)
		return
	}

	originalJSON, err := res.MarshalJSON()
	if err != nil {
		log.Printf("❌ Failed to marshal original CRD: %v", err)
		return
	}

	// Создание обновления
	statusUpdate := StatusUpdate{}
	statusUpdate.Status.JobStatus = jobStatus
	statusJSON, err := json.Marshal(statusUpdate)
	if err != nil {
		log.Printf("❌ Failed to marshal status: %v", err)
		return
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(originalJSON, statusJSON, res.Object)
	if err != nil {
		log.Printf("❌ Failed to create merge patch: %v", err)
		return
	}

	// Применение обновления
	_, err = dynClient.Resource(nomadGVR).Namespace(namespace).Patch(context.TODO(), jobName, types.MergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if err != nil {
		log.Printf("❌ Failed to patch CRD status: %v", err)
		return
	}

	log.Printf("✅ Updated status for %s", jobName)
	return
}
