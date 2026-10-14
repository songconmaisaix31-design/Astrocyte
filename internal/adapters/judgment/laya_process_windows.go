package judgmentadapter

import (
	"os/exec"
	"syscall"
)

func configureLayaProcess(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
