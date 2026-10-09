package domain

import (
	"errors"
	"testing"
	"time"
)

func TestMaterialRevisionHistoryAndReturningContent(t *testing.T) {
	history := []string{"original", "updated"}
	for _, tt := range []struct {
		digest   string
		revision int
		reused   bool
	}{{"updated", 2, true}, {"original", 1, true}, {"new", 3, false}} {
		r, reused, err := NextMaterialRevision(history, tt.digest)
		if err != nil || r != tt.revision || reused != tt.reused {
			t.Fatalf("%+v: %d %v %v", tt, r, reused, err)
		}
	}
	if history[0] != "original" {
		t.Fatal("history overwritten")
	}
}

func TestWorkIdentityFixedVersionsConfigAndNewQuestion(t *testing.T) {
	a := WorkIdentity{Inputs: []InputRef{{"material", "a", 1}, {"material", "b", 2}}, Stage: "theme", Configuration: "backend:model:v1"}
	b := a
	b.Inputs = []InputRef{{"material", "b", 2}, {"material", "a", 1}, {"material", "a", 1}}
	if !a.Same(b) {
		t.Fatal("same inputs should reuse despite order")
	}
	b.Inputs[1].Revision = 2
	if a.Same(b) {
		t.Fatal("changed revision reused")
	}
	b = a
	b.Configuration = "backend:model:v2"
	if a.Same(b) {
		t.Fatal("changed config reused")
	}
	b = a
	b.Question = "What evidence contradicts this?"
	if a.Same(b) {
		t.Fatal("new question reused")
	}
}

func TestThreeStagesPreserveUnansweredQuestions(t *testing.T) {
	refs := []InputRef{{"material", "a", 1}}
	if ValidateDistillation("content", refs, DistillationOutput{Summary: "text"}) == nil {
		t.Fatal("lost original source")
	}
	if err := ValidateDistillation("content", refs, DistillationOutput{SourcePreserved: true, Summary: "A real manual summary"}); err != nil {
		t.Fatal(err)
	}
	if ValidateDistillation("theme", refs, DistillationOutput{}) == nil {
		t.Fatal("empty theme")
	}
	if err := ValidateDistillation("theme", refs, DistillationOutput{OpenQuestions: []string{"Need a second source"}}); err != nil {
		t.Fatal(err)
	}
	project := DistillationOutput{Goal: "g", ExistingAsset: "a", ExpectedImprovement: "i", MinimumOutcome: "m", MissingEvidence: []string{"baseline"}}
	if err := ValidateDistillation("project", refs, project); err != nil {
		t.Fatal(err)
	}
	project.MissingEvidence = nil
	if ValidateDistillation("project", refs, project) == nil {
		t.Fatal("missing evidence assessment not recorded")
	}
}

func TestOpportunityAndFeedbackDoNotGrantExecution(t *testing.T) {
	state, _ := OpportunityState(nil, "use", "step")
	if state != "incubating" {
		t.Fatal(state)
	}
	state, _ = OpportunityState([]InputRef{{"material", "a", 1}}, "use", "step")
	if state != "ready_for_review" {
		t.Fatal(state)
	}
	later, negative, err := FeedbackTransition(state, "later", nil)
	if err != nil || later != "deferred" || negative {
		t.Fatalf("later %s %v %v", later, negative, err)
	}
	if _, _, err := FeedbackTransition(state, "reject", nil); !errors.Is(err, ErrEvidence) {
		t.Fatal("reject needs dimension reason")
	}
	rejected, negative, err := FeedbackTransition(state, "reject", map[string]string{"originality": "Known approach"})
	if err != nil || rejected != "rejected" || !negative {
		t.Fatal("reject lost negative reason")
	}
	adopted, _, _ := FeedbackTransition(state, "adopt", nil)
	if adopted != state {
		t.Fatal("adopt granted admission")
	}
}

func TestUnknownNotZeroAndCompleteProfileRequired(t *testing.T) {
	v := 0.5
	zero := 0.0
	d := map[string]DimensionScore{}
	w := map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}
	for _, name := range DimensionNames {
		d[name] = DimensionScore{Value: &v}
	}
	if CompositeScore(d, nil) != nil {
		t.Fatal("unconfigured profile ranked")
	}
	d["originality"] = DimensionScore{}
	if CompositeScore(d, w) != nil {
		t.Fatal("unknown filled with zero")
	}
	d["originality"] = DimensionScore{Value: &zero}
	if CompositeScore(d, w) == nil {
		t.Fatal("explicit zero treated as unknown")
	}
	w["originality"] = 0
	if CompositeScore(d, w) != nil {
		t.Fatal("dimension disabled")
	}
}

func TestHumanDecayExcludesReadsMachineAndRefresh(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	events := []AttentionEvent{{"human", "reread", now.Add(-time.Hour), 2}, {"agent", "reread", now, 100}, {"human", "read", now, 100}, {"human", "refresh", now, 100}}
	score, err := HumanActivity(events, now, time.Hour)
	if err != nil || score != 1 {
		t.Fatalf("got %f %v", score, err)
	}
	readAgain, _ := HumanActivity(events, now, time.Hour)
	if readAgain != score {
		t.Fatal("query mutated heat")
	}
}

func TestJobRecoveryCannotResendUnknownExternalAction(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	j := JobRules{State: "queued", MaxAttempts: 2, Deadline: now.Add(time.Hour)}
	j, err := j.Start(now)
	if err != nil || j.Attempts != 1 {
		t.Fatal(err)
	}
	j.ExternalStarted = true
	j = j.Recover()
	if j.State != "failed" || !j.DeliveryUnknown {
		t.Fatal("unknown external effect lost")
	}
	if _, err = j.Retry(now); !errors.Is(err, ErrUnknown) {
		t.Fatal("blind retry allowed")
	}
	local := JobRules{State: "failed", Attempts: 1, MaxAttempts: 2, Deadline: now.Add(time.Hour)}
	local, err = local.Retry(now)
	if err != nil || local.State != "queued" {
		t.Fatal(err)
	}
	local, _ = local.Start(now)
	local.State = "failed"
	if _, err = local.Retry(now); err == nil {
		t.Fatal("retry limit bypassed")
	}
	local.Attempts = 0
	if _, err = local.Retry(now.Add(time.Hour)); err == nil {
		t.Fatal("deadline bypassed")
	}
}
