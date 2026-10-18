package app

import (
	"encoding/json"
	"testing"
)

func TestNativeMessageExplicitScopeAndAuthorizationJSON(t *testing.T) {
	// 只检查用途参数和私有类型的序列化，不生成原生来源。
	for _, scope := range []NativeMessageScope{{}, {Kind: "project", BindingID: "unexpected-binding"}, {Kind: "swarm"}, {Kind: "unknown"}} {
		if err := ValidateNativeMessageScope(scope); err == nil {
			t.Fatal("缺少明确用途或绑定时通过检查", scope)
		}
	}
	for _, scope := range []NativeMessageScope{{Kind: "project"}, {Kind: "swarm", BindingID: "explicit-binding-parameter"}} {
		if err := ValidateNativeMessageScope(scope); err != nil {
			t.Fatal(err)
		}
		for _, value := range []any{scope, NativeMessageLease{}, NativeMessagePreparation{Scope: scope}, NativeMessageAuthorization{Preparation: NativeMessagePreparation{Scope: scope}}} {
			encoded, err := json.Marshal(value)
			if err != nil || string(encoded) != "{}" {
				t.Fatal("私有发送授权进入公开 JSON", string(encoded), err)
			}
		}
	}
}
