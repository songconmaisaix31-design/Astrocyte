//go:build !windows

package judgmentadapter

import "os/exec"

func configureLayaProcess(_ *exec.Cmd) {}
