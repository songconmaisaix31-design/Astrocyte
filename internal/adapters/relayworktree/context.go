package relayworktree

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type ProjectContextReader interface {
	ReadProjectContext(context.Context, domain.Caller, string, domain.ContextRequest) (domain.ContextPacket, error)
}

// 固定选择只由后台组装提供，根、文件和正文均不进入 Relay HTTP DTO。
type FixedContextSelection struct {
	ProjectID        string   `json:"-"`
	Root             string   `json:"-"`
	BaselineCommit   string   `json:"-"`
	SettingsRevision int      `json:"-"`
	Files            []string `json:"-"`
}

type FixedContext struct {
	reader   ProjectContextReader
	selected map[string]FixedContextSelection
}

var _ app.RelayLaunchContextReader = (*FixedContext)(nil)
var fullCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)

func NewFixedContext(reader ProjectContextReader, selections []FixedContextSelection) (*FixedContext, error) {
	if reader == nil || len(selections) == 0 {
		return nil, failure(apierrors.ValidationFailed, "authorized fixed context reader and selection required")
	}
	result := &FixedContext{reader: reader, selected: map[string]FixedContextSelection{}}
	for _, selection := range selections {
		if selection.ProjectID == "" || !filepath.IsAbs(selection.Root) || !fullCommit.MatchString(selection.BaselineCommit) || selection.SettingsRevision < 1 || len(selection.Files) == 0 || len(selection.Files) > 32 {
			return nil, failure(apierrors.ValidationFailed, "fixed context requires exact project, root, revision, commit and files")
		}
		if _, exists := result.selected[selection.ProjectID]; exists {
			return nil, failure(apierrors.ValidationFailed, "duplicate fixed context project selection")
		}
		selection.Files = append([]string(nil), selection.Files...)
		seen := map[string]bool{}
		for _, file := range selection.Files {
			if file == "" || filepath.IsAbs(file) || filepath.ToSlash(filepath.Clean(file)) != file || file == ".." || strings.HasPrefix(file, "../") || seen[file] {
				return nil, failure(apierrors.ValidationFailed, "fixed context files must be distinct relative paths")
			}
			seen[file] = true
		}
		result.selected[selection.ProjectID] = selection
	}
	return result, nil
}

func (s *FixedContext) verify(ctx context.Context, selection FixedContextSelection) error {
	root, err := git(ctx, selection.Root, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	if !sameExistingDirectory(strings.TrimSpace(root), selection.Root) {
		return failure(apierrors.ScopeDenied, "fixed context must select the actual repository root")
	}
	head, err := git(ctx, selection.Root, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(head) != selection.BaselineCommit {
		return failure(apierrors.VersionConflict, "fixed context baseline changed")
	}
	status, err := git(ctx, selection.Root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=all")
	if err != nil {
		return err
	}
	if status != "" {
		return failure(apierrors.VersionConflict, "fixed context source has uncommitted changes")
	}
	_, err = git(ctx, selection.Root, append([]string{"ls-files", "--error-unmatch", "--"}, selection.Files...)...)
	return err
}

func (s *FixedContext) ReadRelayLaunchContext(ctx context.Context, caller domain.Caller, launch app.RelayLaunch) (domain.ContextPacket, error) {
	selection, exists := s.selected[launch.Project.ID]
	if !exists || launch.Project.Settings.Revision != selection.SettingsRevision || !sameExistingDirectory(launch.Project.Root, selection.Root) || !sameExistingDirectory(launch.Workspace.SourceRoot, selection.Root) {
		return domain.ContextPacket{}, failure(apierrors.ScopeDenied, "launch differs from the approved fixed context selection")
	}
	// 调用方先核 Relay 的 read_context 授权；原端口再次核当前设置与授权。
	if err := s.verify(ctx, selection); err != nil {
		return domain.ContextPacket{}, err
	}
	packet, err := s.reader.ReadProjectContext(ctx, caller, selection.ProjectID, domain.ContextRequest{Files: append([]string(nil), selection.Files...)})
	if err != nil {
		return domain.ContextPacket{}, err
	}
	if packet.ID == "" || packet.ProjectID != selection.ProjectID || packet.SettingsRevision != selection.SettingsRevision || len(packet.Materials) != 0 || len(packet.Files) != len(selection.Files) {
		return domain.ContextPacket{}, failure(apierrors.ScopeDenied, "authorized context packet differs from its fixed selection")
	}
	for i, file := range packet.Files {
		if file.Path != selection.Files[i] || file.Version == "" {
			return domain.ContextPacket{}, failure(apierrors.ScopeDenied, "authorized context file differs from its fixed selection")
		}
	}
	if err = s.verify(ctx, selection); err != nil {
		return domain.ContextPacket{}, err
	}
	return packet, nil
}
