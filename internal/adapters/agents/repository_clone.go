package agents

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var managedRepositoryID = regexp.MustCompile(`^github-[1-9][0-9]*$`)
var gitObjectID = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

type ManagedRepositoryCloner struct{ base string }

// Construction and metadata sync do not create a checkout directory.
func NewManagedRepositoryCloner(base string) *ManagedRepositoryCloner {
	return &ManagedRepositoryCloner{base: base}
}

func (s *ManagedRepositoryCloner) CloneRepository(ctx context.Context, id, remote string) (domain.RepositoryClone, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if !managedRepositoryID.MatchString(id) || !filepath.IsAbs(s.base) {
		return domain.RepositoryClone{}, errors.New("invalid managed repository target")
	}
	owner, repo, err := ParsePublicGitHubInput(remote)
	if err != nil || repo == "" || remote != "https://github.com/"+owner+"/"+repo+".git" {
		return domain.RepositoryClone{}, errors.New("only canonical public GitHub clone URLs are allowed")
	}
	if err = os.MkdirAll(s.base, 0700); err != nil {
		return domain.RepositoryClone{}, err
	}
	base, err := CanonicalRoot(s.base)
	if err != nil {
		return domain.RepositoryClone{}, err
	}
	// Refuse a redirected application-owned root as well as existing targets.
	if rootKey(base) != rootKey(s.base) {
		return domain.RepositoryClone{}, errors.New("managed checkout root cannot be a symlink")
	}
	target := filepath.Join(base, id)
	if _, err = os.Lstat(target); err == nil {
		return s.inspect(ctx, target, remote)
	} else if !os.IsNotExist(err) {
		return domain.RepositoryClone{}, err
	}
	stage, err := os.MkdirTemp(base, ".clone-"+id+"-")
	if err != nil {
		return domain.RepositoryClone{}, err
	}
	// Failed/interrupted owned stages remain for diagnosis. Never delete or
	// overwrite an existing repository; a human retry uses a fresh bounded stage.
	checkout := filepath.Join(stage, "checkout")
	_, err = isolatedRepositoryGit(ctx, stage, []string{"clone", "--depth=1", "--single-branch", "--no-tags", "--no-recurse-submodules", "--template=", "--", remote, checkout})
	if err != nil {
		return domain.RepositoryClone{}, err
	}
	result, err := s.inspect(ctx, checkout, remote)
	if err != nil {
		return result, err
	}
	if _, err = os.Lstat(target); !os.IsNotExist(err) {
		return domain.RepositoryClone{}, errors.New("managed target already exists; refusing overwrite")
	}
	if err = publishManagedCheckout(checkout, target); err != nil {
		return domain.RepositoryClone{}, err
	}
	result.Root = target
	// Only remove the now-empty directory we just created. Never recursive delete.
	_ = os.Remove(stage)
	return result, nil
}

