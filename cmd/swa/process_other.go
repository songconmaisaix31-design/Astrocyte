//go:build !windows

package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
)

func configureCurrentProcess() error { return errors.New("此产品安装候选仅支持 Windows") }

func configureServiceChild(_ *exec.Cmd) {}

type serviceJob struct{}

func ownServiceTree() (*serviceJob, error) {
	return nil, errors.New("此产品安装候选仅支持 Windows")
}

func (*serviceJob) awaitDescendants(_ io.Writer) error {
	return errors.New("此产品安装候选仅支持 Windows")
}

func (*serviceJob) observeChild(_ *exec.Cmd, _ io.Writer) error {
	return errors.New("此产品安装候选仅支持 Windows")
}

func (*serviceJob) close() error { return nil }

func resourceAdmission(_ context.Context) (any, error) {
	return nil, errors.New("此产品安装候选仅支持 Windows")
}
