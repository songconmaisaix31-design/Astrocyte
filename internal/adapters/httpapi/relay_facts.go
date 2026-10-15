package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type relayFact struct {
	ID          string    `json:"fact_id"`
	TaskID      string    `json:"task_id"`
	TerminalID  string    `json:"terminal_id,omitempty"`
	WorktreeID  string    `json:"worktree_id,omitempty"`
	ReceiptID   string    `json:"receipt_id,omitempty"`
	OperationID string    `json:"operation_id,omitempty"`
	Kind        string    `json:"kind"`
	Status      string    `json:"status"`
	Source      string    `json:"source"`
	Reason      string    `json:"reason,omitempty"`
	Sequence    int64     `json:"sequence"`
	TS          time.Time `json:"ts"`
}

type relayRelation struct {
	TaskID         string    `json:"task_id"`
	PreviousTaskID string    `json:"previous_task_id,omitempty"`
	WorktreeID     string    `json:"worktree_id,omitempty"`
	TerminalID     string    `json:"terminal_id,omitempty"`
	OperationID    string    `json:"operation_id"`
	Kind           string    `json:"kind"`
	Status         string    `json:"status"`
	Source         string    `json:"source"`
	TS             time.Time `json:"ts"`
}

func relayFactValue(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func relayFactIDs(required string, optional ...string) bool {
	for _, value := range append([]string{required}, optional...) {
		if value == "" && value != required {
			continue
		}
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || id.String() != value {
			return false
		}
	}
	return true
}

// 只复制闭合词表和业务身份，私有控制器身份及原生内容不进入响应。
func (h *handler) relayFacts(r *http.Request) (any, error) {
	id := r.PathValue("id")
	if !relayFactIDs(id) {
		return nil, invalid("task UUID required")
	}
	reader, ok := h.services.Relay.(workspace.RelayRuntimeFactsReader)
	if !ok {
		return nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	facts, relations, err := reader.ListRuntimeFacts(r.Context(), localCaller(r), id)
	if err != nil {
		return nil, err
	}
	receipts, err := h.services.Relay.ListReceipts(r.Context(), localCaller(r), id)
	if err != nil {
		return nil, err
	}
	receiptIDs := map[string]bool{}
	for _, receipt := range receipts {
		if receipt.TaskID != id || !relayFactIDs(receipt.OperationID) || !relayFactValue(receipt.Operator, "human:local-human", "service:observer") || !relayFactValue(receipt.Action, "intent", "prepare", "dispatch", "send", "read", "stop", "resume", "transition", "observe", "reconcile_events") || receipt.ID != receipt.Operator+":"+receipt.Action+":"+receipt.OperationID {
			return nil, &apierrors.ServiceError{Code: apierrors.EvidenceMissing}
		}
		receiptIDs[receipt.ID] = true
	}
	ownershipReader, ok := h.services.Relay.(workspace.RelayCurrentOwnershipReader)
	if !ok {
		return nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	ownership, err := ownershipReader.GetCurrentOwnership(r.Context(), localCaller(r), id)
	if err != nil {
		return nil, err
	}
	if !relayFactValue(ownership.Status, "confirmed", "unknown") || !relayFactValue(ownership.Reason, "", "controller_binding_unrecorded", "workspace_identity_unverified", "controller_metadata_unsupported", "controller_not_held", "controller_exited") || ownership.Status == "confirmed" && !relayFactValue(ownership.Reason, "", "controller_exited") {
		return nil, &apierrors.ServiceError{Code: apierrors.EvidenceMissing}
	}
	outFacts := make([]relayFact, 0, len(facts))
	outRelations := make([]relayRelation, 0, len(relations))
	for _, f := range facts {
		if f.TaskID != id || !relayFactIDs(f.ID, f.TaskID, f.TerminalID, f.WorktreeID, f.OperationID) || f.ReceiptID != "" && !receiptIDs[f.ReceiptID] || f.Sequence < 0 || f.TS.IsZero() ||
			!relayFactValue(f.Kind, "action_claimed", "action_result", "observation_rejected", "handoff_declared", "human_target_confirmed", "controller_initialized", "session_confirmed", "input_channel_open", "input_written", "input_closed", "input_accepted", "turn_started", "input_consumption") ||
			!relayFactValue(f.Status, "confirmed", "unknown", "pending", "accepted", "rejected", "delivery_unknown") ||
			!relayFactValue(f.Source, "relay_store", "human_intent", "human_transition", "codex_initialize", "codex_thread_identity", "claude_initialize", "claude_system_init", "pi_get_state", "native_request_pipe", "codex_turn_acknowledgement", "pi_prompt_acknowledgement", "codex_turn_started", "pi_agent_start", "claude_stream_input", "native_protocol", "owned_stdin_pipe", "stdin_pipe", "stdin_eof") ||
			!relayFactValue(f.Reason, "", "client_consumption_not_acknowledged", "operation_not_bound", "controller_evidence_rejected", "not_found", "validation_failed", "unsupported_capability", "version_conflict", "context_stale", "index_stale", "scope_denied", "approval_required", "approval_revoked", "lease_lost", "budget_exhausted", "provider_unavailable", "evidence_missing", "delivery_unknown", "internal_error") {
			return nil, &apierrors.ServiceError{Code: apierrors.EvidenceMissing}
		}
		outFacts = append(outFacts, relayFact{f.ID, f.TaskID, f.TerminalID, f.WorktreeID, f.ReceiptID, f.OperationID, f.Kind, f.Status, f.Source, f.Reason, f.Sequence, f.TS.UTC()})
	}
	for _, relation := range relations {
		if relation.TaskID != id || !relayFactIDs(relation.TaskID, relation.PreviousTask, relation.WorktreeID, relation.TerminalID, relation.OperationID) || relation.OperationID == "" || relation.TS.IsZero() ||
			!relayFactValue(relation.Kind, "new", "resume", "handoff", "unknown") || !relayFactValue(relation.Status, "confirmed", "unknown") || !relayFactValue(relation.Source, "human_intent", "controller_session") {
			return nil, &apierrors.ServiceError{Code: apierrors.EvidenceMissing}
		}
		outRelations = append(outRelations, relayRelation{relation.TaskID, relation.PreviousTask, relation.WorktreeID, relation.TerminalID, relation.OperationID, relation.Kind, relation.Status, relation.Source, relation.TS.UTC()})
	}
	return struct {
		SchemaVersion    int             `json:"schema_version"`
		TaskID           string          `json:"task_id"`
		Facts            []relayFact     `json:"facts"`
		SessionRelations []relayRelation `json:"session_relations"`
		CurrentOwnership relayOwnership  `json:"current_ownership"`
	}{1, id, outFacts, outRelations, relayOwnership{ownership.Status, ownership.Reason}}, nil
}

type relayOwnership struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}
