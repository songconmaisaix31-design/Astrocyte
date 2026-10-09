package s1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type localArxivTransport struct{ target *url.URL }

func (tr localArxivTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	copy.URL.Scheme, copy.URL.Host = tr.target.Scheme, tr.target.Host
	return http.DefaultTransport.RoundTrip(copy)
}

// Contract-local official-source response data; never public-network acceptance.
func TestContractLocalArxivAliasesPinOriginalVersion(t *testing.T) {
	pdf := "%PDF-1.7\ncontract_local representative original bytes\n"
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/query":
			w.Header().Set("Content-Type", "application/atom+xml")
			_, _ = w.Write([]byte(`<feed><entry><id>https://arxiv.org/abs/2501.12948v1</id><title>contract_local paper</title><summary>contract_local abstract</summary></entry></feed>`))
		case "/pdf/2501.12948v1":
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte(pdf))
		default:
			http.NotFound(w, r)
		}
	}))
	defer fixture.Close()
	target, err := url.Parse(fixture.URL)
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	reader.Arxiv.Client = &http.Client{Transport: localArxivTransport{target: target}}
	var first app.ImportedSource
	for i, locator := range []string{"2501.12948v1", "https://arxiv.org/abs/2501.12948v1", "https://arxiv.org/pdf/2501.12948v1.pdf#page=1", "https://arxiv.org/html/2501.12948v1"} {
		source, err := reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", Kind: "paper", SourceLocator: locator})
		if err != nil {
			t.Fatalf("%s: %v", locator, err)
		}
		if source.SourceKey != "arxiv:2501.12948" || source.SourceLocator != "https://arxiv.org/abs/2501.12948v1" {
			t.Fatalf("alias changed identity/version: %+v", source)
		}
		if !strings.Contains(source.Text, "metadata only") || len(source.Attachments) != 2 || string(source.Attachments[0].Data) != pdf {
			t.Fatalf("original attachment/metadata distinction lost: %+v", source)
		}
		if i == 0 {
			first = source
		} else if first.Text != source.Text {
			t.Fatal("aliases changed fixed-version source content")
		}
	}
	_, err = reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v2"})
	if err == nil {
		t.Fatal("requested v2 accepted a v1 source response")
	}
}
