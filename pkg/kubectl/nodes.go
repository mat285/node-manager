package kubectl

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/mat285/node-manager/pkg/log"
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

func GetNodes(ctx context.Context) ([]string, error) {
	output, err := exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `jsonpath="{.items[*].metadata.name}"`).Output()
	if err != nil {
		return nil, err
	}
	out := strings.TrimSpace(string(output))
	out = strings.Trim(out, "\"")
	out = strings.TrimSpace(out)
	nodes := strings.Split(out, " ")
	return nodes, nil
}

func LabelNode(ctx context.Context, name string, key string, value string) error {
	output, err := exec.CommandContext(ctx, "kubectl", "label", "node", name, key+"="+value).Output()
	if err != nil {
		log.GetLogger(ctx).Infof("Error labeling node %s: %v: %s", name, err, string(output))
		return err
	}
	return nil
}
