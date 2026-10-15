package events

import (
	"context"
	"os"
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
