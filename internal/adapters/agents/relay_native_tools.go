package agents

import "github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"

// 当前 Codex 公开 app-server 无原生文件工具排除入口，受限执行立即拒绝。
func relayNativeToolsDisabled(id string) error {
	if id == "codex" {
		return nativeError(apierrors.UnsupportedCapability, "configured Codex app-server cannot disable native file tools for restricted execution")
	}
	return nil
}
