package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"reflect"
)

var _ RelayRuntimeFactsReader = (*relayService)(nil)

func (s *relayService) ListOwnedControllerHistory(ctx context.Context, id string) ([]domain.RelayControllerEvidence, error) {
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	t, exists := st.Tasks[id]
	if !exists {
		return nil, apierrors.NewNotFound("relay task", id)
	}
	p, err := s.permission(ctx, domain.Caller{}, t, "observe")
	if err != nil {
		return nil, err
	}
	if len(st.ControllerHistory[id]) != 0 || st.ControllerEvidence[id].ControllerID != "" {
		workspace, exists := st.Workspaces[t.WorktreeID]
		if !exists || workspace.ProjectID != t.ProjectID || !sameRelayRoot(workspace.SourceRoot, p.Root) || !s.existingOwnedRoot(workspace.Root) || !s.existingOwnedRoot(workspace.SourceRoot) {
			return nil, relayError(apierrors.ScopeDenied, "historical controller workspace is unavailable or changed")
		}
	}
	result := []domain.RelayControllerEvidence{}
	for _, evidence := range st.ControllerHistory[id] {
		evidence.Facts = append([]domain.RelayDriverFact{}, evidence.Facts...)
		result = append(result, evidence)
	}
	if evidence, exists := st.ControllerEvidence[id]; exists {
		evidence.Facts = append([]domain.RelayDriverFact{}, evidence.Facts...)
		result = append(result, evidence)
	}
	return result, nil
}

func (s *relayService) ListRuntimeFacts(ctx context.Context, caller domain.Caller, id string) ([]domain.RelayRuntimeFact, []domain.RelaySessionRelation, error) {
	st, err := s.snapshot(ctx, caller)
	if err != nil {
		return nil, nil, err
	}
	task, ok := st.Tasks[id]
	if !ok {
		return nil, nil, apierrors.NewNotFound("relay task", id)
	}
	if _, err = s.permission(ctx, caller, task, "observe"); err != nil {
		return nil, nil, err
	}
	facts := []domain.RelayRuntimeFact{}
	relations := []domain.RelaySessionRelation{}
	for _, fact := range st.RuntimeFacts {
		if fact.TaskID == id {
			facts = append(facts, fact)
		}
	}
	for _, relation := range st.SessionRelations {
		if relation.TaskID == id {
			relations = append(relations, relation)
		}
	}
	return facts, relations, nil
}

func (s *relayService) claimFact(st *RelayState, r domain.RelayReceipt) {
	t := st.Tasks[r.TaskID]
	st.RuntimeFacts = append(st.RuntimeFacts, domain.RelayRuntimeFact{ID: uuid.NewString(), TaskID: t.ID, TerminalID: t.TerminalID, WorktreeID: t.WorktreeID, ReceiptID: r.ID, OperationID: r.OperationID, Kind: "action_claimed", Status: "confirmed", Source: "relay_store", TS: s.now()})
}

