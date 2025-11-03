package kubectl

import (
	"context"
	"encoding/json"
	"os/exec"
)

type PodList struct {
	Items []Pod `json:"items"`
}

type Pod struct {
	Metadata PodMetadata `json:"metadata"`
}

type PodMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

func GetPodsForNode(ctx context.Context, node string) ([]Pod, error) {
	output, err := exec.CommandContext(ctx, "kubectl", "get", "pods", "--all-namespaces", "-o", "json", "--field-selector", "spec.nodeName="+node).Output()
	if err != nil {
		return nil, err
	}
	var pods PodList
	err = json.Unmarshal(output, &pods)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}
