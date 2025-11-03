package kubectl

import (
	"context"
	"encoding/json"
	"os/exec"
)

type Node struct {
	Metadata NodeMetadata `json:"metadata"`
	Spec     NodeSpec     `json:"spec"`
}

type NodeMetadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type NodeSpec struct {
}

func GetNode(ctx context.Context, name string) (*Node, error) {
	output, err := exec.CommandContext(ctx, "kubectl", "get", "node", name, "-o", `json`).Output()
	if err != nil {
		return nil, err
	}
	var node Node
	err = json.Unmarshal(output, &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}