func (s *relayService) observeRuntimeFacts(st *RelayState, task domain.RelayTask, obs RelayObservation) error {
	if obs.Evidence == nil {
		return nil
	}
	evidence := *obs.Evidence
	if evidence.StartupOperationID == "" {
		return nil
	}
	if evidence.ControllerID == "" || !sameRelayRoot(evidence.WorkspaceRoot, st.Workspaces[task.WorktreeID].Root) || evidence.Provider != task.Agent {
		return relayError(apierrors.ScopeDenied, "controller differs from the original workspace or provider")
	}
	if obs.Session.ID != "" && (evidence.SessionID != obs.Session.ID || evidence.NativeID != obs.Session.NativeID || obs.Session.ProjectID != task.ProjectID || obs.Session.CLI != task.Agent || obs.Session.Ownership != "owned") {
		return relayError(apierrors.ScopeDenied, "controller differs from the original native session")
	}
	startupBound := false
	for _, r := range st.Receipts {
		if r.TaskID == task.ID && r.OperationID == evidence.StartupOperationID && (r.Action == "dispatch" || r.Action == "resume") {
			startupBound = true
		}
	}
	if !startupBound {
		return relayError(apierrors.ScopeDenied, "controller startup operation is not bound to this task")
	}
	if obs.NativeCapture != nil && evidence.ControllerID != "" && (obs.NativeCapture.ControllerID != evidence.ControllerID || obs.NativeCapture.StartupOperationID != evidence.StartupOperationID) {
		return relayError(apierrors.ScopeDenied, "native capture belongs to another controller connection")
	}
	previous := st.ControllerEvidence[task.ID]
	if previous.ControllerID != "" && previous.ControllerID != evidence.ControllerID && (evidence.Relation != "resume" || evidence.PreviousControllerID != previous.ControllerID || evidence.SessionID != previous.SessionID || evidence.NativeID != previous.NativeID) {
		return relayError(apierrors.ScopeDenied, "resumed controller identity differs")
	}
	var last int64
	if previous.ControllerID == evidence.ControllerID {
		for _, f := range previous.Facts {
			if f.Sequence > last {
				last = f.Sequence
			}
		}
	}
	newFacts := []domain.RelayRuntimeFact{}
	for _, f := range evidence.Facts {
		if f.Sequence <= 0 || f.TS.IsZero() || f.TS.After(s.now()) {
			return relayError(apierrors.EvidenceMissing, "controller fact sequence or timestamp is invalid")
		}
		if f.Sequence <= last {
			matched := false
			for _, old := range previous.Facts {
				if old.Sequence == f.Sequence && reflect.DeepEqual(old, f) {
					matched = true
					break
				}
			}
			if !matched {
				return relayError(apierrors.ScopeDenied, "controller fact changed at an existing sequence")
			}
			continue
		}
		if f.Sequence != last+1 {
			return relayError(apierrors.EvidenceMissing, "controller fact sequence has a gap")
		}
		last = f.Sequence
		var receipt domain.RelayReceipt
		for _, r := range st.Receipts {
			if r.TaskID == task.ID && r.OperationID == f.OperationID && (r.Action == "dispatch" || r.Action == "resume" || r.Action == "send") {
				receipt = r
				break
			}
		}
		// 无准确动作关系的观察保留未知，不绑定到最近的其他动作。
		status, reason := f.Status, f.Reason
		if receipt.ID == "" {
			status, reason = "unknown", "operation_not_bound"
		}
		newFacts = append(newFacts, domain.RelayRuntimeFact{ID: uuid.NewString(), TaskID: task.ID, TerminalID: task.TerminalID, WorktreeID: task.WorktreeID, ReceiptID: receipt.ID, OperationID: f.OperationID, Kind: f.Kind, Status: status, Source: f.Source, Reason: reason, Sequence: f.Sequence, TS: f.TS})
	}
	st.RuntimeFacts = append(st.RuntimeFacts, newFacts...)
	confirmed, previouslyConfirmed := false, false
	for _, f := range evidence.Facts {
		if f.Kind == "session_confirmed" && f.Status == "confirmed" {
			confirmed = true
		}
	}
	for _, f := range previous.Facts {
		if f.Kind == "session_confirmed" && f.Status == "confirmed" {
			previouslyConfirmed = true
		}
	}
	if previous.ControllerID != evidence.ControllerID || confirmed && !previouslyConfirmed {
		kind, prior, status := evidence.Relation, "", "unknown"
		for _, f := range evidence.Facts {
			if f.Kind == "session_confirmed" && f.Status == "confirmed" {
				status = "confirmed"
			}
		}
		if kind != "new" && kind != "resume" {
			kind, status = "unknown", "unknown"
		}
		if kind == "new" && st.HandoffDeclarations[task.ID] != "" {
			kind, prior = "handoff", st.HandoffDeclarations[task.ID]
		}
		st.SessionRelations = append(st.SessionRelations, domain.RelaySessionRelation{TaskID: task.ID, PreviousTask: prior, WorktreeID: task.WorktreeID, TerminalID: task.TerminalID, OperationID: evidence.StartupOperationID, Kind: kind, Status: status, Source: "controller_session", TS: s.now()})
	}
	if st.ControllerEvidence == nil {
		st.ControllerEvidence = map[string]domain.RelayControllerEvidence{}
	}
	if previous.ControllerID != "" && previous.ControllerID != evidence.ControllerID {
		if st.ControllerHistory == nil {
			st.ControllerHistory = map[string][]domain.RelayControllerEvidence{}
		}
		st.ControllerHistory[task.ID] = append(st.ControllerHistory[task.ID], previous)
	}
	st.ControllerEvidence[task.ID] = evidence
	return nil
}
