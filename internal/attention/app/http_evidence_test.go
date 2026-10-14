package app_test

import (
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
	response, err := e.base.RoundTrip(request)
	if err != nil {
		return response, err
	}
	file, err := os.OpenFile(filepath.Join(e.root, fmt.Sprintf("response-%d.bin", e.next.Add(1))), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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
	root, err := filepath.Abs("../../../.verification")
	if err != nil {
		t.Fatal("resolve raw HTTP evidence root")
	}
	dir, err := os.MkdirTemp(root, "public-http-")
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
