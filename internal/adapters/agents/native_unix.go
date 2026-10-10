//go:build !windows

package agents

import (
	"os/exec"
	"syscall"
)

func startOwnedNative(cmd *exec.Cmd) (func(), error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	return func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }, nil
}
