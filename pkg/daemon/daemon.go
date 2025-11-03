package daemon

import (
	"context"
	"errors"
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
	d.runCPUSync(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Infof("stopping service watcher")
			return
		case <-time.After(time.Duration(d.Config.SyncIntervalSeconds) * time.Second):
		}
		d.runCPUSync(ctx)
	}
}

func (d *Daemon) runCPUSync(ctx context.Context) error {
	err1 := d.updateNodeCPU(ctx)
	err2 := d.labelCurrentCPUGovernor(ctx)
	return errors.Join(err1, err2)
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

	if _, ok := node.Metadata.Labels[NodeLabelCPUOverclock]; !ok {
		logger.Infof("overclock not enabled, skipping")
		return nil
	}

	pods, err := kubectl.GetPodsForNode(ctx, d.Config.Node)
	if err != nil {
		logger.Infof("Error getting pods: %v", err)
		return err
	}

	govenor := CPUGovernorPowersave
	logger.Infof("pods: %v", pods)
	for _, pod := range pods {
		if _, ok := pod.Metadata.Labels[PodLabelCPUOverclock]; ok {
			govenor = CPUGovernorPerformance
			break
		}
	}
	logger.Infof("setting CPU governor to %v", govenor)
	err = SetCPUGovernor(ctx, govenor)
	if err != nil {
		logger.Infof("Error setting CPU governor: %v", err)
		return err
	}
	return nil
}

func (d *Daemon) labelCurrentCPUGovernor(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	governor, err := GetCPUGovernor(ctx)
	if err != nil {
		logger.Infof("Error getting CPU governor: %v", err)
		return err
	}
	logger.Infof("current CPU governor: %v", governor)
	err = kubectl.LabelNode(ctx, d.Config.Node, NodeLabelCPUGovernor, governor)
	if err != nil {
		logger.Infof("Error labeling node: %v", err)
		return err
	}
	return nil
}
