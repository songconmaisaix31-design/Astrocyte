package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type SwarmObjectReader interface {
	Read(context.Context, string) ([]byte, error)
}

type SwarmGitEvidence struct {
	objects SwarmObjectReader
	checks  SwarmProjectChecker
}

type SwarmProjectChecker interface {
	VerifyProjectChecks(context.Context, domain.Lease, domain.Artifact) (SwarmProjectCheckResult, error)
}

type SwarmProjectCheckResult struct {
	Passed     bool
	ReceiptRef string
}

func (e *SwarmGitEvidence) ReadVerifiedDiff(ctx context.Context, a domain.Artifact) (string, error) {
	if a.ObjectRef == "" || a.Commit == "" || a.BaseCommit == "" || (a.State != "submitted" && a.State != "accepted" && a.State != "revision_needed") {
		return "", domain.Failure(apierrors.EvidenceMissing, "artifact has no committed verified diff")
	}
	data, err := e.objects.Read(ctx, a.ObjectRef)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(data) {
		return "", domain.Failure(apierrors.UnsupportedCapability, "verified diff is not UTF-8 text")
	}
	return string(data), nil
}

// 此记录对应固定 git diff --check 检查，其他测试使用独立证据引用。
type SwarmGitChecks struct {
	SchemaVersion      int    `json:"schema_version"`
	BaseCommit         string `json:"base_commit"`
	Commit             string `json:"commit"`
	DiffCheckPassed    bool   `json:"diff_check_passed"`
	DiffCheckOutputRef string `json:"diff_check_output_ref"`
}

func NewSwarmGitEvidence(objects SwarmObjectReader, checks ...SwarmProjectChecker) (*SwarmGitEvidence, error) {
	if objects == nil {
		return nil, domain.Failure(apierrors.ValidationFailed, "immutable object store required")
	}
	if len(checks) > 1 {
		return nil, domain.Failure(apierrors.ValidationFailed, "exactly one actual project check implementation is supported")
	}
	e := &SwarmGitEvidence{objects: objects}
	if len(checks) == 1 {
		if checks[0] == nil {
			return nil, domain.Failure(apierrors.ValidationFailed, "project check implementation is nil")
		}
		e.checks = checks[0]
	}
	return e, nil
}

func swarmGit(ctx context.Context, root string, args ...string) ([]byte, error) {
	argv := []string{"--no-pager", "-c", "core.fsmonitor=false", "-c", "diff.external=", "-C", root}
	argv = append(argv, args...)
	cmd := exec.CommandContext(ctx, "git", argv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, err
	}
	return out, nil
}

