//go:build !windows

package main

import (
	"context"
	"errors"
	"os/exec"
)

func configureCurrentProcess() error { return errors.New("此产品安装候选仅支持 Windows") }

func configureServiceChild(_ *exec.Cmd) {}

func resourceAdmission(_ context.Context) (any, error) {
	return nil, errors.New("此产品安装候选仅支持 Windows")
}
