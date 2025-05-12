package nomadapply

import (
    "fmt"
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
