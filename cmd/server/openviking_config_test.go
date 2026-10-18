package main

import "testing"

func TestOpenVikingConfigurationRules(t *testing.T) {
	names := []string{"OPENVIKING_BASE_URL", "OPENVIKING_API_KEY", "OPENVIKING_ACCOUNT", "OPENVIKING_USER", "OPENVIKING_ROOT_URI"}
	for _, name := range names {
		t.Setenv(name, "")
	}
	gateway, err := resolveOpenViking()
	if err != nil || gateway != nil {
		t.Fatal("未配置必须保持可选", err)
	}
	for _, input := range []struct{ name, value string }{{"OPENVIKING_API_KEY", "parameter"}, {"OPENVIKING_ROOT_URI", "viking://~/resources/astrocyte"}, {"OPENVIKING_BASE_URL", "ftp://127.0.0.1"}, {"OPENVIKING_BASE_URL", "https://user:password@127.0.0.1"}, {"OPENVIKING_BASE_URL", "http://127.0.0.1/?query"}} {
		for _, name := range names {
			t.Setenv(name, "")
		}
		t.Setenv(input.name, input.value)
		if _, err := resolveOpenViking(); err == nil {
			t.Fatal("无效配置未拒绝", input.name)
		}
	}
	for _, name := range names {
		t.Setenv(name, "")
	}
	t.Setenv("OPENVIKING_BASE_URL", "http://127.0.0.1:1933")
	t.Setenv("OPENVIKING_ACCOUNT", "parameter")
	if _, err := resolveOpenViking(); err == nil {
		t.Fatal("不完整身份未拒绝")
	}
	for _, name := range names {
		t.Setenv(name, "")
	}
	t.Setenv("OPENVIKING_BASE_URL", "http://127.0.0.1:1933/")
	if gateway, err := resolveOpenViking(); err != nil || gateway == nil {
		t.Fatal("合法尾斜线地址应按标准URL规范化", err)
	}
}