func swarmCanonical(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}
func swarmSamePath(a, b string) (bool, error) {
	ia, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	ib, err := os.Stat(b)
	if err != nil {
		return false, err
	}
	return os.SameFile(ia, ib), nil
}
func swarmScalar(ctx context.Context, root string, args ...string) (string, error) {
	b, err := swarmGit(ctx, root, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func (e *SwarmGitEvidence) VerifyWorktree(ctx context.Context, root, worktree string) (string, error) {
	r, err := swarmCanonical(root)
	if err != nil {
		return "", err
	}
	w, err := swarmCanonical(worktree)
	if err != nil {
		return "", err
	}
	if w != worktree {
		return "", domain.Failure(apierrors.ScopeDenied, "worktree must use its canonical absolute path")
	}
	same, err := swarmSamePath(r, w)
	if err != nil {
		return "", err
	}
	if same {
		return "", domain.Failure(apierrors.ScopeDenied, "execution requires an independent project worktree")
	}
	for _, p := range []string{r, w} {
		top, err := swarmScalar(ctx, p, "rev-parse", "--show-toplevel")
		if err != nil {
			return "", err
		}
		top, err = swarmCanonical(top)
		if err != nil {
			return "", err
		}
		same, err := swarmSamePath(top, p)
		if err != nil {
			return "", err
		}
		if !same {
			return "", domain.Failure(apierrors.ScopeDenied, "execution root must be the Git worktree root")
		}
	}
	rCommon, err := swarmScalar(ctx, r, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	wCommon, err := swarmScalar(ctx, w, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	same, err = swarmSamePath(rCommon, wCommon)
	if err != nil {
		return "", err
	}
	if !same {
		return "", domain.Failure(apierrors.ScopeDenied, "worktree does not belong to the approved project repository")
	}
	// 使用 Git 自身判定修改，不解析 status 或 diff 格式。
	if _, err = swarmGit(ctx, w, "diff", "--quiet", "--no-ext-diff", "HEAD", "--"); err != nil {
		return "", domain.Failure(apierrors.ContextStale, "worktree contains tracked changes before claim")
	}
	return swarmScalar(ctx, w, "rev-parse", "--verify", "HEAD^{commit}")
}

func swarmExitOne(err error) bool {
	var e *exec.ExitError
	return errors.As(err, &e) && e.ExitCode() == 1
}

func (e *SwarmGitEvidence) VerifyArtifact(ctx context.Context, l domain.Lease, a domain.Artifact) (swarmapp.ArtifactVerification, error) {
	var result swarmapp.ArtifactVerification
	if l.BaseCommit == "" || a.Commit == "" || strings.HasPrefix(a.Commit, "-") {
		return result, domain.Failure(apierrors.ValidationFailed, "fixed Git commits are required")
	}
	resolved, err := swarmScalar(ctx, l.Worktree, "rev-parse", "--verify", a.Commit+"^{commit}")
	if err != nil {
		return result, err
	}
	if resolved != a.Commit {
		return result, domain.Failure(apierrors.ContextStale, "artifact commit must be a full immutable commit id")
	}
	if _, err = swarmGit(ctx, l.Worktree, "merge-base", "--is-ancestor", l.BaseCommit, a.Commit); err != nil {
		return result, domain.Failure(apierrors.ContextStale, "artifact commit does not descend from the claimed snapshot")
	}
	if a.Commit == l.BaseCommit {
		return result, domain.Failure(apierrors.EvidenceMissing, "artifact has no new committed result")
	}
	paths := []string{"diff", "--quiet", "--no-ext-diff", l.BaseCommit, a.Commit, "--", "."}
	for _, p := range a.ChangedPaths {
		if err = domain.ValidateWritePath(p); err != nil {
			return result, err
		}
		paths = append(paths, ":(exclude,top,literal)"+p)
	}
	if _, err = swarmGit(ctx, l.Worktree, paths...); err != nil {
		if swarmExitOne(err) {
			return result, domain.Failure(apierrors.ScopeDenied, "commit includes paths omitted from submitted write scope")
		}
		return result, err
	}
	for _, p := range a.ChangedPaths {
		_, err = swarmGit(ctx, l.Worktree, "diff", "--quiet", "--no-ext-diff", l.BaseCommit, a.Commit, "--", ":(top,literal)"+p)
		if err == nil {
			return result, domain.Failure(apierrors.EvidenceMissing, "submitted changed path has no committed change")
		}
		if !swarmExitOne(err) {
			return result, err
		}
	}
	actualDiff, err := swarmGit(ctx, l.Worktree, "diff", "--no-ext-diff", "--no-color", "--binary", l.BaseCommit, a.Commit, "--")
	if err != nil {
		return result, err
	}
	publishedDiff, err := e.objects.Read(ctx, a.ObjectRef)
	if err != nil {
		return result, err
	}
	if !bytes.Equal(actualDiff, publishedDiff) {
		return result, domain.Failure(apierrors.ContextStale, "published diff differs from the immutable commit")
	}
	data, err := e.objects.Read(ctx, a.ChecksRef)
	if err != nil {
		return result, err
	}
	var checks SwarmGitChecks
	if err = json.Unmarshal(data, &checks); err != nil {
		return result, err
	}
	if checks.SchemaVersion != 1 || checks.BaseCommit != l.BaseCommit || checks.Commit != a.Commit || checks.DiffCheckOutputRef == "" {
		return result, domain.Failure(apierrors.ContextStale, "check evidence does not match the committed snapshot")
	}
	actualCheck, err := swarmGit(ctx, l.Worktree, "diff", "--check", "--no-ext-diff", l.BaseCommit, a.Commit, "--")
	if err != nil && !swarmExitOne(err) && !swarmExitTwo(err) {
		return result, err
	}
	result.ChecksPassed = err == nil
	publishedCheck, readErr := e.objects.Read(ctx, checks.DiffCheckOutputRef)
	if readErr != nil {
		return result, readErr
	}
	if checks.DiffCheckPassed != result.ChecksPassed || !bytes.Equal(actualCheck, publishedCheck) {
		return result, domain.Failure(apierrors.EvidenceMissing, "check record differs from the actual fixed Git check")
	}
	if e.checks != nil {
		checked, err := e.checks.VerifyProjectChecks(ctx, l, a)
		if err != nil {
			return result, err
		}
		if checked.ReceiptRef == "" {
			return result, domain.Failure(apierrors.EvidenceMissing, "actual project checks have no immutable receipt")
		}
		result.ChecksPassed = result.ChecksPassed && checked.Passed
		result.ApplicableChecksVerified = true
		result.ProjectChecksRef = checked.ReceiptRef
	}
	return result, nil
}

func swarmExitTwo(err error) bool {
	var e *exec.ExitError
	return errors.As(err, &e) && e.ExitCode() == 2
}
