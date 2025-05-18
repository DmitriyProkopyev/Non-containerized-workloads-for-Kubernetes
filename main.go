package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var nomadGVR = schema.GroupVersionResource{
	Group:    "nomad.hashicorp.com",
	Version:  "v1alpha1",
	Resource: "nomadstatefulworkloads",
}

type NomadWorkloadSpec struct {
	Replicas int `json:"replicas"`
	Resources struct {
		CPU    int `json:"cpu"`
		Memory int `json:"memory"`
	} `json:"resources"`
}

func main() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("❌ failed to get cluster config: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("❌ failed to create dynamic client: %v", err)
	}

	watcher, err := dynClient.Resource(nomadGVR).Namespace("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("❌ failed to set up watch: %v", err)
	}

	log.Println("🚀 Watching NomadStatefulWorkload CRDs...")
	for event := range watcher.ResultChan() {
		if event.Type == watch.Added || event.Type == watch.Modified {
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				log.Println("❌ failed to cast event.Object to *unstructured.Unstructured")
				continue
			}
			handleEvent(dynClient, obj)
		}
	}
}

func handleEvent(dynClient dynamic.Interface, obj *unstructured.Unstructured) {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	specRaw, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		log.Printf("❌ failed to get spec: %v", err)
		return
	}

	specBytes, err := json.Marshal(specRaw)
	if err != nil {
		log.Printf("❌ failed to marshal spec: %v", err)
		return
	}

	var spec NomadWorkloadSpec
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		log.Printf("❌ failed to unmarshal spec: %v", err)
		return
	}

	log.Printf("📦 Desired state - Replicas: %d, CPU: %d, Mem: %d", spec.Replicas, spec.Resources.CPU, spec.Resources.Memory)

	actualReplicas := fetchNomadState(name)

	if actualReplicas != spec.Replicas {
		log.Printf("🔁 Reconciling %s: actual replicas %d ≠ desired %d", name, actualReplicas, spec.Replicas)
		ReconcileNomadStatus(dynClient, name, namespace)
	}
}

func fetchNomadState(name string) int {
	log.Printf("📡 Fetching actual state for job %s from Nomad...", name)
	time.Sleep(1 * time.Second)
	return 0
}
