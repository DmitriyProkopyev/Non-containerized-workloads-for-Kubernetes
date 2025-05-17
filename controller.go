package main

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

// ------------------ Affinity Types and Nomad Spec ------------------

type Affinity struct {
    NodeAffinity struct {
        RequiredDuringSchedulingIgnoredDuringExecution struct {
            NodeSelectorTerms []struct {
                MatchExpressions []struct {
                    Key      string   `json:"key"`
                    Operator string   `json:"operator"`
                    Values   []string `json:"values"`
                } `json:"matchExpressions"`
            } `json:"nodeSelectorTerms"`
        } `json:"requiredDuringSchedulingIgnoredDuringExecution"`
    } `json:"nodeAffinity"`
}

type NomadStatefulWorkloadSpec struct {
    Replicas int     `json:"replicas"`
    Affinity Affinity `json:"affinity"`
}

// ------------------ Utility Functions ------------------

func mapOperator(op string) string {
    switch op {
    case "In":
        return "="
    case "NotIn":
        return "!="
    default:
        return "="
    }
}

func affinityToConstraints(affinity Affinity) []*nomad.Constraint {
    var constraints []*nomad.Constraint
    terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
    for _, term := range terms {
        for _, expr := range term.MatchExpressions {
            for _, v := range expr.Values {
                c := &nomad.Constraint{
                    LTarget: fmt.Sprintf("${node.meta.%s}", expr.Key),
                    Operand: mapOperator(expr.Operator),
                    RTarget: v,
                }
                constraints = append(constraints, c)
            }
        }
    }
    return constraints
}

// ------------------ Apply Desired State ------------------

func ApplyNomadDesiredState(spec NomadStatefulWorkloadSpec, jobName string, nomadClient *nomad.Client) (string, error) {
    job, _, err := nomadClient.Jobs().Info(jobName, nil)
    if err != nil || job == nil {
        newJob := nomad.NewServiceJob(jobName, jobName, "default", 100)
        tg := nomad.TaskGroup{Name: &jobName, Count: &spec.Replicas}
        tg.Constraints = affinityToConstraints(spec.Affinity)
        newJob.TaskGroups = []*nomad.TaskGroup{&tg}
        _, _, err := nomadClient.Jobs().Register(newJob, nil)
        if err != nil {
            return "", fmt.Errorf("Nomad Job creation was not successful: %v", err)
        }
        return *newJob.ID, nil
    }

    needsUpdate := false
    if len(job.TaskGroups) == 0 || job.TaskGroups[0].Count == nil || *job.TaskGroups[0].Count != spec.Replicas {
        needsUpdate = true
        job.TaskGroups[0].Count = &spec.Replicas
    }

    newConstraints := affinityToConstraints(spec.Affinity)
    if len(job.TaskGroups[0].Constraints) != len(newConstraints) {
        needsUpdate = true
        job.TaskGroups[0].Constraints = newConstraints
    }

    if needsUpdate {
        _, _, err := nomadClient.Jobs().Register(job, nil)
        if err != nil {
            return "", fmt.Errorf("Nomad Job updation was not successful: %v", err)
        }
    }

    return *job.ID, nil
}

// ------------------ CRD Status Reconciliation ------------------

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
    nomadClient, err := nomad.NewClient(nomad.DefaultConfig())
    if err != nil {
        log.Printf("❌ Failed to connect to Nomad: %v", err)
        return
    }

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

    _, err = dynClient.Resource(nomadGVR).Namespace(namespace).Patch(context.TODO(), jobName, types.MergePatchType, patchBytes, metav1.PatchOptions{}, "status")
    if err != nil {
        log.Printf("❌ Failed to patch CRD status: %v", err)
        return
    }

    log.Printf("✅ Updated status for %s", jobName)
}
