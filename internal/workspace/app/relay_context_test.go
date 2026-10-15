package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 仅验证私有正文的序列化边界，不声明会话或控制器已经持有。
func TestRelayStoredSessionExcludesFixedContextBody(t *testing.T) {
	session := domain.NativeSession{ContextPacket: domain.ContextPacket{ID: "context-boundary", ProjectID: "public-context", SettingsRevision: 3, Files: []domain.ContextFile{{Path: "README.md", Version: "fixed-version", Text: "PRIVATE_FIXED_FILE_CANARY"}}, Materials: []domain.ContextMaterial{{Text: "PRIVATE_FIXED_MATERIAL_CANARY"}}}}
	metadata := relaySessionMetadata(session)
	encoded, err := json.Marshal(metadata)
	if err != nil || strings.Contains(string(encoded), "PRIVATE_FIXED_") || metadata.ContextPacket.ID != session.ContextPacket.ID || metadata.ContextPacket.SettingsRevision != 3 || metadata.ContextPacket.Files[0].Version != "fixed-version" {
		t.Fatal("stored binding disclosed fixed context or discarded its version")
	}
	if session.ContextPacket.Files[0].Text != "PRIVATE_FIXED_FILE_CANARY" || session.ContextPacket.Materials[0].Text != "PRIVATE_FIXED_MATERIAL_CANARY" {
		t.Fatal("metadata projection changed the actual original input packet")
	}
}
