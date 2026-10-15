package sqlite

import (
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 这些输入只验证匹配与未知值规则，不创建 Registry 或原生执行记录。
func TestSwarmNativeTupleParameterRules(t *testing.T) {
	l := domain.Lease{SessionID: "session", NativeSessionID: "native", AdapterID: "adapter", ControllerID: "controller", StartupOperationID: "startup", OperationID: "input", ProjectID: "project", Worktree: "canonical-worktree"}
	acceptedAt := time.Now().UTC()
	turn := "turn"
	e := workspace.RelayMemberEvidence{Purpose: "submit_artifact", SessionID: l.SessionID, NativeID: l.NativeSessionID, CLI: l.AdapterID, CLIVersion: "version", ControllerID: l.ControllerID, StartupOperationID: l.StartupOperationID, InputOperationID: l.OperationID, ProjectID: l.ProjectID, CWD: l.Worktree, ControllerHeld: true, ExitConfirmed: true, InputAccepted: true, AcceptedTurnID: &turn, InputAcceptedAt: &acceptedAt}
	if err := swarmMatchLeasedNative(e, l); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*workspace.RelayMemberEvidence)
	}{
		{"session", func(v *workspace.RelayMemberEvidence) { v.SessionID = "other" }},
		{"native", func(v *workspace.RelayMemberEvidence) { v.NativeID = "other" }},
		{"adapter", func(v *workspace.RelayMemberEvidence) { v.CLI = "other" }},
		{"controller", func(v *workspace.RelayMemberEvidence) { v.ControllerID = "other" }},
		{"startup", func(v *workspace.RelayMemberEvidence) { v.StartupOperationID = "other" }},
		{"input", func(v *workspace.RelayMemberEvidence) { v.InputOperationID = "other" }},
		{"project", func(v *workspace.RelayMemberEvidence) { v.ProjectID = "other" }},
		{"root", func(v *workspace.RelayMemberEvidence) { v.CWD = "other" }},
		{"purpose", func(v *workspace.RelayMemberEvidence) { v.Purpose = "claim_work_item" }},
		{"ownership", func(v *workspace.RelayMemberEvidence) { v.ControllerHeld = false }},
		{"version", func(v *workspace.RelayMemberEvidence) { v.CLIVersion = "" }},
		{"unaccepted", func(v *workspace.RelayMemberEvidence) { v.InputAccepted = false }},
		{"turn", func(v *workspace.RelayMemberEvidence) { v.AcceptedTurnID = nil }},
		{"accepted_at", func(v *workspace.RelayMemberEvidence) { v.InputAcceptedAt = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := e
			tc.change(&changed)
			if err := swarmMatchLeasedNative(changed, l); err == nil {
				t.Fatal("changed original tuple or missing evidence was accepted")
			}
		})
	}
	for _, change := range []func(*domain.Lease){
		func(v *domain.Lease) { v.NativeSessionID = "" },
		func(v *domain.Lease) { v.ControllerID = "" },
		func(v *domain.Lease) { v.StartupOperationID = "" },
		func(v *domain.Lease) { v.OperationID = v.StartupOperationID },
	} {
		changed := l
		change(&changed)
		if err := swarmMatchLeasedNative(e, changed); err == nil {
			t.Fatal("incomplete original lease or startup input was accepted")
		}
	}
	e.Purpose = "claim_work_item"
	if err := swarmMatchNativeTuple(e, e.Purpose, l.SessionID, l.NativeSessionID, l.AdapterID, l.ControllerID, l.StartupOperationID, l.OperationID, l.ProjectID, l.Worktree); err == nil {
		t.Fatal("exited controller was accepted for a new claim")
	}
	e.ExitConfirmed = false
	e.InputAccepted = false
	e.InputAcceptedAt = nil
	e.AcceptedTurnID = nil
	if err := swarmMatchNativeTuple(e, e.Purpose, l.SessionID, l.NativeSessionID, l.AdapterID, l.ControllerID, l.StartupOperationID, l.OperationID, l.ProjectID, l.Worktree); err != nil {
		t.Fatal("pre-input claim rule required a previously sent turn", err)
	}
}

