package app

import "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"

// Native resume restores prior turns too; keep their minimal permission refs.
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
