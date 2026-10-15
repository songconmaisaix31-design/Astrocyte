package app

import "testing"

// 纯输入分类，不声明网络、PDF 或外部导入成功。
func TestImportFixedArxivBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, locator, adapter, kind, file string
		fixed                              bool
	}{
		{name: "implicit arxiv", locator: "https://arxiv.org/abs/2401.01234v1", kind: "paper", fixed: true},
		{name: "official pdf", locator: "https://arxiv.org/pdf/2401.01234v1.pdf", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "official html", locator: "https://www.arxiv.org/html/2401.01234v1?tracking=1#s2", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "official export", locator: "http://export.arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "bare fixed ID", locator: "2401.01234v12", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "prefixed legacy ID", locator: "arXiv:hep-th/9901001v2", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "legacy category ID", locator: "https://arxiv.org/abs/math.GT/0309136v1", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "unversioned arxiv", locator: "https://arxiv.org/abs/2401.01234", adapter: "arxiv", kind: "paper"},
		{name: "unversioned with query", locator: "https://arxiv.org/abs/2401.01234?version=v1", adapter: "arxiv", kind: "paper"},
		{name: "nonofficial versioned URL", locator: "https://example.test/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "host suffix", locator: "https://arxiv.org.example.test/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "userinfo", locator: "https://reader@arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "invalid scheme", locator: "ftp://arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "unrecognized path", locator: "https://arxiv.org/other/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "invalid ID", locator: "https://arxiv.org/abs/arbitraryv1", adapter: "arxiv", kind: "paper"},
		{name: "invalid version zero", locator: "https://arxiv.org/abs/2401.01234v0", adapter: "arxiv", kind: "paper"},
		{name: "invalid version leading zero", locator: "https://arxiv.org/abs/2401.01234v01", adapter: "arxiv", kind: "paper"},
		{name: "manual source", locator: "https://arxiv.org/abs/2401.01234v1", adapter: "manual", kind: "paper"},
		{name: "summarize local export", locator: "https://arxiv.org/abs/2401.01234v1", adapter: "summarize_json", kind: "paper", file: "mutable.json"},
		{name: "implicit local file", locator: "https://arxiv.org/abs/2401.01234v1", kind: "paper", file: "mutable.md"},
		{name: "implicit text", locator: "https://arxiv.org/abs/2401.01234v1", kind: "text"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := ImportMaterialCommand{SourceLocator: tc.locator, Adapter: tc.adapter, Kind: tc.kind, LocalFileRef: tc.file}
			if (fixedArxivVersion(c) != "") != tc.fixed {
				t.Fatalf("fixed input classification differs; expected %t", tc.fixed)
			}
		})
	}
}
