// 自然到期后台验证程序。
// 正常运行时通过 ASTROCYTE_VERIFY_EXPIRY=1 激活长等待。
// 签发真实 1h token，记录签发/到期时间，等待自然到期后验证 HTTP 和服务层拒绝。
package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestProgressExpiryBackgroundVerification(t *testing.T) {
	if os.Getenv("ASTROCYTE_VERIFY_EXPIRY") != "1" {
		t.Skip("set ASTROCYTE_VERIFY_EXPIRY=1 to run natural expiry background verification")
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

	logLine := fmt.Sprintf("PID=%d ISSUED=%s EXPIRES=%s PROJECT=%s AGENT=%s TOKEN_PREFIX=%s\n",
		pid, issuedAt, expiresAt, s.project.ID, agentID, token.Token[:20])

	logFile := ".verification/expiry-bg-token.log"
	if err := os.WriteFile(logFile, []byte(logLine), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("expiry background: %s", logLine)

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
	encoded, _ := json.MarshalIndent(result, "", "  ")
	if err := os.WriteFile(".verification/expiry-bg-result.json", encoded, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("natural expiry verified: %s", string(encoded))
}
