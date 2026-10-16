// 自然到期后台验证程序。
// 仅在 ASTROCYTE_VERIFY_EXPIRY=1 时运行；未设置时直接失败，不新增 skip。
// 签发真实 1h token，记录派生身份和 UTC 时间（不记录 token 内容），
// 等待自然到期后验证 HTTP 和服务层拒绝。
package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// verificationDir 返回仓库根目录的 .verification 绝对路径。
func verificationDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", ".verification"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestProgressExpiryBackgroundVerification(t *testing.T) {
	if os.Getenv("ASTROCYTE_VERIFY_EXPIRY") != "1" {
		t.Fatal("ASTROCYTE_VERIFY_EXPIRY=1 required; do not run in normal suite")
	}

	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}

	if _, err := s.progress.SetProjectProgress(ctx, human, s.project.ID, domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}

	agentID := "expiry-bg-agent"
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	issuedAt := now.Format(time.RFC3339)
	expiresAt := token.ExpiresAt.Format(time.RFC3339)
	pid := os.Getpid()

	// 仅记录派生身份和 UTC 时间，绝不记录 token 内容或片段。
	logLine := fmt.Sprintf("PID=%d ISSUED=%s EXPIRES=%s PROJECT=%s AGENT=%s\n",
		pid, issuedAt, expiresAt, s.project.ID, agentID)

	vDir := verificationDir(t)
	logFile := filepath.Join(vDir, "expiry-bg-token.log")
	if err := os.WriteFile(logFile, []byte(logLine), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("expiry background: PID=%d ISSUED=%s EXPIRES=%s", pid, issuedAt, expiresAt)

	// 到期前验证
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusOK)

	// 等待到 ExpiresAt 后 +10s 安全边距
	waitDuration := time.Until(token.ExpiresAt) + 10*time.Second
	if waitDuration < 0 {
		t.Fatal("token already expired at issuance")
	}
	t.Logf("waiting %v until natural expiry", waitDuration)
	time.Sleep(waitDuration)

	// 到期后 HTTP 拒绝
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusForbidden)

	// 到期后服务层拒绝
	agent := domain.Caller{Kind: "agent", ID: agentID, ProjectID: s.project.ID}
	if _, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err == nil {
		t.Fatal("expired grant still accepted after natural expiry")
	}

	result := map[string]any{
		"pid": pid, "issued_at": issuedAt, "expires_at": expiresAt,
		"http_rejected": true, "service_rejected": true,
		"verified_at": time.Now().UTC().Format(time.RFC3339),
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	resultFile := filepath.Join(vDir, "expiry-bg-result.json")
	if err := os.WriteFile(resultFile, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("natural expiry verified: PID=%d verified_at=%s", pid, time.Now().UTC().Format(time.RFC3339))
}
