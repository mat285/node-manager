package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/mat285/node-manager/pkg/kubectl"
)

var (
	command = `sh -c "$(curl -fsSL https://github.com/mat285/node-manager/releases/download/%s/install.sh)"`
)

func main() {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lock := &sync.Mutex{}
	nodes, err := kubectl.GetNodes(ctx)
	if err != nil {
		log.Fatalf("Error getting nodes: %v", err)
		os.Exit(1)
	}

	version, err := getVersion()
	if err != nil {
		log.Fatalf("Error getting version: %v", err)
		os.Exit(1)
	}

	command = fmt.Sprintf(command, version)

	for _, node := range nodes {
		wg.Add(1)
		func(node string) {
			defer wg.Done()
			lock.Lock()
			fmt.Println("Setting up node:", node)
			lock.Unlock()
			cmd := exec.Command("ssh", "-A", node, "-T", command)
			cmd.Env = append(os.Environ(), `SUDO_OPTS="-S"`, `VERSION=${`+version+`}`)
			output, err := cmd.CombinedOutput()
			lock.Lock()
			fmt.Println(string(output))
			lock.Unlock()
			if err != nil {
				fmt.Printf("Error running command on %s: %v\n", node, err)
				return
			}
			fmt.Printf("Successfully updated %s\n", node)
		}(node)
	}
	wg.Wait()
}

func getVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/mat285/node-manager/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var release struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", err
	}
	return release.Name, nil
}