func (s *ManagedRepositoryCloner) inspect(ctx context.Context, root, remote string) (domain.RepositoryClone, error) {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return domain.RepositoryClone{}, errors.New("existing target is not a managed checkout directory")
	}
	gitDir := filepath.Join(root, ".git")
	for _, name := range []string{".git", ".git/config", ".git/HEAD", ".git/objects", ".git/refs"} {
		info, err = os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return domain.RepositoryClone{}, errors.New("managed Git metadata is unavailable or redirected")
		}
	}
	for _, name := range []string{"commondir", "gitdir", "objects/info/alternates"} {
		if _, e := os.Lstat(filepath.Join(gitDir, filepath.FromSlash(name))); !os.IsNotExist(e) {
			return domain.RepositoryClone{}, errors.New("managed Git metadata has an external storage redirection")
		}
	}
	for _, name := range []string{"packed-refs", "shallow"} {
		if info, e := os.Lstat(filepath.Join(gitDir, name)); e == nil && !info.Mode().IsRegular() {
			return domain.RepositoryClone{}, errors.New("managed Git metadata is redirected")
		} else if e != nil && !os.IsNotExist(e) {
			return domain.RepositoryClone{}, e
		}
	}
	count := 0
	err = filepath.WalkDir(filepath.Join(gitDir, "refs"), func(_ string, entry fs.DirEntry, e error) error {
		count++
		if e != nil {
			return e
		}
		if count > 128 || entry.Type()&os.ModeSymlink != 0 {
			return errors.New("managed refs are redirected or over bound")
		}
		return nil
	})
	if err != nil {
		return domain.RepositoryClone{}, err
	}
	// Reading only this small generated config prevents includes from expanding
	// a recovery probe into unrelated private config. No shell/helper is parsed.
	config, err := os.Open(filepath.Join(gitDir, "config"))
	if err != nil {
		return domain.RepositoryClone{}, err
	}
	defer config.Close()
	data, err := io.ReadAll(io.LimitReader(config, 64*1024+1))
	if err != nil || len(data) > 64*1024 {
		return domain.RepositoryClone{}, errors.New("managed Git config exceeds bound")
	}
	allowed := true
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(line, "[include") || strings.HasPrefix(line, "[credential") || strings.HasPrefix(line, "[filter") || strings.HasPrefix(line, "[http") || strings.HasPrefix(line, "[url ") || strings.HasPrefix(line, "worktree") || strings.HasPrefix(line, "[extensions") {
			allowed = false
		}
	}
	if !allowed {
		return domain.RepositoryClone{}, errors.New("managed Git configuration changed; manual review required")
	}
	data, err = isolatedRepositoryGit(ctx, root, []string{"--git-dir", gitDir, "--work-tree", root, "config", "--local", "--get", "remote.origin.url"})
	if err != nil || strings.TrimSpace(string(data)) != remote {
		return domain.RepositoryClone{}, errors.New("existing managed checkout has a different origin")
	}
	data, err = isolatedRepositoryGit(ctx, root, []string{"--git-dir", gitDir, "--work-tree", root, "rev-parse", "--verify", "HEAD"})
	head := strings.TrimSpace(string(data))
	if err != nil || !gitObjectID.MatchString(head) {
		return domain.RepositoryClone{}, errors.New("managed checkout has no valid HEAD")
	}
	return domain.RepositoryClone{Root: root, Head: head}, nil
}

func isolatedRepositoryGit(ctx context.Context, dir string, args []string) ([]byte, error) {
	entry, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	if ext := strings.ToLower(filepath.Ext(entry)); ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		return nil, errors.New("native Git executable required")
	}
	flags := []string{"--no-pager", "-c", "credential.helper=", "-c", "core.askPass=", "-c", "core.hooksPath=" + os.DevNull, "-c", "core.fsmonitor=false", "-c", "protocol.allow=never", "-c", "protocol.https.allow=always", "-c", "http.followRedirects=false", "-c", "filter.lfs.process=", "-c", "filter.lfs.smudge=", "-c", "filter.lfs.required=false", "-c", "submodule.recurse=false"}
	cmd := exec.Command(entry, append(flags, args...)...)
	cmd.Dir = dir
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		key = strings.ToUpper(key)
		if strings.HasPrefix(key, "GIT_") || strings.HasPrefix(key, "SSH_") || key == "HOME" || key == "USERPROFILE" || key == "XDG_CONFIG_HOME" {
			continue
		}
		cmd.Env = append(cmd.Env, env)
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "GIT_LFS_SKIP_SMUDGE=1", "HOME="+dir, "USERPROFILE="+dir, "XDG_CONFIG_HOME="+dir, "GIT_OPTIONAL_LOCKS=0")
	return runOwnedRepositoryGit(ctx, cmd)
}

func runOwnedRepositoryGit(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := &discoveryOutput{limit: 64 * 1024, overflow: make(chan struct{})}
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
		return nil, errors.New("Git output exceeded bound")
	}
	if out.exceeded {
		return nil, errors.New("Git output exceeded bound")
	}
	return append([]byte(nil), out.buffer.Bytes()...), err
}
