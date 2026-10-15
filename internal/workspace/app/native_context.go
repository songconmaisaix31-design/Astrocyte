package app

import "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"

// SQLite 保存固定引用；原控制器保留正文，比较时只移除正文副本。
func nativeContextMetadata(packet domain.ContextPacket) domain.ContextPacket {
	packet.Materials = append([]domain.ContextMaterial{}, packet.Materials...)
	for i := range packet.Materials {
		packet.Materials[i].Text = ""
	}
	packet.Files = append([]domain.ContextFile{}, packet.Files...)
	for i := range packet.Files {
		packet.Files[i].Text = ""
	}
	return packet
}

// 原生恢复包含此前的 turns，因此保留原有的必要权限引用。
func mergeNativeContext(old, current domain.ContextPacket) domain.ContextPacket {
	refs := map[domain.FixedReference]bool{}
	for _, v := range current.Materials {
		refs[v.Reference] = true
	}
	for _, v := range old.Materials {
		if !refs[v.Reference] {
			v.Text = ""
			current.Materials = append(current.Materials, v)
			refs[v.Reference] = true
		}
	}
	files := map[string]bool{}
	for _, v := range current.Files {
		files[v.Path+"\x00"+v.Version] = true
	}
	for _, v := range old.Files {
		key := v.Path + "\x00" + v.Version
		if !files[key] {
			v.Text = ""
			current.Files = append(current.Files, v)
			files[key] = true
		}
	}
	return current
}
