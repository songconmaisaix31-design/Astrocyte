//go:build !windows

package agents

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

func runCLIProbe(ctx context.Context, path, flag string) (string, error) {
	if err := validateProbe(path, flag); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, path, flag)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	output := &boundedOutput{}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	return string(output.bytes), err
}
