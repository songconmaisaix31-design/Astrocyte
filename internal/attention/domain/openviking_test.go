package domain

import "testing"

func TestOpenVikingURIParameterRules(t *testing.T) {
	root := "viking://user/alice/resources/astrocyte/projects/project"
	for _, raw := range []string{root + "/中文 笔记.md", root + "/%E4%B8%AD%E6%96%87%20%E7%AC%94%E8%AE%B0.md"} {
		canonical, err := CanonicalVikingURI(raw)
		if err != nil {
			t.Fatal(err)
		}
		if canonical != root+"/%E4%B8%AD%E6%96%87%20%E7%AC%94%E8%AE%B0.md" {
			t.Fatal("合法文件名没有使用相同规范表示", canonical)
		}
		if err := VikingWithin(raw, root); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{root + "/../other", root + "/%2e%2e/other", root + "/%2Foutside", root + "/%5coutside", root + "/%252e%252e", root + "/%00name", root + "/name?query=1", root + "/name?", root + "/name#", root + "/name#fragment", root + "//name", root + "-other/name"} {
		if err := VikingWithin(raw, root); err == nil {
			t.Fatal("越界或歧义参数未拒绝", raw)
		}
	}
}