func TestSwarmNativeMappingPreservesUnknownInputOutcomeAndBudgetUsage(t *testing.T) {
	responses, reported := int64(1), int64(19)
	e := workspace.RelayMemberEvidence{SessionID: "session", InputOperationID: "input", ExitConfirmed: true, InputAccepted: true, ToolsKnown: true, ObservedCompletedResponses: &responses, ReportedTotalTokens: &reported, Capabilities: []string{"observe"}, Tools: []string{"tool"}, ExternalTargets: []string{"target"}}
	d := swarmNativeDescriptor(e)
	if d.Known {
		t.Fatal("incomplete target evidence was promoted into a known execution policy")
	}
	e.TargetsKnown = true
	d = swarmNativeDescriptor(e)
	if !d.Known {
		t.Fatal("explicit policy dimension evidence was lost")
	}
	d.Capabilities[0], d.Tools[0], d.ExternalTargets[0] = "changed", "changed", "changed"
	if e.Capabilities[0] != "observe" || e.Tools[0] != "tool" || e.ExternalTargets[0] != "target" {
		t.Fatal("descriptor mutation changed the original source observation")
	}
	f, err := swarmNativeExecutionFact(e)
	if err != nil {
		t.Fatal(err)
	}
	if f.OperationID != e.InputOperationID || f.SessionID != e.SessionID || !f.StopConfirmed || f.Outcome != "" || f.FailureCode != "" || f.Usage != nil || f.CallsKnown || f.TokensKnown {
		t.Fatal("controller exit or response observations changed unknown input outcome or budget usage")
	}
	if _, err := NewSwarmNativeBridge(nil, nil); err == nil {
		t.Fatal("native bridge without its actual dependencies was accepted")
	}
}

func TestSwarmNativeKnownResultsPreserveIndependentStopAndUnknownUsage(t *testing.T) {
	observed := time.Now().UTC()
	turn := "accepted-turn-parameter"
	for _, outcome := range []string{"completed", "failed", "interrupted"} {
		t.Run(outcome, func(t *testing.T) {
			e := workspace.RelayMemberEvidence{Purpose: "submit_artifact", CLI: "codex", SessionID: "session-parameter", InputOperationID: "input-parameter", InputAccepted: true, AcceptedTurnID: &turn, InputAcceptedAt: &observed, Outcome: &outcome, ResultSource: "codex_turn_completed", ResultObservedAt: &observed}
			failure := "native_failure"
			if outcome == "failed" {
				e.FailureCode = &failure
			}
			f, err := swarmNativeExecutionFact(e)
			if err != nil {
				t.Fatal(err)
			}
			if f.Outcome != outcome || e.FailureCode != nil && f.FailureCode != *e.FailureCode {
				t.Fatal("original accepted input result was lost")
			}
			if f.StopConfirmed || f.Usage != nil || f.CallsKnown || f.TokensKnown {
				t.Fatal("observed input result changed independent stop or unknown budget facts")
			}
			e.ExitConfirmed = true
			f, err = swarmNativeExecutionFact(e)
			if err != nil {
				t.Fatal(err)
			}
			if !f.StopConfirmed || f.Outcome != outcome || f.Usage != nil || f.CallsKnown || f.TokensKnown {
				t.Fatal("original confirmed exit and input result were conflated")
			}
		})
	}
}

func TestSwarmNativeIncompleteOrContradictoryResultParametersFail(t *testing.T) {
	observed := time.Now().UTC()
	turn, outcome := "accepted-turn-parameter", "completed"
	e := workspace.RelayMemberEvidence{Purpose: "submit_artifact", CLI: "codex", SessionID: "session-parameter", InputOperationID: "input-parameter", InputAccepted: true, AcceptedTurnID: &turn, InputAcceptedAt: &observed, Outcome: &outcome, ResultSource: "codex_turn_completed", ResultObservedAt: &observed}
	for _, tc := range []struct {
		name   string
		change func(*workspace.RelayMemberEvidence)
	}{
		{"outcome", func(v *workspace.RelayMemberEvidence) { v.Outcome = nil }},
		{"source", func(v *workspace.RelayMemberEvidence) { v.ResultSource = "" }},
		{"other_source", func(v *workspace.RelayMemberEvidence) { v.ResultSource = "controller_exit" }},
		{"observed", func(v *workspace.RelayMemberEvidence) { v.ResultObservedAt = nil }},
		{"zero_observed", func(v *workspace.RelayMemberEvidence) { v.ResultObservedAt = &time.Time{} }},
		{"unsupported", func(v *workspace.RelayMemberEvidence) { value := "unknown"; v.Outcome = &value }},
		{"provider", func(v *workspace.RelayMemberEvidence) { v.CLI = "other" }},
		{"purpose", func(v *workspace.RelayMemberEvidence) { v.Purpose = "claim_work_item" }},
		{"unaccepted", func(v *workspace.RelayMemberEvidence) { v.InputAccepted = false }},
		{"turn", func(v *workspace.RelayMemberEvidence) { v.AcceptedTurnID = nil }},
		{"accepted_time", func(v *workspace.RelayMemberEvidence) { v.InputAcceptedAt = nil }},
		{"failure_code", func(v *workspace.RelayMemberEvidence) {
			value := "unclassified-result-parameter"
			v.FailureCode = &value
		}},
		{"empty_failure_code", func(v *workspace.RelayMemberEvidence) { value := ""; v.FailureCode = &value }},
		{"completed_failure", func(v *workspace.RelayMemberEvidence) { value := "native_failure"; v.FailureCode = &value }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := e
			tc.change(&changed)
			if _, err := swarmNativeExecutionFact(changed); err == nil {
				t.Fatal("incomplete or contradictory result observation was accepted")
			}
		})
	}
}
