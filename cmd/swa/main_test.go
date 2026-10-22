package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func requireLocalRefusal(t *testing.T, code, wantedExit int, wantedCode, wantedRequest string, stdout *bytes.Buffer) {
	t.Helper()
	decoder := json.NewDecoder(stdout)
	var reply workspaceapp.OrcaEnvelope
	if err := decoder.Decode(&reply); err != nil {
		t.Fatal("CLI did not publish one structured refusal")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatal("CLI published more than one result")
	}
	if code != wantedExit || reply.OK || reply.Error == nil || reply.Error.Code != wantedCode || reply.Error.RequiredAction == "" || reply.Error.Retryable || reply.RequestID != wantedRequest || reply.RuntimeID != "" || reply.AuthorityDomain != "" || reply.RunID != "" {
		t.Fatal("The pretransport refusal lost its expected result or invented native identity")
	}
}

func TestOrcaCLIRejectsAuthorityArgumentsBeforeTransport(t *testing.T) {
	for _, option := range []string{"--from", "--actor", "--human", "--runtime-id", "--pid"} {
		t.Run(option, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"worker", "stop", "--controller", "unresolved-controller", "--dispatch", "unresolved-dispatch", option, "untrusted-input", "--json"}, &stdout, &stderr)
			requireLocalRefusal(t, code, 2, "validation_failed", "", &stdout)
		})
	}
	request := uuid.NewString()
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{"worker", "stop", "--request-id", request, "--from", "untrusted-input", "--json"}, request},
		{[]string{"worker", "stop", "--from", "untrusted-input", "--request-id", request, "--json"}, ""},
		{[]string{"worker", "stop", "--request-id", "invalid-request", "--from", "untrusted-input", "--json"}, ""},
	} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), test.args, &stdout, &stderr)
		requireLocalRefusal(t, code, 2, "validation_failed", test.want, &stdout)
	}
}

func TestOrcaCLIRejectsUnknownActionAndIrrelevantInput(t *testing.T) {
	for _, args := range [][]string{
		{"worker", "launch", "--json"},
		{"status", "--input-file", "unapproved.txt", "--json"},
		{"capabilities", "--limit", "100", "--json"},
		{"worker", "read", "--controller", "unresolved-controller", "--dispatch", "unresolved-dispatch", "--source", "terminal", "--json"},
	} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, &stdout, &stderr)
		requireLocalRefusal(t, code, 2, "validation_failed", "", &stdout)
	}
}

func TestOrcaCLIMissingCurrentGrantIsDenied(t *testing.T) {
	t.Setenv("SWA_PROJECT_TOKEN", "")
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"status", "--project", "unresolved-project", "--json"}, &stdout, &stderr)
	requireLocalRefusal(t, code, 3, "scope_denied", "", &stdout)
}

func TestOrcaCLIHelpDoesNotRequireGrantOrService(t *testing.T) {
	t.Setenv("SWA_PROJECT_TOKEN", "")
	t.Setenv("SWA_ENDPOINT", "invalid-endpoint")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, args := range [][]string{
		{"--help"}, {"controller", "--help"}, {"worker", "-h"},
		{"status", "--help"}, {"task", "create", "--help"},
		{"worker", "start", "--help"}, {"request", "show", "-h"},
		{"serve", "--help"}, {"serve", "--json", "--help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(ctx, args, &stdout, &stderr)
			if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "swa worker start") || !strings.Contains(stdout.String(), "--web-dir") || !strings.Contains(stdout.String(), "6=效果未知") || strings.Contains(stdout.String(), "invalid-endpoint") {
				t.Fatalf("帮助依赖了服务或未返回完整本地用法：exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestOrcaCLIHelpPreservesArgumentRefusals(t *testing.T) {
	for _, args := range [][]string{
		{"status", "--unknown-option", "--help", "--json"},
		{"worker", "start", "--unknown-option", "--help", "--json"},
		{"worker", "start", "--json"},
		{"serve", "--port", "invalid", "--help", "--json"},
		{"serve", "--unknown-option", "--help", "--json"},
		{"serve", "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, &stdout, &stderr)
			requireLocalRefusal(t, code, 2, "validation_failed", "", &stdout)
		})
	}
}
