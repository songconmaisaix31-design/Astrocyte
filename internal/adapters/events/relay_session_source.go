package events

import (
	"context"
	"io"
	"math"
	"os"

	"github.com/olivere/ndjson"
)

// 只核已由原控制器选定的文件；不读取正文、发现来源或产生采集游标。
// 返回的文件身份只供本机后续同来源复核，不能序列化为公共 DTO。
func VerifyRelaySessionSource(ctx context.Context, path, workspaceRoot string, approvedRoots []string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", sessionFailure("capture_cancelled")
	}
	if err := validateRelaySessionSource(path, workspaceRoot, approvedRoots); err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", sessionFailure("source_open_failed")
	}
	identity, verifyErr := verifyRelaySessionSourceFile(f, path)
	closeErr := f.Close()
	if verifyErr != nil {
		return "", verifyErr
	}
	if closeErr != nil {
		return "", sessionFailure("source_close_failed")
	}
	if ctx.Err() != nil {
		return "", sessionFailure("capture_cancelled")
	}
	return identity, nil
}

func verifyRelaySessionSourceFile(f *os.File, path string) (string, error) {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", sessionFailure("source_not_regular")
	}
	identity, err := relaySessionFileIdentity(f, info)
	if err != nil {
		return "", sessionFailure("source_identity_unverifiable_or_aliased")
	}
	named, err := os.Stat(path)
	if err != nil || !os.SameFile(info, named) {
		return "", sessionFailure("source_replaced")
	}
	after, err := relaySessionFileIdentity(f, info)
	if err != nil || after != identity {
		return "", sessionFailure("source_identity_unverifiable_or_aliased")
	}
	return identity, nil
}

// 使用同一文件句柄核对原档首条记录，不产生事件、正文输出或采集游标。
func VerifyRelaySessionArchive(ctx context.Context, path, workspaceRoot string, approvedRoots []string, nativeID, format string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", sessionFailure("capture_cancelled")
	}
	if nativeID == "" || format != "codex" && format != "claude_code" {
		return "", sessionFailure("session_header_unverifiable")
	}
	if err := validateRelaySessionSource(path, workspaceRoot, approvedRoots); err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", sessionFailure("source_open_failed")
	}
	identity, verifyErr := verifyRelaySessionSourceFile(f, path)
	if verifyErr == nil {
		verifyErr = verifyRelaySessionArchiveHeader(ctx, f, RelaySessionBinding{Path: path, WorkspaceRoot: workspaceRoot, NativeID: nativeID, Format: format})
	}
	if verifyErr == nil {
		current, err := verifyRelaySessionSourceFile(f, path)
		if err != nil {
			verifyErr = err
		} else if current != identity {
			verifyErr = sessionFailure("source_replaced")
		}
	}
	closeErr := f.Close()
	if verifyErr != nil {
		return "", verifyErr
	}
	if closeErr != nil {
		return "", sessionFailure("source_close_failed")
	}
	if ctx.Err() != nil {
		return "", sessionFailure("capture_cancelled")
	}
	return identity, nil
}

func verifyRelaySessionArchiveHeader(ctx context.Context, f *os.File, binding RelaySessionBinding) error {
	info, err := f.Stat()
	if err != nil {
		return sessionFailure("source_unavailable_or_redirected")
	}
	if info.Size() >= int64(math.MaxInt) {
		return sessionFailure("source_extent_unsupported")
	}
	reader := ndjson.NewReaderSize(sessionContextReader{ctx, io.NewSectionReader(f, 0, info.Size())}, int(info.Size())+1)
	if !reader.Next() {
		if reader.Err() != nil {
			return sessionFailure("header_read_failed")
		}
		return sessionFailure("session_header_unverifiable")
	}
	header := reader.Bytes()
	_, complete, err := sessionLineEnd(f, 0, int64(len(header)), info.Size())
	if err != nil {
		return sessionFailure("header_read_failed")
	}
	if !complete {
		return sessionFailure("session_header_unverifiable")
	}
	matches, err := sessionHeaderMatches(binding, header)
	if err != nil {
		return sessionFailure("malformed_json")
	}
	if !matches {
		return sessionFailure("session_header_unverifiable")
	}
	return nil
}
