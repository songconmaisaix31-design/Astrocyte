package repositorycontext

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type Git struct{ EnvironmentID string }

type fileVersion struct {
	Path   string
	Digest string
}

func failure(code apierrors.Code, message string) error {
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: "review_repository_context_scope_and_index"}
}

func isNotFound(err error) bool {
	var serviceErr *apierrors.ServiceError
	return errors.As(err, &serviceErr) && serviceErr.Code == apierrors.NotFound
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	argv := append([]string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", argv...)
	// 固定注册仓库的真实Git状态，不接受宿主变量重定向index或worktree。
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, failure(apierrors.ProviderUnavailable, "Git observation failed")
	}
	return output, nil
}

func canonicalRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", failure(apierrors.ScopeDenied, "repository root must be a directory")
	}
	return filepath.Clean(resolved), nil
}

func safeRelative(value string) bool {
	return value != "" && !strings.ContainsAny(value, "\\\x00\r\n") && !strings.HasPrefix(value, "/") && !strings.Contains(value, ":") && path.Clean(value) == value && value != ".." && !strings.HasPrefix(value, "../")
}

func under(file string, roots []string) bool {
	for _, root := range roots {
		if root == "." || file == root || strings.HasPrefix(file, root+"/") {
			return true
		}
	}
	return false
}

func privatePath(file string) bool {
	for _, segment := range strings.Split(strings.ToLower(file), "/") {
		if slices.Contains([]string{".git", ".aoci", ".codex", ".claude", ".ssh", "node_modules", "dist", "build", "target", "vendor", "credentials", "secrets"}, segment) || strings.HasPrefix(segment, ".codegraph") {
			return true
		}
		if segment == ".env" || strings.HasPrefix(segment, ".env.") || strings.HasSuffix(segment, ".pem") || strings.HasSuffix(segment, ".key") || segment == "id_rsa" || segment == "id_ed25519" {
			return true
		}
	}
	return false
}

func validateBudget(b domain.RepositoryBudget) error {
	if b.MaxFiles < 1 || b.MaxFiles > 20000 || b.MaxBytes < 1 || b.MaxBytes > 128*1024*1024 || b.MaxResults < 1 || b.MaxResults > 200 || b.MaxDepth < 1 || b.MaxDepth > 5 {
		return failure(apierrors.BudgetExhausted, "repository budget is absent or exceeds supported limits")
	}
	return nil
}

