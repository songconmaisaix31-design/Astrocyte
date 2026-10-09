package main

import (
	"path/filepath"
	"testing"
)

func TestImportRootsExplicitAndNoDefault(t *testing.T) {
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", "")
	roots, err := resolveImportRoots()
	if err != nil || len(roots) != 0 {
		t.Fatalf("default roots: %v %v", roots, err)
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", "relative")
	if _, err := resolveImportRoots(); err == nil {
		t.Fatal("relative import root accepted")
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", filepath.Join(t.TempDir(), "missing"))
	if _, err := resolveImportRoots(); err == nil {
		t.Fatal("missing root accepted")
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", t.TempDir())
	if roots, err := resolveImportRoots(); err != nil || len(roots) != 1 {
		t.Fatalf("explicit root: %v %v", roots, err)
	}
}

func TestSummarizeExtractionRequiresExplicitInstalledPaths(t *testing.T) {
	t.Setenv("ASTROCYTE_SUMMARIZE_CLI", "")
	t.Setenv("ASTROCYTE_NODE", "missing")
	reader, err := resolveSourceReader(nil)
	if err != nil || reader.Arxiv.TextExtractor != nil {
		t.Fatalf("default must retain metadata/PDF-only reader: %v %v", reader, err)
	}
	t.Setenv("ASTROCYTE_SUMMARIZE_CLI", "relative.js")
	if _, err := resolveSourceReader(nil); err == nil {
		t.Fatal("accepted unverified extraction paths")
	}
	t.Setenv("ASTROCYTE_NODE", filepath.Join(t.TempDir(), "missing-node"))
	t.Setenv("ASTROCYTE_SUMMARIZE_CLI", filepath.Join(t.TempDir(), "missing-cli.js"))
	if _, err := resolveSourceReader(nil); err == nil {
		t.Fatal("accepted missing configured extraction tools")
	}
}

func TestAttentionPolicyRejectsInvalidWeights(t *testing.T) {
	t.Setenv("ASTROCYTE_ATTENTION_HALF_LIFE_SECONDS", "3600")
	for _, raw := range []string{"null", "[]", "invalid", `{"reread":-1}`, `{"agent_read":100}`} {
		t.Setenv("ASTROCYTE_ATTENTION_WEIGHTS", raw)
		if _, _, err := resolveAttentionPolicy(); err == nil {
			t.Fatalf("invalid attention weights accepted: %s", raw)
		}
	}
	t.Setenv("ASTROCYTE_ATTENTION_WEIGHTS", `{"reread":4}`)
	halfLife, weights, err := resolveAttentionPolicy()
	if err != nil || halfLife.Seconds() != 3600 || weights["reread"] != 4 {
		t.Fatalf("valid explicit attention parameters rejected: %v %v %v", halfLife, weights, err)
	}
}
func TestAllowedOriginsDoNotTrustOtherHosts(t *testing.T) {
	t.Setenv("ASTROCYTE_WEB_PORT", "15555")
	t.Setenv("ASTROCYTE_ALLOWED_ORIGINS", "")
	origins, err := resolveAllowedOrigins()
	if err != nil || origins[0] != "http://127.0.0.1:15555" {
		t.Fatalf("origin config %v %v", origins, err)
	}
	for _, origin := range []string{"http://example.com", "http://127.0.0.1.evil.test", "http://127.0.0.1:8080/path", "http://user@localhost", "file://localhost", "null"} {
		t.Setenv("ASTROCYTE_ALLOWED_ORIGINS", origin)
		if _, err := resolveAllowedOrigins(); err == nil {
			t.Fatalf("accepted %s", origin)
		}
	}
}
