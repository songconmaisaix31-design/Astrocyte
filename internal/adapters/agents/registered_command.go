package agents

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Only fixed read commands are callable. Neither a request nor repository
// contents can select executable names, arguments, hooks or a remote operation.
func registeredReadCommand(ctx context.Context, kind, root string) ([]byte, error) {
	tool, args, limit := "git", []string{}, 64*1024
	switch kind {
	case "orca_repos":
		tool, args, limit = registeredOrcaCommand(), []string{"repo", "list", "--json"}, 2*1024*1024
	case "orca_worktrees":
		tool, args, limit = registeredOrcaCommand(), []string{"worktree", "list", "--limit", "256", "--json"}, 2*1024*1024
	case "git_head", "git_branch", "git_time", "git_status", "git_common_dir":
		if !filepath.IsAbs(root) {
			return nil, errors.New("absolute registered root required")
		}
		args = []string{"--no-optional-locks", "--no-pager", "--work-tree", root, "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "core.hooksPath=", "-C", root}
		switch kind {
		case "git_common_dir":
			args = append(args, "rev-parse", "--path-format=absolute", "--git-common-dir")
		case "git_head":
			args = append(args, "rev-parse", "--verify", "HEAD")
		case "git_branch":
			args = append(args, "symbolic-ref", "--quiet", "--short", "HEAD")
		case "git_time":
			args = append(args, "log", "--no-show-signature", "-1", "--format=%cI")
		case "git_status":
			args = append(args, "status", "--porcelain=v1", "-z", "--untracked-files=no", "--ignore-submodules=all")
		}
	default:
		return nil, errors.New("read command not allowed")
	}
	entry, err := exec.LookPath(tool)
	if err != nil {
		return nil, err
	}
	// No shell wrapper or command interpolation. Orca's installed public binary
	// is used; an unverified wrapper stays explicitly unavailable.
	if ext := strings.ToLower(filepath.Ext(entry)); ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		return nil, errors.New("read-only binary entrypoint unavailable")
	}
	cmd := exec.Command(entry, args...)
	cmd.Env = []string{}
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		// Git command environment must not redirect the repository/index/config,
		// enable tracing sensitive config or invoke an inherited pager/helper.
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat")
	return runRegisteredRead(ctx, cmd, limit)
}

func registeredOrcaCommand() string {
	// Entrypoint environment only; no HTTP caller can supply a binary or args.
	// Linux's unrelated GNOME `orca` must never be launched as an IDE source.
	if configured := os.Getenv("ORCA_CLI_COMMAND"); configured != "" {
		return configured
	}
	if os.Getenv("ORCA_DEV_REPO_ROOT") != "" {
		return "orca-dev"
	}
	if runtime.GOOS == "linux" {
		return "orca-ide"
	}
	return "orca"
}

type discoveryOutput struct {
	buffer   bytes.Buffer
	limit    int
	overflow chan struct{}
	exceeded bool
}

func (b *discoveryOutput) Write(p []byte) (int, error) {
	n := len(p)
	if n > b.limit-b.buffer.Len() {
		if !b.exceeded {
			b.exceeded = true
			close(b.overflow)
		}
		return n, nil
	}
	_, _ = b.buffer.Write(p)
	return n, nil
}

func runRegisteredRead(ctx context.Context, cmd *exec.Cmd, limit int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := &discoveryOutput{limit: limit, overflow: make(chan struct{})}
	cmd.Stdout, cmd.Stderr = out, io.Discard
	cmd.WaitDelay = time.Second
	cleanup, err := startOwnedNative(cmd)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err = <-done:
	case <-ctx.Done():
		_ = cmd.Cancel()
		<-done
		return nil, ctx.Err()
	case <-out.overflow:
		_ = cmd.Cancel()
		<-done
		return nil, errors.New("read command output exceeds bound")
	}
	if out.exceeded {
		return nil, errors.New("read command output exceeds bound")
	}
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), out.buffer.Bytes()...), nil
}
