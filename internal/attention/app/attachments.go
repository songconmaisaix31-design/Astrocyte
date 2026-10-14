package app

// 私有原生处理元数据使用现有 MediaType 分类，保持持久化与 HTTP 契约字段。
const NativePrivateAttachmentMediaType = "application/vnd.astrocyte.private-native+json"

func privateNativeAttachment(revisions []MaterialRevision, target AttachmentRef) bool {
	for _, revision := range revisions {
		for _, attachment := range revision.Attachments {
			if attachment.ObjectRef == target.ObjectRef && (attachment.MediaType == NativePrivateAttachmentMediaType || attachment.Name == "whisper-original.json") {
				return true
			}
		}
	}
	return false
}
