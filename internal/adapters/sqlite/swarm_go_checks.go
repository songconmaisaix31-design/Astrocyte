package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type SwarmCheckObjects interface {
	Read(context.Context, string) ([]byte, error)
	Publish(context.Context, []byte) (string, error)
}

type SwarmGoProjectChecks struct {
	root         string
	baseline     string
	objects      SwarmCheckObjects
	goExecutable string
}

type SwarmFixedCheckReceipt struct {
	Name       string    `json:"name"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Passed     bool      `json:"passed"`
	StdoutRef  string    `json:"stdout_ref"`
	StderrRef  string    `json:"stderr_ref"`
}

type SwarmProjectChecksReceipt struct {
	SchemaVersion int                      `json:"schema_version"`
	Baseline      string                   `json:"baseline"`
	Commit        string                   `json:"commit"`
	Checks        []SwarmFixedCheckReceipt `json:"checks"`
}

func NewSwarmGoProjectChecks(root, baseline string, objects SwarmCheckObjects) (*SwarmGoProjectChecks, error) {
	if objects == nil || baseline == "" {
		return nil, domain.Failure(apierrors.ValidationFailed, "explicit project baseline and immutable objects required")
	}
	canonical, err := swarmCanonical(root)
	if err != nil {
		return nil, err
	}
	if canonical != root {
		return nil, domain.Failure(apierrors.ScopeDenied, "project check root must be canonical")
	}
	resolved, err := swarmScalar(context.Background(), root, "rev-parse", "--verify", baseline+"^{commit}")
	if err != nil {
		return nil, err
	}
	if resolved != baseline {
		return nil, domain.Failure(apierrors.ContextStale, "project check baseline must be the full immutable commit")
	}
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}
	return &SwarmGoProjectChecks{root: root, baseline: baseline, objects: objects, goExecutable: goExecutable}, nil
}

func swarmCleanSnapshot(ctx context.Context, worktree, commit string) error {
	head, err := swarmScalar(ctx, worktree, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return err
	}
	if head != commit {
		return domain.Failure(apierrors.ContextStale, "project check worktree HEAD differs from submitted commit")
	}
	if _, err = swarmGit(ctx, worktree, "diff", "--quiet", "--no-ext-diff", "HEAD", "--"); err != nil {
		return domain.Failure(apierrors.ContextStale, "project check worktree has tracked changes")
	}
	untracked, err := swarmGit(ctx, worktree, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return err
	}
	if len(untracked) != 0 {
		return domain.Failure(apierrors.ContextStale, "project check worktree has untracked input")
	}
	return nil
}

func (c *SwarmGoProjectChecks) VerifyProjectChecks(ctx context.Context, l domain.Lease, a domain.Artifact) (SwarmProjectCheckResult, error) {
	var result SwarmProjectCheckResult
	if !time.Now().Before(l.ExpiresAt) {
		return result, domain.Failure(apierrors.LeaseLost, "project check lease expired")
	}
	ctx, cancel := context.WithDeadline(ctx, l.ExpiresAt)
	defer cancel()
	rootCommon, err := swarmScalar(ctx, c.root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return result, err
	}
	worktreeCommon, err := swarmScalar(ctx, l.Worktree, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return result, err
	}
	same, err := swarmSamePath(rootCommon, worktreeCommon)
	if err != nil {
		return result, err
	}
	if !same {
		return result, domain.Failure(apierrors.ScopeDenied, "fixed project checks cannot execute another repository")
	}
	if _, err = swarmGit(ctx, l.Worktree, "merge-base", "--is-ancestor", c.baseline, a.Commit); err != nil {
		return result, domain.Failure(apierrors.ContextStale, "submitted commit is outside the approved baseline")
	}
	if err = swarmCleanSnapshot(ctx, l.Worktree, a.Commit); err != nil {
		return result, err
	}
	receipt := SwarmProjectChecksReceipt{SchemaVersion: 1, Baseline: c.baseline, Commit: a.Commit, Checks: []SwarmFixedCheckReceipt{}}
	result.Passed = true
	for _, check := range []struct {
		name string
		args []string
	}{{"go_test", []string{"test", "./..."}}, {"go_build", []string{"build", "./..."}}} {
		command := exec.CommandContext(ctx, c.goExecutable, check.args...)
		command.Dir = l.Worktree
		command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0")
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		record := SwarmFixedCheckReceipt{Name: check.name, StartedAt: time.Now().UTC()}
		runErr := command.Run()
		record.FinishedAt = time.Now().UTC()
		record.Passed = runErr == nil
		if runErr != nil {
			var exit *exec.ExitError
			if !errors.As(runErr, &exit) {
				return result, runErr
			}
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			result.Passed = false
		}
		record.StdoutRef, err = c.objects.Publish(ctx, stdout.Bytes())
		if err != nil {
			return result, err
		}
		record.StderrRef, err = c.objects.Publish(ctx, stderr.Bytes())
		if err != nil {
			return result, err
		}
		receipt.Checks = append(receipt.Checks, record)
		if runErr != nil {
			break
		}
	}
	if err = swarmCleanSnapshot(ctx, l.Worktree, a.Commit); err != nil {
		return result, err
	}
	data, err := json.Marshal(receipt)
	if err != nil {
		return result, err
	}
	result.ReceiptRef, err = c.objects.Publish(ctx, data)
	return result, err
}
