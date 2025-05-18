package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"

    nomad "github.com/hashicorp/nomad/api"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
)

// ------------------ CRD Spec Types ------------------

type Resources struct {
    Cpu    int `json:"cpu"`
    Memory int `json:"memory"`
}

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

type TaskSpec struct {
    Name   string                 `json:"name"`
    Driver string                 `json:"driver"`
    Config map[string]interface{} `json:"config"`
}

type NomadStatefulWorkloadSpec struct {
    Replicas  int       `json:"replicas"`
    Resources Resources `json:"resources"`
    Affinity  Affinity  `json:"affinity"`
    Task      TaskSpec  `json:"task"`
}

// ------------------ Affinity Translation ------------------

func affinityToConstraints(affinity Affinity) []*nomad.Constraint {
    var constraints []*nomad.Constraint
    terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

    for _, term := range terms {
        for _, expr := range term.MatchExpressions {
            key := expr.Key
            values := expr.Values

            switch expr.Operator {
            case "In", "NotIn":
                if len(values) == 0 {
                    continue
                }
                operand := "regexp"
                if expr.Operator == "NotIn" {
                    operand = "not_regexp"
                }
                constraints = append(constraints, &nomad.Constraint{
                    LTarget: fmt.Sprintf("${node.meta.%s}", key),
                    Operand: operand,
                    RTarget: strings.Join(values, "|"),
                })

            case "Exists":
                constraints = append(constraints, &nomad.Constraint{
                    LTarget: fmt.Sprintf("${node.meta.%s}", key),
                    Operand: "is_set",
                })

            case "DoesNotExist":
                constraints = append(constraints, &nomad.Constraint{
                    LTarget: fmt.Sprintf("${node.meta.%s}", key),
                    Operand: "not_is_set",
                })

            default:
                log.Printf("Unsupported operator: %s", expr.Operator)
            }
        }
    }
    return constraints
}

// ------------------ Job Management ------------------

func ApplyNomadDesiredState(spec NomadStatefulWorkloadSpec, jobName, namespace string, nomadClient *nomad.Client) (string, error) {
    job, _, err := nomadClient.Jobs().Info(jobName, &nomad.QueryOptions{Namespace: namespace})
    if err != nil || job == nil {
        return createNewNomadJob(spec, jobName, namespace, nomadClient)
    }
    return updateExistingNomadJob(spec, jobName, namespace, job, nomadClient)
}

func createNewNomadJob(spec NomadStatefulWorkloadSpec, jobName, namespace string, nomadClient *nomad.Client) (string, error) {
    newJob := nomad.NewServiceJob(jobName, jobName, namespace, 100)

    region := os.Getenv("NOMAD_REGION")
    if region == "" {
        region = "dc1"
    }
    newJob.Region = &region

    task := &nomad.Task{
        Name:   spec.Task.Name,
        Driver: spec.Task.Driver,
        Config: spec.Task.Config,
        Resources: &nomad.Resources{
            CPU:      &spec.Resources.Cpu,
            MemoryMB: &spec.Resources.Memory,
        },
    }

    tg := nomad.TaskGroup{
        Name:        &jobName,
        Count:       &spec.Replicas,
        Constraints: affinityToConstraints(spec.Affinity),
        Tasks:       []*nomad.Task{task},
    }

    newJob.TaskGroups = []*nomad.TaskGroup{&tg}

    resp, _, err := nomadClient.Jobs().Register(newJob, &nomad.WriteOptions{
        Region: region,
    })
    if err != nil {
        return "", fmt.Errorf("job creation failed: %w", err)
    }
    return resp.EvalID, nil
}

func updateExistingNomadJob(spec NomadStatefulWorkloadSpec, jobName, namespace string, job *nomad.Job, nomadClient *nomad.Client) (string, error) {
    needsUpdate := false
    tg := job.TaskGroups[0]

    region := os.Getenv("NOMAD_REGION")
    if region == "" {
        region = "dc1"
    }
    job.Region = &region

    if len(tg.Tasks) == 0 {
        tg.Tasks = []*nomad.Task{{
            Name:   spec.Task.Name,
            Driver: spec.Task.Driver,
            Config: spec.Task.Config,
        }}
        needsUpdate = true
    }

    task := tg.Tasks[0]

    if task.Driver != spec.Task.Driver {
        task.Driver = spec.Task.Driver
        needsUpdate = true
    }

    if !equalMaps(task.Config, spec.Task.Config) {
        task.Config = spec.Task.Config
        needsUpdate = true
    }

    if task.Resources == nil {
        task.Resources = &nomad.Resources{}
        needsUpdate = true
    }
    if task.Resources.CPU == nil || *task.Resources.CPU != spec.Resources.Cpu {
        task.Resources.CPU = &spec.Resources.Cpu
        needsUpdate = true
    }
    if task.Resources.MemoryMB == nil || *task.Resources.MemoryMB != spec.Resources.Memory {
        task.Resources.MemoryMB = &spec.Resources.Memory
        needsUpdate = true
    }

    if needsUpdate {
        resp, _, err := nomadClient.Jobs().Register(job, &nomad.WriteOptions{
            Region: region,
        })
        if err != nil {
            return "", fmt.Errorf("job update failed: %w", err)
        }
        return resp.EvalID, nil
    }
    return *job.ID, nil
}

func equalMaps(a, b map[string]interface{}) bool {
    if len(a) != len(b) {
        return false
    }
    for k, v := range a {
        if b[k] != v {
            return false
        }
    }
    return true
}

// ------------------ Status Reconciliation ------------------

func ReconcileNomadStatus(dynClient dynamic.Interface, name, namespace string) {
    cfg := nomad.DefaultConfig()
    cfg.Address = os.Getenv("NOMAD_ADDR")
    cfg.Region = os.Getenv("NOMAD_REGION")

    if cfg.Address == "" {
        cfg.Address = "http://localhost:4646"
    }
    if cfg.Region == "" {
        cfg.Region = "dc1"
    }

    client, err := nomad.NewClient(cfg)
    if err != nil {
        log.Printf("❌ Failed to create Nomad client: %v", err)
        return
    }

    u, err := dynClient.Resource(schema.GroupVersionResource{
        Group:    "nomad.hashicorp.com",
        Version:  "v1alpha1",
        Resource: "nomadstatefulworkloads",
    }).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
    if err != nil {
        log.Printf("❌ Failed to get CR %s: %v", name, err)
        return
    }

    specData, err := json.Marshal(u.Object["spec"])
    if err != nil {
        log.Printf("❌ Failed to marshal spec for %s: %v", name, err)
        return
    }

    var spec NomadStatefulWorkloadSpec
    if err := json.Unmarshal(specData, &spec); err != nil {
        log.Printf("❌ Failed to unmarshal spec for %s: %v", name, err)
        return
    }

    evalID, err := ApplyNomadDesiredState(spec, name, namespace, client)
    if err != nil {
        log.Printf("❌ Failed to apply desired state for %s: %v", name, err)
        return
    }

    log.Printf("✅ Nomad job %s reconciled successfully (Driver: %s, EvalID: %s)",
        name, spec.Task.Driver, evalID)
}