// Capture只读取授权范围的真实文件；os.Root限制路径逃逸，链接逐项拒绝。
func (g Git) Capture(ctx context.Context, p domain.LocalProject, roots []string, budget domain.RepositoryBudget) (domain.RepositoryVersion, error) {
	var v domain.RepositoryVersion
	if g.EnvironmentID == "" {
		return v, failure(apierrors.ValidationFailed, "environment identity is required")
	}
	if err := validateBudget(budget); err != nil {
		return v, err
	}
	if len(roots) == 0 {
		return v, failure(apierrors.ScopeDenied, "explicit repository roots are required")
	}
	for _, root := range roots {
		if !safeRelative(root) || privatePath(root) {
			return v, failure(apierrors.ScopeDenied, "repository scope contains an unsafe path")
		}
	}
	roots = slices.Clone(roots)
	slices.Sort(roots)
	roots = slices.Compact(roots)
	root, err := canonicalRoot(p.Root)
	if err != nil {
		return v, err
	}
	top, err := git(ctx, root, "rev-parse", "--show-toplevel")
	if err != nil {
		return v, err
	}
	topRoot, err := canonicalRoot(strings.TrimSpace(string(top)))
	if err != nil {
		return v, err
	}
	if !strings.EqualFold(topRoot, root) {
		return v, failure(apierrors.ScopeDenied, "registered root must equal the Git worktree root")
	}
	common, err := git(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return v, err
	}
	gitdir, err := git(ctx, root, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return v, err
	}
	head, err := git(ctx, root, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return v, err
	}
	indexArgs := append([]string{"ls-files", "--stage", "-z", "--"}, roots...)
	index, err := git(ctx, root, indexArgs...)
	if err != nil {
		return v, err
	}
	listArgs := append([]string{"ls-files", "-z", "--cached", "--others", "--exclude-standard", "--"}, roots...)
	listed, err := git(ctx, root, listArgs...)
	if err != nil {
		return v, err
	}
	files := bytes.Split(bytes.TrimSuffix(listed, []byte{0}), []byte{0})
	if len(files) > budget.MaxFiles {
		return v, failure(apierrors.BudgetExhausted, "repository file count exceeds budget")
	}
	handle, err := os.OpenRoot(root)
	if err != nil {
		return v, err
	}
	defer handle.Close()
	versions := []fileVersion{}
	var total int64
	seen := map[string]bool{}
	for _, item := range files {
		if err := ctx.Err(); err != nil {
			return v, err
		}
		file := string(item)
		if file == "" || seen[file] {
			continue
		}
		seen[file] = true
		if !safeRelative(file) || !under(file, roots) {
			return v, failure(apierrors.ScopeDenied, "Git returned an unsafe file path")
		}
		if privatePath(file) {
			continue
		}
		for current := file; current != "."; current = path.Dir(current) {
			info, err := handle.Lstat(current)
			if errors.Is(err, os.ErrNotExist) && current == file {
				break
			}
			if err != nil {
				return v, err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return v, failure(apierrors.ScopeDenied, "repository symlinks are outside the readable scope")
			}
		}
		info, err := handle.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			versions = append(versions, fileVersion{file, "deleted"})
			continue
		}
		if err != nil {
			return v, err
		}
		if !info.Mode().IsRegular() {
			return v, failure(apierrors.ScopeDenied, "repository entry is not a regular file")
		}
		if info.Size() > budget.MaxBytes-total {
			return v, failure(apierrors.BudgetExhausted, "repository bytes exceed budget")
		}
		f, err := handle.Open(file)
		if err != nil {
			return v, err
		}
		hash := sha256.New()
		n, readErr := io.Copy(hash, io.LimitReader(f, budget.MaxBytes-total+1))
		closeErr := f.Close()
		if readErr != nil {
			return v, readErr
		}
		if closeErr != nil {
			return v, closeErr
		}
		total += n
		if total > budget.MaxBytes {
			return v, failure(apierrors.BudgetExhausted, "repository bytes exceed budget")
		}
		versions = append(versions, fileVersion{file, hex.EncodeToString(hash.Sum(nil))})
	}
	slices.SortFunc(versions, func(a, b fileVersion) int { return strings.Compare(a.Path, b.Path) })
	manifest, err := json.Marshal(versions)
	if err != nil {
		return v, err
	}
	statusArgs := append([]string{"status", "--porcelain=v1", "-z", "--untracked-files=all", "--"}, roots...)
	status, err := git(ctx, root, statusArgs...)
	if err != nil {
		return v, err
	}
	v = domain.RepositoryVersion{EnvironmentID: g.EnvironmentID, ProjectID: p.ID, Root: root, RepoID: digest(bytes.TrimSpace(common)), WorktreeID: digest(append(bytes.TrimSpace(gitdir), []byte(root)...)), Head: strings.TrimSpace(string(head)), IndexDigest: digest(index), WorktreeDigest: digest(manifest), Roots: roots, SettingsRevision: p.Settings.Revision, Dirty: len(status) > 0, ObservedAt: time.Now().UTC()}
	return v, nil
}

func SameVersion(a, b domain.RepositoryVersion) bool {
	return a.EnvironmentID == b.EnvironmentID && a.ProjectID == b.ProjectID && a.RepoID == b.RepoID && a.WorktreeID == b.WorktreeID && a.Head == b.Head && a.IndexDigest == b.IndexDigest && a.WorktreeDigest == b.WorktreeDigest && a.SettingsRevision == b.SettingsRevision && slices.Equal(a.Roots, b.Roots)
}
