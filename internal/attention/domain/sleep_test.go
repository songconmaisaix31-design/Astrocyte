package domain

import (
	"math"
	"testing"
)

func TestClusterTopicsGroupsByNeighborhoodAndAdjacency(t *testing.T) {
	events := []EventTopic{
		{EventID: "e1", Topic: "agent-swarm"},
		{EventID: "e2", Topic: "agent-swarm"},
		{EventID: "e3", Topic: "paper-reading"},
		{EventID: "e4", Topic: "video-summarize"},
	}
	related := map[string][]string{
		"paper-reading": {"agent-swarm"}, // adjacent → same cluster
		"video-summarize": nil,
	}

	clusters := ClusterTopics(events, related)
	if len(clusters) != 2 {
		t.Fatalf("got %d clusters: %+v", len(clusters), clusters)
	}
	var merged, isolated *Cluster
	for i := range clusters {
		if len(clusters[i].Topics) > 1 {
			merged = &clusters[i]
		} else {
			isolated = &clusters[i]
		}
	}
	if merged == nil || merged.Events != 3 {
		t.Fatalf("merged cluster wrong: %+v", clusters)
	}
	if isolated == nil || isolated.Events != 1 || isolated.Topics[0] != "video-summarize" {
		t.Fatalf("isolated cluster wrong: %+v", clusters)
	}
}

func TestClusterTopicsSkipsBlankTopic(t *testing.T) {
	events := []EventTopic{{EventID: "e1", Topic: "  "}, {EventID: "e2", Topic: "reading"}}
	clusters := ClusterTopics(events, nil)
	if len(clusters) != 1 || clusters[0].Topics[0] != "reading" || clusters[0].Events != 1 {
		t.Fatalf("blank topic leaked: %+v", clusters)
	}
}

func TestDenoiseDropsSingleIsolatedSignal(t *testing.T) {
	clusters := []Cluster{
		{Topics: []string{"noise"}, Events: 1},
		{Topics: []string{"connected"}, Events: 1},
		{Topics: []string{"solid"}, Events: 3},
	}
	related := map[string][]string{"connected": {"solid"}}
	kept := Denoise(clusters, related)
	if len(kept) != 2 {
		t.Fatalf("denoise kept %d: %+v", len(kept), kept)
	}
	for _, c := range kept {
		if c.Topics[0] == "noise" {
			t.Fatalf("isolated signal promoted: %+v", kept)
		}
	}
}

func TestResolveClustersDecidesFromShapeAlone(t *testing.T) {
	clusters := []Cluster{
		{Topics: []string{"agent-swarm"}, Events: 2},
		{Topics: []string{"a", "b"}, Events: 4},
	}
	decisions := ResolveClusters(clusters)
	if len(decisions) != 2 {
		t.Fatalf("got %d decisions", len(decisions))
	}
	if decisions[0].Kind != NodeCreate {
		t.Fatalf("single topic not create: %+v", decisions[0])
	}
	if decisions[1].Kind != NodeMerge {
		t.Fatalf("multi topic not merge: %+v", decisions[1])
	}
}

// The interpreter only explains; its output never enters a decision. Running
// ResolveClusters with no interpreter (LLM disconnected) leaves decisions
// unchanged, and a garbage interpreter only adds a label, never a decision.
func TestInterpreterOutputNeverChangesDecisions(t *testing.T) {
	cluster := Cluster{Topics: []string{"agent-swarm"}, Events: 2}
	decisions := ResolveClusters([]Cluster{cluster})
	if len(decisions) != 1 || decisions[0].Kind != NodeCreate {
		t.Fatalf("decision changed without interpreter: %+v", decisions)
	}
	exp, err := garbageInterpreter{}.Explain(cluster)
	if err != nil || exp.Label == "" {
		t.Fatal(err)
	}
	for _, d := range decisions {
		for _, topic := range d.Topics {
			if topic == exp.Label {
				t.Fatalf("interpreter label entered a decision: %+v", d)
			}
		}
	}
}

type garbageInterpreter struct{}

func (garbageInterpreter) Explain(Cluster) (ClusterExplanation, error) {
	return ClusterExplanation{Label: "untrusted label", Note: "must not decide"}, nil
}

func TestRetentionPrecisionMeasuresRecall(t *testing.T) {
	if got := RetentionPrecision(nil, nil); got != 1 {
		t.Fatalf("empty before = %f", got)
	}
	if got := RetentionPrecision([]string{"a", "b"}, []string{"a", "b"}); got != 1 {
		t.Fatalf("full survival = %f", got)
	}
	if got := RetentionPrecision([]string{"a", "b"}, []string{"a", "c"}); got != 0.5 {
		t.Fatalf("partial = %f", got)
	}
	if got := RetentionPrecision([]string{"a"}, []string{"b"}); got != 0 {
		t.Fatalf("none = %f", got)
	}
}

func TestStorageReductionClampsToUnit(t *testing.T) {
	if got := StorageReduction(0, 0); got != 0 {
		t.Fatalf("zero before = %f", got)
	}
	if got := StorageReduction(-1, -1); got != 0 {
		t.Fatalf("negative before = %f", got)
	}
	if got := StorageReduction(100, 200); got != 0 {
		t.Fatalf("growth = %f", got)
	}
	got := StorageReduction(100, 40)
	if math.Abs(got-0.6) > 1e-9 {
		t.Fatalf("reduction = %f", got)
	}
	if got := StorageReduction(100, -5); got != 1 {
		t.Fatalf("negative after clamp = %f", got)
	}
}

func TestComputeMetricsAssemblesBothIndicators(t *testing.T) {
	m := ComputeMetrics([]string{"a", "b"}, []string{"a"}, 100, 70)
	if m.BeforeFacts != 2 || m.AfterFacts != 1 {
		t.Fatalf("counts: %+v", m)
	}
	if m.RetentionPrecision != 0.5 {
		t.Fatalf("retention: %+v", m)
	}
	if math.Abs(m.StorageReduction-0.3) > 1e-9 {
		t.Fatalf("storage: %+v", m)
	}
}
