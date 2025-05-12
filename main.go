package main

import (
    "context"
    "fmt"
    "time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/watch"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

    nomad "github.com/hashicorp/nomad/api"
)


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

func syncNomadState(spec NomadStatefulWorkloadSpec, jobName string, nomadClient *nomad.Client) (string, error) {
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


func main() {
    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err)
    }
    dyn, err := dynamic.NewForConfig(config)
    if err != nil {
        panic(err)
    }

    nomadClient, err := nomad.NewClient(nomad.DefaultConfig())
    if err != nil {
        panic(err)
    }

    // TODO: update parameters here
    gvr := schema.GroupVersionResource{
        Group:    "your.group",
        Version:  "v1alpha1",
        Resource: "nomadstatefulworkloads",
    }

    for {
        w, err := dyn.Resource(gvr).Namespace("").Watch(context.Background(), metav1.ListOptions{})
        if err != nil {
            fmt.Println("Error in watch:", err)
            time.Sleep(5 * time.Second)
            continue
        }
        ch := w.ResultChan()
        for event := range ch {
            obj := event.Object.(*unstructured.Unstructured)
            name := obj.GetName()
            specMap, found, _ := unstructured.NestedMap(obj.Object, "spec")
            if !found {
                continue
            }
            specBytes, _ := json.Marshal(specMap)
            var spec NomadStatefulWorkloadSpec
            json.Unmarshal(specBytes, &spec)

            jobID, err := syncNomadState(spec, name, nomadClient)
            if err != nil {
                fmt.Printf("Error in synchronization for %s: %v\n", name, err)
                continue
            }

            status := map[string]interface{}{
                "jobId": jobID,
            }
            _, err = dyn.Resource(gvr).Namespace(obj.GetNamespace()).UpdateStatus(
                context.Background(),
                &unstructured.Unstructured{
                    Object: map[string]interface{}{
                        "apiVersion": obj.GetAPIVersion(),
                        "kind":       obj.GetKind(),
                        "metadata": map[string]interface{}{
                            "name":      name,
                            "namespace": obj.GetNamespace(),
                        },
                        "status": status,
                    },
                },
                metav1.UpdateOptions{},
            )
            if err != nil {
                fmt.Printf("Error in CRD status update: %v\n", err)
            }
        }
        fmt.Println("Watch channel is closed, reconnecting...")
    }
}
