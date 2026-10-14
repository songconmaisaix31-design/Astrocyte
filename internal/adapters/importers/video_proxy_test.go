package importers

import (
	"context"
	"strings"
	"testing"
)

func TestVideoExplicitAnonymousProxyIsolation(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:2")
	t.Setenv("NODE_USE_ENV_PROXY", "1")
	t.Setenv("NO_PROXY", "*")
	e := installedExtractor(t)
	for _, entry := range e.isolatedEnv(t.TempDir(), true) {
		for _, key := range []string{"HTTP_PROXY=", "HTTPS_PROXY=", "NODE_USE_ENV_PROXY=", "NO_PROXY="} {
			if strings.HasPrefix(entry, key) {
				t.Fatal("inherited unapproved proxy configuration", key)
			}
		}
	}
	for _, proxy := range []string{"http://127.0.0.1:7890", "https://proxy.example.org:443/"} {
		configured, err := NewSummarizeExtractorWithOptions(e.NodeExecutable, e.CLIPath, SummarizeOptions{ProxyURL: proxy})
		if err != nil {
			t.Fatal("rejected explicit anonymous proxy", err)
		}
		for _, media := range []bool{false, true} {
			env := configured.isolatedEnv(t.TempDir(), media)
			for _, expected := range []string{"HTTP_PROXY=" + proxy, "HTTPS_PROXY=" + proxy, "NODE_USE_ENV_PROXY=1", "NO_PROXY=localhost,127.0.0.1,::1"} {
				found := false
				for _, entry := range env {
					found = found || entry == expected
				}
				if !found {
					t.Fatal("missing explicit proxy setting", expected)
				}
			}
			if err := configured.verifyVersion(context.Background(), env); err != nil {
				t.Fatal("actual pinned CLI rejected isolated proxy environment", err)
			}
		}
	}
	for _, proxy := range []string{" http://127.0.0.1:7890", "http://", "socks5://127.0.0.1:7890", "http://user:secret@127.0.0.1:7890", "http://127.0.0.1:7890/path", "http://127.0.0.1:7890?token=secret", "http://127.0.0.1:7890#fragment"} {
		if _, err := NewSummarizeExtractorWithOptions(e.NodeExecutable, e.CLIPath, SummarizeOptions{ProxyURL: proxy}); err == nil {
			t.Fatal("accepted invalid or credentialed proxy")
		}
	}
}
