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

func TestNativeMessageExplicitApprovedActionClassification(t *testing.T) {
	// 仅分类动作参数，不构造会话绑定或已批准来源。
	for _, actions := range [][]string{nil, {}, {""}, {"unrecognized-action"}, {"claim_work_item", "unrecognized-action"}} {
		if _, err := NativeMessageRequiresSwarmLease(actions); err == nil {
			t.Fatal("无法确定用途的动作参数产生了 project 用途", actions)
		}
	}
	for _, test := range []struct {
		actions []string
		swarm   bool
	}{
		{[]string{"send"}, false},
		{[]string{"read_swarm"}, false},
		{[]string{"claim_work_item"}, true},
		{[]string{"submit_artifact"}, true},
		{[]string{"send", "claim_work_item"}, true},
	} {
		actual, err := NativeMessageRequiresSwarmLease(test.actions)
		if err != nil || actual != test.swarm {
			t.Fatal("动作参数分类与明确用途不同", test.actions, actual, err)
		}
	}
}
