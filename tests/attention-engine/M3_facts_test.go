package attention_engine_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestM3LiteralSleepFactsMap(t *testing.T) {
	var projector app.SleepFactProjector = app.LiteralSleepFactsMap
	tests := []struct {
		name string
		raw  string
		want []app.SleepFact
	}{
		{name: "nil"},
		{name: "null", raw: `null`},
		{name: "missing", raw: `{"original":"正文中存在事实也不提取"}`},
		{name: "null facts", raw: `{"facts":null}`},
		{name: "empty facts", raw: `{"facts":{}}`},
		{name: "unknown value", raw: `{"facts":{"unknown":null}}`},
		{
			name: "literal stable order",
			raw:  `{"facts":{"z":"值"," key ":" literal ","":"ignored"," \t":"ignored","a":"","unknown":null},"original":"不得提取"}`,
			want: []app.SleepFact{{Key: " key ", Value: " literal "}, {Key: "a", Value: ""}, {Key: "z", Value: "值"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := json.RawMessage(tt.raw)
			original := bytes.Clone(raw)
			for range 10 {
				got, err := projector(raw)
				if err != nil || !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("projection=%+v err=%v want=%+v", got, err, tt.want)
				}
				if !bytes.Equal(raw, original) {
					t.Fatal("projection mutated the original event bytes")
				}
			}
		})
	}
	for _, raw := range []string{`{`, `[]`, `{"facts":[]}`, `{"facts":{"n":1}}`, `{"facts":{"b":true}}`, `{"facts":{"object":{}}}`} {
		t.Run("invalid "+raw, func(t *testing.T) {
			got, err := projector(json.RawMessage(raw))
			if err == nil || got != nil {
				t.Fatalf("invalid format supplied facts: %+v err=%v", got, err)
			}
		})
	}
}
