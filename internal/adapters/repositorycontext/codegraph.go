package repositorycontext

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 路径由可信启动配置注入，HTTP不能选择程序、SDK或脚本。
type CodeGraph struct {
	Node   string
	SDK    string
	Script string
}

type graphRequest struct {
	Action string   `json:"action"`
	Root   string   `json:"root"`
	SDK    string   `json:"sdk"`
	Text   string   `json:"text"`
	Symbol string   `json:"symbol"`
	Limit  int      `json:"limit"`
	Depth  int      `json:"depth"`
	Roots  []string `json:"roots"`
}

type graphResponse struct {
	Error     string                      `json:"error"`
	Version   string                      `json:"version"`
	Indexed   bool                        `json:"indexed"`
	Symbols   []domain.RepositorySymbol   `json:"symbols"`
	Relations []domain.RepositoryRelation `json:"relations"`
}

type boundedOutput struct {
	bytes.Buffer
	max int64
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if int64(len(p)) > b.max-int64(b.Len()) {
		return 0, failure(apierrors.BudgetExhausted, "repository adapter output exceeds budget")
	}
	return b.Buffer.Write(p)
}

func (c CodeGraph) invoke(ctx context.Context, root, action string, q domain.RepositoryQuery) (graphResponse, error) {
	var result graphResponse
	if c.Node == "" || c.SDK == "" || c.Script == "" {
		return result, failure(apierrors.ProviderUnavailable, "CodeGraph is not configured")
	}
	payload, err := json.Marshal(graphRequest{action, root, c.SDK, q.Text, q.SymbolID, q.Budget.MaxResults, q.Budget.MaxDepth, q.Roots})
	if err != nil {
		return result, err
	}
	cmd := exec.CommandContext(ctx, c.Node, "--liftoff-only", c.Script)
	cmd.Env = append(os.Environ(), "CODEGRAPH_TELEMETRY=0", "CODEGRAPH_NO_DOWNLOAD=1", "CODEGRAPH_DIR=.codegraph")
	cmd.Stdin = bytes.NewReader(payload)
	output := &boundedOutput{max: q.Budget.MaxBytes}
	cmd.Stdout = output
	stderr := &boundedOutput{max: 1024 * 1024}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if err := saveGraphDiagnostic(root, append(stderr.Bytes(), []byte("\n"+err.Error())...)); err != nil {
			return result, err
		}
		return result, failure(apierrors.ProviderUnavailable, "CodeGraph operation failed; inspect the local index and runtime")
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		if err := saveGraphDiagnostic(root, output.Bytes()); err != nil {
			return result, err
		}
		return result, failure(apierrors.ProviderUnavailable, "CodeGraph response is invalid")
	}
	if result.Version != "1.0.1" {
		return result, failure(apierrors.IndexStale, "CodeGraph version differs")
	}
	if result.Error == "budget_exhausted" {
		return result, failure(apierrors.BudgetExhausted, "CodeGraph result count exceeds the approved budget")
	}
	if result.Error != "" {
		return result, failure(apierrors.ProviderUnavailable, "CodeGraph returned an unsupported error")
	}
	if action == "index" || action == "sync" || action == "rebuild" {
		if !result.Indexed {
			return result, failure(apierrors.EvidenceMissing, "CodeGraph did not complete indexing")
		}
	}
	return result, nil
}

func saveGraphDiagnostic(root string, data []byte) error {
	directory := filepath.Join(root, ".codegraph")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return failure(apierrors.InternalError, "repository diagnostic could not be preserved")
	}
	if err := os.WriteFile(filepath.Join(directory, "astrocyte-error-"+uuid.NewString()+".log"), data, 0600); err != nil {
		return failure(apierrors.InternalError, "repository diagnostic could not be preserved")
	}
	return nil
}

// copyFile只在本平台拥有的投影目录写入，源文件由os.Root约束。
func copyFile(root *os.Root, source, target string, remaining *int64) error {
	for current := source; current != "."; current = filepath.ToSlash(filepath.Dir(current)) {
		info, err := root.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return failure(apierrors.ScopeDenied, "projection source contains a symlink")
		}
	}
	info, err := root.Lstat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return failure(apierrors.ScopeDenied, "projection source must be a regular file")
	}
	if info.Size() > *remaining {
		return failure(apierrors.BudgetExhausted, "projection source exceeds byte budget")
	}
	input, err := root.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return err
	}
	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	n, err := io.Copy(output, io.LimitReader(input, *remaining+1))
	closeErr := output.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	*remaining -= n
	if *remaining < 0 {
		return failure(apierrors.BudgetExhausted, "projection source exceeds byte budget")
	}
	return nil
}
