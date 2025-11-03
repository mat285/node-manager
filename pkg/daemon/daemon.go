package daemon

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mat285/node-manager/pkg/kubectl"
	"github.com/mat285/node-manager/pkg/log"
	"github.com/mat285/node-manager/pkg/wait"
)

type Config struct {
	SyncIntervalSeconds int    `yaml:"syncIntervalSeconds"`
	Node                string `yaml:"node"`
}

type Daemon struct {
	Lock   *sync.Mutex
	Config Config

	cancel context.CancelFunc
	done   chan struct{}
}

func NewDaemon(config Config) *Daemon {
	if config.SyncIntervalSeconds <= 0 {
		config.SyncIntervalSeconds = 30
	}
	return &Daemon{
		Lock:   new(sync.Mutex),
		Config: config,
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	ctx, d.cancel = context.WithCancel(ctx)
	defer d.cancel()
	d.done = make(chan struct{})
	defer close(d.done)

	wg := wait.NewGroup()
	run := func(ctx context.Context, f func(ctx context.Context)) {
		wg.Add(1)
		go func() {
			defer d.cancel()
			defer wg.Done()
			f(ctx)
		}()
	}

	run(ctx, d.sync)

	err := wg.WaitContext(ctx)
	if err != nil {
		logger.Infof("error waiting for exit %v", err)
	}
	logger.Infof("Node manager daemon stopped")
	return err
}

func (d *Daemon) sync(ctx context.Context) {
	logger := log.GetLogger(ctx)
	logger.Infof("starting service watcher")
	d.updateNodeCPU(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Infof("stopping service watcher")
			return
		case <-time.After(time.Duration(d.Config.SyncIntervalSeconds) * time.Second):
		}
		d.updateNodeCPU(ctx)
	}
}
func (d *Daemon) updateNodeCPU(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	logger.Infof("fetching nodes")
	node, err := kubectl.GetNode(ctx, d.Config.Node)
	if err != nil {
		logger.Infof("Error getting node %v", err)
		return err
	}
	logger.Infof("node %v", node)

	if _, ok := node.Metadata.Labels[NodeLabelOverclock]; !ok {
		logger.Infof("overclock not enabled, skipping")
		return nil
	}

	overclock, err := strconv.ParseBool(strings.TrimSpace(node.Metadata.Labels[NodeLabelOverclock]))
	if err != nil || !overclock {
		logger.Infof("overclock not enabled, skipping")
		return SetCPUGovernor(ctx, d.Config.Node, "powersave")
	}

	pods, err := kubectl.GetPodsForNode(ctx, d.Config.Node)
	if err != nil {
		logger.Infof("Error getting pods: %v", err)
		return err
	}

	govenor := "powersave"
	logger.Infof("pods: %v", pods)
	for _, pod := range pods {
		if _, ok := pod.Metadata.Labels[PodLabelOverclock]; !ok {
			continue
		}
		overclock, err := strconv.ParseBool(strings.TrimSpace(pod.Metadata.Labels[PodLabelOverclock]))
		if err != nil {
			logger.Infof("Error parsing overclock label: %v", err)
			continue
		}
		if overclock {
			govenor = "performance"
			break
		}
	}
	logger.Infof("setting CPU governor to %v", govenor)
	return SetCPUGovernor(ctx, d.Config.Node, govenor)
}
