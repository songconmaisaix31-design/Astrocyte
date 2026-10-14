package judgmentadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type GitSnapshots struct {
	Git   string
	Files agents.ProjectFiles
}

func (g GitSnapshots) Read(ctx context.Context, project domain.LocalProject, files []string) (string, string, error) {
	if g.Git == "" {
		return "", "", fmt.Errorf("judgment Git snapshot requires an explicit Git executable")
	}
	root, err := g.Files.CanonicalRoot(project.Root)
	if err != nil {
		return "", "", err
	}
	// 使用 Git 原始完整输出作为版本输入，不解析 Git 文件格式。
	head, err := exec.CommandContext(ctx, g.Git, "-C", root, "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return "", "", fmt.Errorf("read actual judgment Git HEAD: %w", err)
	}
	if len(strings.TrimSpace(string(head))) == 0 {
		return "", "", fmt.Errorf("judgment Git HEAD is absent")
	}
	selected, err := g.Files.ReadProjectFiles(ctx, project, files)
	if err != nil {
		return "", "", err
	}
	versions := make([]struct {
		Path    string
		Version string
		Digest  string
	}, len(selected))
	for i, file := range selected {
		versions[i].Path, versions[i].Version = file.Path, file.Version
		versions[i].Digest = judgment.DigestText(file.Text)
	}
	data, err := json.Marshal(struct {
		Head  string
		Files any
	}{string(head), versions})
	if err != nil {
		return "", "", err
	}
	return judgment.DigestText(root), judgment.DigestText(string(data)), nil
}
