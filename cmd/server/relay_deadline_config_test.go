package main

import "testing"

func TestRelayNativeDeadlineStartupConfiguration(t *testing.T) {
	for _, test := range []struct {
		value   string
		seconds int
		invalid bool
	}{{"", 0, false}, {"900", 900, false}, {"1", 1, false}, {"1800", 1800, false}, {"0", 0, true}, {"1801", 0, true}, {"invalid", 0, true}} {
		t.Run(test.value, func(t *testing.T) {
			t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", test.value)
			seconds, err := relayNativeDeadlineSeconds()
			if (err != nil) != test.invalid || seconds != test.seconds {
				t.Fatal("原生会话启动寿命配置未按边界立即失败")
			}
		})
	}
}
