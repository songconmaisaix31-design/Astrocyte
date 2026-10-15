package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayPublicContext struct {
	projects workspace.LocalProjects
	root     string
	identity os.FileInfo
	commit   string
	files    []string
}

func relayContextFailure(code apierrors.Code, reason string) error {
	return &apierrors.ServiceError{Code: code, Message: reason}
}

// 固定上下文只接受可信本机配置，文件正文由原项目授权入口读取。
func relayPublicLaunchContexts(projects workspace.LocalProjects) ([]workspace.RelayLaunchContextReader, error) {
	selected := os.Getenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT")
	commit := os.Getenv("ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT")
	fileJSON := os.Getenv("ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES")
	if selected == "" && commit == "" && fileJSON == "" {
		return nil, nil
	}
	decodedCommit, err := hex.DecodeString(commit)
	if projects == nil || selected == "" || fileJSON == "" || err != nil || len(decodedCommit) != 20 || strings.ToLower(commit) != commit {
		return nil, relayContextFailure(apierrors.ValidationFailed, "fixed public context requires its original project reader, root, commit and explicit files")
	}
	root, err := workbenchCanonical(selected)
	if err != nil || !filepath.IsAbs(selected) || !strings.EqualFold(filepath.Clean(selected), filepath.Clean(root)) {
		return nil, relayContextFailure(apierrors.ScopeDenied, "fixed public context root is redirected")
	}
	accessRoot, err := relayRoot()
	if err != nil || !relayInside(filepath.Join(accessRoot, "verification"), root) {
		return nil, relayContextFailure(apierrors.ScopeDenied, "fixed public context root is outside isolated verification state")
	}
	identity, err := os.Stat(root)
	if err != nil || !identity.IsDir() {
		return nil, relayContextFailure(apierrors.EvidenceMissing, "fixed public context root is unavailable")
	}
	decoder := json.NewDecoder(strings.NewReader(fileJSON))
	var files []string
	if decoder.Decode(&files) != nil || decoder.Decode(new(any)) != io.EOF || len(files) == 0 || len(files) > 32 {
		return nil, relayContextFailure(apierrors.ValidationFailed, "fixed public context files require one bounded JSON array")
	}
	seen := make(map[string]bool, len(files))
	for _, file := range files {
		if !fs.ValidPath(file) || file == "." || filepath.VolumeName(file) != "" || filepath.ToSlash(file) != file || seen[strings.ToLower(file)] {
			return nil, relayContextFailure(apierrors.ScopeDenied, "fixed public context requires unique relative file selections")
		}
		seen[strings.ToLower(file)] = true
	}
	return []workspace.RelayLaunchContextReader{&relayPublicContext{projects: projects, root: root, identity: identity, commit: commit, files: files}}, nil
}

func (r *relayPublicContext) git(ctx context.Context, arguments ...string) ([]byte, error) {
	args := []string{"--no-pager", "--no-replace-objects", "-c", "core.fsmonitor=false", "-c", "color.ui=false", "-C", r.root}
	command := exec.CommandContext(ctx, "git", append(args, arguments...)...)
	command.Stderr = io.Discard
	data, err := command.Output()
	if err != nil {
		return nil, relayContextFailure(apierrors.EvidenceMissing, "fixed public Git context is unavailable")
	}
	return data, nil
}

func (r *relayPublicContext) verify(ctx context.Context, project domain.LocalProject) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	canonical, err := workbenchCanonical(project.Root)
	if err != nil || !strings.EqualFold(filepath.Clean(canonical), filepath.Clean(r.root)) {
		return relayContextFailure(apierrors.ScopeDenied, "fixed public context differs from the confirmed project")
	}
	identity, err := os.Stat(r.root)
	if err != nil || !os.SameFile(r.identity, identity) {
		return relayContextFailure(apierrors.ScopeDenied, "fixed public context root identity changed")
	}
	head, err := r.git(ctx, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(head)) != r.commit {
		return relayContextFailure(apierrors.ContextStale, "fixed public context commit changed")
	}
	return nil
}

func (r *relayPublicContext) ReadRelayLaunchContext(ctx context.Context, caller domain.Caller, launch workspace.RelayLaunch) (domain.ContextPacket, error) {
	if err := r.verify(ctx, launch.Project); err != nil {
		return domain.ContextPacket{}, err
	}
	packet, err := r.projects.ReadProjectContext(ctx, caller, launch.Project.ID, domain.ContextRequest{Files: append([]string{}, r.files...)})
	if err != nil {
		return domain.ContextPacket{}, err
	}
	if packet.ID == "" || packet.ProjectID != launch.Project.ID || packet.SettingsRevision != launch.Project.Settings.Revision || packet.Mode != "selected_context" || len(packet.Materials) != 0 || len(packet.Files) != len(r.files) {
		return domain.ContextPacket{}, relayContextFailure(apierrors.ScopeDenied, "fixed public context differs from the confirmed revision and selection")
	}
	for i, file := range packet.Files {
		if file.Path != r.files[i] || file.Version == "" || !utf8.ValidString(file.Text) {
			return domain.ContextPacket{}, relayContextFailure(apierrors.EvidenceMissing, "fixed public context file version is unverified")
		}
		original, err := r.git(ctx, "show", r.commit+":"+file.Path)
		if err != nil {
			return domain.ContextPacket{}, err
		}
		if !bytes.Equal(original, []byte(file.Text)) {
			return domain.ContextPacket{}, relayContextFailure(apierrors.ContextStale, "fixed public context file differs from the approved commit")
		}
	}
	if err := r.verify(ctx, launch.Project); err != nil {
		return domain.ContextPacket{}, err
	}
	return packet, nil
}
