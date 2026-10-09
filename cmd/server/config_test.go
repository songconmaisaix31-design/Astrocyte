package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectSummarizeResolvesLocalPinWithoutGlobalCLI(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"astrocyte"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveProjectSummarize(); err == nil {
		t.Fatal("missing installation silently accepted")
	}
	packageRoot := filepath.Join(root, "node_modules", "@steipete", "summarize")
	if err := os.MkdirAll(packageRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"0.21.8", "0.25.1"} {
		if err := os.WriteFile(filepath.Join(packageRoot, "package.json"), []byte(`{"version":"`+version+`"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		path, err := resolveProjectSummarize()
		if version == "0.21.8" && err == nil {
			t.Fatal("stale global-era version accepted")
		}
		if version == "0.25.1" && (err != nil || path != filepath.Join(packageRoot, "dist", "cli.js")) {
			t.Fatalf("local CLI not resolved: %q %v", path, err)
		}
	}
}

func TestCodexOptInCannotImplicitlyAuthorizeSources(t *testing.T) {
	t.Setenv("ASTROCYTE_ENABLE_CODEX_DISTILLATION", "")
	t.Setenv("ASTROCYTE_PROCESSING_SOURCE_KEYS", "invalid")
	processor, sources, err := resolveDistiller(context.Background(), t.TempDir())
	if err != nil || processor != nil || len(sources) != 0 {
		t.Fatalf("default enabled processor: %v %v %v", processor, sources, err)
	}
	t.Setenv("ASTROCYTE_ENABLE_CODEX_DISTILLATION", "yes")
	if _, _, err := resolveDistiller(context.Background(), t.TempDir()); err == nil {
		t.Fatal("accepted ambiguous opt-in")
	}
	t.Setenv("ASTROCYTE_ENABLE_CODEX_DISTILLATION", "true")
	for _, raw := range []string{"", "null", "[]", `[""]`, `["arxiv:x","arxiv:x"]`, `[" arxiv:x"]`} {
		t.Setenv("ASTROCYTE_PROCESSING_SOURCE_KEYS", raw)
		if _, _, err := resolveDistiller(context.Background(), t.TempDir()); err == nil {
			t.Fatalf("accepted missing/ambiguous source scope: %s", raw)
		}
	}
	t.Setenv("ASTROCYTE_PROCESSING_SOURCE_KEYS", `["arxiv:2504.16054"]`)
	t.Setenv("ASTROCYTE_CODEX_EXECUTABLE", "codex.cmd")
	if _, _, err := resolveDistiller(context.Background(), t.TempDir()); err == nil {
		t.Fatal("accepted unverified native executable")
	}
}

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

func TestSummarizeExtractionCanBeDisabledAndRejectsInvalidPaths(t *testing.T) {
	t.Setenv("ASTROCYTE_ENABLE_SUMMARIZE", "false")
	t.Setenv("ASTROCYTE_SUMMARIZE_CLI", "")
	t.Setenv("ASTROCYTE_NODE", "missing")
	reader, err := resolveSourceReader(nil)
	if err != nil || reader.Arxiv.TextExtractor != nil {
		t.Fatalf("disabled must retain metadata/PDF-only reader: %v %v", reader, err)
	}
	t.Setenv("ASTROCYTE_ENABLE_SUMMARIZE", "yes")
	if _, err := resolveSourceReader(nil); err == nil {
		t.Fatal("accepted ambiguous summarize setting")
	}
	t.Setenv("ASTROCYTE_ENABLE_SUMMARIZE", "")
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
