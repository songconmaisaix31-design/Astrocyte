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
	f := swarmNativeExecutionFact(e)
	if f.OperationID != e.InputOperationID || f.SessionID != e.SessionID || !f.StopConfirmed || f.Outcome != "" || f.FailureCode != "" || f.Usage != nil || f.CallsKnown || f.TokensKnown {
		t.Fatal("controller exit or response observations changed unknown input outcome or budget usage")
	}
	if _, err := NewSwarmNativeBridge(nil, nil); err == nil {
		t.Fatal("native bridge without its actual dependencies was accepted")
	}
}
