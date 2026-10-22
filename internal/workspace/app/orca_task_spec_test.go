package app

import (
	"encoding/json"
	"testing"
)

func TestOrcaTaskSpecWorkerLimit(t *testing.T) {
	for _, limit := range []int64{1, 2, 3, 5, 9007199254740991} {
		spec := OrcaTaskSpec{Goal: "审阅公开接口", SourceRepo: "https://github.com/songconmaisaix31-design/swarm-anything", SourceRevision: "6195f4cfb6b32aedd43845443c029ecc11d5578d", SuccessCriteria: []string{"按真实来源报告"}, MaxWorkers: limit}
		if err := validateOrcaTaskSpec(spec); err != nil {
			t.Fatalf("max_workers=%d: %v", limit, err)
		}
		for _, invalid := range []int64{0, -1, 9007199254740992} {
			spec.MaxWorkers = invalid
			if err := validateOrcaTaskSpec(spec); err == nil {
				t.Fatalf("accepted max_workers=%d", invalid)
			}
		}
	}
	var spec OrcaTaskSpec
	if err := json.Unmarshal([]byte(`{"max_workers":3.5}`), &spec); err == nil {
		t.Fatal("fractional worker count must not decode as an integer")
	}
}
