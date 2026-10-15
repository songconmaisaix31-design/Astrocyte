package app

import "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"

// 原始固定文件正文仅留在控制器内存和真实输入；持久化绑定保留版本与身份。
func relaySessionMetadata(session domain.NativeSession) domain.NativeSession {
	session.ContextPacket.Files = append([]domain.ContextFile(nil), session.ContextPacket.Files...)
	for i := range session.ContextPacket.Files {
		session.ContextPacket.Files[i].Text = ""
	}
	session.ContextPacket.Materials = append([]domain.ContextMaterial(nil), session.ContextPacket.Materials...)
	for i := range session.ContextPacket.Materials {
		session.ContextPacket.Materials[i].Text = ""
	}
	return session
}
