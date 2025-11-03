package daemon

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const (
	CPUGovernorPowersave   = "powersave"
	CPUGovernorPerformance = "performance"
	CPUGovernorUnknown     = "unknown"
)

func CPUGovenors() []string {
	return []string{CPUGovernorPowersave, CPUGovernorPerformance}
}

func SetCPUGovernor(ctx context.Context, governor string) error {
	cmd := exec.CommandContext(ctx, "cpupower", "frequency-set", "--governor", governor)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set CPU governor: %w: %s", err, string(output))
	}
	return nil
}

func GetCPUGovernor(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "cpupower", "frequency-info", "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get CPU governor: %w: %s", err, string(output))
	}
	for _, governor := range CPUGovenors() {
		if strings.Contains(string(output), governor) {
			return governor, nil
		}
	}
	return CPUGovernorUnknown, nil
}
