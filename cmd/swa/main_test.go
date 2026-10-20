package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestOrcaCLIRejectsAuthorityArgumentsBeforeTransport(t *testing.T) {
	for _, option := range []string{"--from", "--actor", "--human", "--runtime-id", "--pid"} {
		t.Run(option, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"worker", "stop", "--controller", "unresolved-controller", "--dispatch", "unresolved-dispatch", option, "untrusted-input", "--json"}, &stdout, &stderr)
			if code != 2 || stdout.Len() != 0 {
				t.Fatalf("authority argument must be rejected before transport: exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
			}
		})
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
		if code := run(context.Background(), args, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Fatalf("invalid action or irrelevant input reached transport: exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
		}
	}
}

func TestOrcaCLIMissingCurrentGrantIsDenied(t *testing.T) {
	t.Setenv("SWA_PROJECT_TOKEN", "")
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"status", "--project", "unresolved-project", "--json"}, &stdout, &stderr)
	var result struct {
		OK    bool     `json:"ok"`
		Error *failure `json:"error"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if code != 3 || result.OK || result.Error == nil || result.Error.Code != "scope_denied" || stdout.Len() != 0 {
		t.Fatalf("missing grant must deny without native invocation: exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
}
