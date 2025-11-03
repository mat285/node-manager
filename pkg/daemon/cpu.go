package daemon

import (
	"context"
	"fmt"
	"os/exec"
)

func SetCPUGovernor(ctx context.Context, node string, governor string) error {
	cmd := exec.CommandContext(ctx, "sudo", "cpupower", "frequency-set", "--governor", governor)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set CPU governor: %w: %s", err, string(output))
	}
	return nil
}
