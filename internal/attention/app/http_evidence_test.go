package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// 原始响应仅写本机忽略目录，网络请求和响应仍来自真实 Transport。
type evidenceTransport struct {
	base http.RoundTripper
	root string
	next atomic.Int32
}

type evidenceBody struct {
	io.Reader
	body io.ReadCloser
	file *os.File
}

func (b *evidenceBody) Close() error {
	err := b.body.Close()
	fileErr := b.file.Close()
	if err != nil {
		return err
	}
	return fileErr
}

func (e *evidenceTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	sequence := e.next.Add(1)
	response, err := e.base.RoundTrip(request)
	metadata := struct {
		Method string `json:"method"`
		URL    string `json:"url"`
		Status int    `json:"status"`
		Error  string `json:"error,omitempty"`
	}{Method: request.Method, URL: request.URL.Redacted()}
	if response != nil {
		metadata.Status = response.StatusCode
	}
	if err != nil {
		metadata.Error = err.Error()
	}
	data, encodeErr := json.Marshal(metadata)
	if encodeErr != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		return nil, encodeErr
	}
	if writeErr := os.WriteFile(filepath.Join(e.root, fmt.Sprintf("response-%d.json", sequence)), data, 0600); writeErr != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		return nil, writeErr
	}
	if err != nil {
		return response, err
	}
	file, err := os.OpenFile(filepath.Join(e.root, fmt.Sprintf("response-%d.bin", sequence)), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		response.Body.Close()
		return nil, err
	}
	body := response.Body
	response.Body = &evidenceBody{Reader: io.TeeReader(body, file), body: body, file: file}
	return response, nil
}

func captureResponses(t *testing.T, client *http.Client) {
	t.Helper()
	dir, err := os.MkdirTemp(evidenceRoot(t), "public-http-")
	if err != nil {
		t.Fatal("create raw HTTP evidence directory")
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	client.Transport = &evidenceTransport{base: base, root: dir}
	t.Logf("raw actual public HTTP responses retained at=%s", dir)
}
