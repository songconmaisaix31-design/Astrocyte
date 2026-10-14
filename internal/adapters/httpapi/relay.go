package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayReceipt struct {
	ID          string                  `json:"receipt_id"`
	TaskID      string                  `json:"task_id"`
	OperationID string                  `json:"operation_id"`
	Action      string                  `json:"action"`
	Operator    string                  `json:"operator"`
	Status      string                  `json:"status"`
	Error       *apierrors.ServiceError `json:"error"`
	CreatedAt   time.Time               `json:"created_at"`
	FinishedAt  *time.Time              `json:"finished_at"`
}

// Never serialize a runtime/provider error message. Even typed errors can carry
// private paths, command input or native text in their implementation detail.
func relaySafeError(err error) *apierrors.ServiceError {
	code := apierrors.InternalError
	var service *apierrors.ServiceError
	if errors.As(err, &service) {
		code = service.Code
	}
	message, action := "Relay service unavailable", "inspect_local_relay_state"
	switch code {
	case apierrors.NotFound:
		message, action = "Relay resource unavailable", "select_existing_relay_resource"
	case apierrors.ValidationFailed, apierrors.EvidenceMissing:
		message, action = "Invalid relay command or missing evidence", "review_relay_command"
	case apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.ApprovalRevoked:
		message, action = "Current human project permission required", "review_project_permissions"
	case apierrors.VersionConflict, apierrors.ContextStale:
		message, action = "Relay state or workspace conflict", "refresh_relay_state"
	case apierrors.DeliveryUnknown:
		message, action = "Relay delivery unknown", "inspect_original_receipt_do_not_resend"
	case apierrors.UnsupportedCapability:
		message, action = "Relay capability unavailable", "select_supported_relay_capability"
	case apierrors.ProviderUnavailable:
		message, action = "Relay provider unavailable", "inspect_local_relay_provider"
	default:
		code = apierrors.InternalError
	}
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: action, Retryable: false}
}

// Only a canonical UUID already readable through the runtime's public task
// projection can refine the fixed conflict action. Provider text is discarded.
func (h *handler) relaySafeError(r *http.Request, err error) *apierrors.ServiceError {
	safe := relaySafeError(err)
	var service *apierrors.ServiceError
	if safe.Code != apierrors.VersionConflict || !errors.As(err, &service) || h.services.Relay == nil {
		return safe
	}
	const prefix = "view_existing_task:"
	if !strings.HasPrefix(service.RequiredAction, prefix) {
		return safe
	}
	id := strings.TrimPrefix(service.RequiredAction, prefix)
	parsed, e := uuid.Parse(id)
	if e != nil || parsed == uuid.Nil || parsed.String() != id {
		return safe
	}
	view, e := h.services.Relay.GetTask(r.Context(), localCaller(r), id)
	if e == nil && view.Task.ID == id {
		safe.RequiredAction = prefix + id
	}
	return safe
}

func (h *handler) publicRelayReceipt(req *http.Request, v domain.RelayReceipt) relayReceipt {
	r := relayReceipt{ID: v.ID, TaskID: v.TaskID, OperationID: v.OperationID, Action: v.Action, Operator: v.Operator, Status: v.Status, CreatedAt: v.CreatedAt, FinishedAt: v.FinishedAt}
	if v.Error != nil {
		r.Error = h.relaySafeError(req, &apierrors.ServiceError{Code: apierrors.Code(v.Error.Code), RequiredAction: v.Error.RequiredAction})
		r.Error.RequestID = v.OperationID
	}
	return r
}

func publicRelayView(v domain.RelayTaskView) domain.RelayTaskView {
	if v.Transitions == nil {
		v.Transitions = []domain.RelayTransition{}
	}
	if v.EventPublication.Status == "" {
		v.EventPublication.Status = "none"
	}
	return v
}

// A closed six-field projection is the only event serialization boundary.
func publicRelayEvents(items []domain.RelayEvent) []workbench.Event {
	items = append([]domain.RelayEvent(nil), items...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].TS.Equal(items[j].TS) {
			return items[i].ID > items[j].ID
		}
		return items[i].TS.After(items[j].TS)
	})
	result := make([]workbench.Event, 0, len(items))
	for _, v := range items {
		result = append(result, workbench.Event{ID: v.ID, Kind: v.Kind, Topic: v.Topic, Tokens: v.Tokens, Scope: v.Scope, TS: v.TS.UTC().Format(time.RFC3339Nano)})
	}
	return result
}

func (h *handler) relayResult(w http.ResponseWriter, r *http.Request, value any, err error) {
	if err != nil {
		h.attentionResult(w, r, 0, nil, h.relaySafeError(r, err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, value)
}

func relayGuard(w http.ResponseWriter, r *http.Request, writing bool) bool {
	if principal(r).Kind != "human" || principal(r).ID == "" {
		writeError(w, r, 403, relaySafeError(&apierrors.ServiceError{Code: apierrors.ScopeDenied}))
		return false
	}
	if (!writing && r.Method != http.MethodGet) || r.URL.RawQuery != "" {
		writeError(w, r, 400, relaySafeError(invalid("invalid relay request")))
		return false
	}
	if writing {
		if _, err := uuid.Parse(r.Header.Get("Idempotency-Key")); err != nil {
			writeError(w, r, 400, relaySafeError(invalid("operation UUID required")))
			return false
		}
	}
	return true
}

func relayCommand[T any](h *handler, call func(*http.Request, T) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !relayGuard(w, r, true) {
			return
		}
		body, err := decodeAttention[T](w, r)
		var result any
		if err == nil {
			result, err = call(r, body)
		}
		h.relayResult(w, r, result, err)
	}
}

func relayTaskResult(v domain.RelayTaskView) any {
	return struct {
		SchemaVersion int                  `json:"schema_version"`
		View          domain.RelayTaskView `json:"view"`
	}{1, publicRelayView(v)}
}
func (h *handler) relayActionResult(r *http.Request, v domain.RelayActionResult) any {
	return struct {
		SchemaVersion int                  `json:"schema_version"`
		View          domain.RelayTaskView `json:"view"`
		Receipt       relayReceipt         `json:"receipt"`
	}{1, publicRelayView(v.View), h.publicRelayReceipt(r, v.Receipt)}
}

func (h *handler) registerRelay(mux *http.ServeMux) {
	s := h.services.Relay
	read := func(call func(*http.Request) (any, error)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !relayGuard(w, r, false) {
				return
			}
			value, err := call(r)
			h.relayResult(w, r, value, err)
		}
	}
	if s == nil {
		for _, route := range []string{"GET /api/v1/relay/agents", "GET /api/v1/relay/projects", "GET /api/v1/relay/tasks", "GET /api/v1/relay/tasks/{id}", "GET /api/v1/relay/tasks/{id}/events", "GET /api/v1/relay/tasks/{id}/receipts", "POST /api/v1/relay/intents", "POST /api/v1/relay/intents/{id}/confirm", "POST /api/v1/relay/tasks/{id}/send", "POST /api/v1/relay/tasks/{id}/read", "POST /api/v1/relay/tasks/{id}/stop", "POST /api/v1/relay/tasks/{id}/resume", "POST /api/v1/relay/tasks/{id}/transition", "POST /api/v1/relay/tasks/{id}/events/reconcile"} {
			mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				h.relayResult(w, r, nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable})
			})
		}
		return
	}
	mux.HandleFunc("GET /api/v1/relay/agents", read(func(r *http.Request) (any, error) {
		a, err := s.ListAgents(r.Context(), localCaller(r))
		if err != nil {
			return nil, err
		}
		if a == nil {
			a = []domain.AgentSpec{}
		}
		for i := range a {
			if a[i].LaunchCmd == nil {
				a[i].LaunchCmd = []string{}
			}
			if a[i].Capabilities == nil {
				a[i].Capabilities = map[string]domain.CapabilityObservation{}
			}
		}
		inventory := []any{}
		if h.services.LocalAgents != nil {
			v, e := h.services.LocalAgents.ListLocalAgents(r.Context())
			if e != nil {
				return nil, e
			}
			if v.Items != nil {
				inventory = v.Items
			}
		}
		return struct {
			SchemaVersion int                `json:"schema_version"`
			Agents        []domain.AgentSpec `json:"agents"`
			Inventory     []any              `json:"inventory"`
		}{1, a, inventory}, nil
	}))
	mux.HandleFunc("GET /api/v1/relay/projects", read(h.relayProjects))
	mux.HandleFunc("GET /api/v1/relay/tasks", read(func(r *http.Request) (any, error) {
		v, err := s.ListTasks(r.Context(), localCaller(r))
		if err != nil {
			return nil, err
		}
		if v == nil {
			v = []domain.RelayTaskView{}
		}
		ids := []string{}
		seen := map[string]bool{}
		for i := range v {
			v[i] = publicRelayView(v[i])
			receipts, e := s.ListReceipts(r.Context(), localCaller(r), v[i].Task.ID)
			if e != nil {
				return nil, e
			}
			for _, receipt := range receipts {
				if receipt.Action == "dispatch" && receipt.Status == "accepted" && !seen[v[i].Task.WorkID] {
					ids = append(ids, v[i].Task.WorkID)
					seen[v[i].Task.WorkID] = true
				}
			}
		}
		sort.Strings(ids)
		return struct {
			SchemaVersion int                    `json:"schema_version"`
			Tasks         []domain.RelayTaskView `json:"tasks"`
			Dispatched    []string               `json:"dispatched_work_ids"`
		}{1, v, ids}, nil
	}))
	mux.HandleFunc("GET /api/v1/relay/tasks/{id}", read(func(r *http.Request) (any, error) {
		v, e := s.GetTask(r.Context(), localCaller(r), r.PathValue("id"))
		return relayTaskResult(v), e
	}))
	mux.HandleFunc("GET /api/v1/relay/tasks/{id}/events", read(func(r *http.Request) (any, error) {
		v, e := s.ListEvents(r.Context(), localCaller(r), r.PathValue("id"))
		return struct {
			SchemaVersion int               `json:"schema_version"`
			Events        []workbench.Event `json:"events"`
		}{1, publicRelayEvents(v)}, e
	}))
	mux.HandleFunc("GET /api/v1/relay/tasks/{id}/receipts", read(func(r *http.Request) (any, error) {
		v, e := s.ListReceipts(r.Context(), localCaller(r), r.PathValue("id"))
		out := make([]relayReceipt, 0, len(v))
		for _, item := range v {
			out = append(out, h.publicRelayReceipt(r, item))
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
		return struct {
			SchemaVersion int            `json:"schema_version"`
			Receipts      []relayReceipt `json:"receipts"`
		}{1, out}, e
	}))
	mux.HandleFunc("POST /api/v1/relay/intents", relayCommand(h, func(r *http.Request, c workspace.RelayIntentCommand) (any, error) {
		v, e := s.CreateIntent(r.Context(), localCaller(r), c)
		return relayTaskResult(v), e
	}))
	mux.HandleFunc("POST /api/v1/relay/intents/{id}/confirm", relayCommand(h, func(r *http.Request, c workspace.RelayConfirmCommand) (any, error) {
		v, e := s.ConfirmIntent(r.Context(), localCaller(r), r.PathValue("id"), c)
		return h.relayActionResult(r, v), e
	}))
	mux.HandleFunc("POST /api/v1/relay/tasks/{id}/send", relayCommand(h, func(r *http.Request, c workspace.RelaySendCommand) (any, error) {
		v, e := s.Send(r.Context(), localCaller(r), r.PathValue("id"), c)
		return h.relayActionResult(r, v), e
	}))
	for _, a := range []struct {
		name string
		call func(*http.Request) (domain.RelayActionResult, error)
	}{{"read", func(r *http.Request) (domain.RelayActionResult, error) {
		return s.Read(r.Context(), localCaller(r), r.PathValue("id"))
	}}, {"stop", func(r *http.Request) (domain.RelayActionResult, error) {
		return s.Stop(r.Context(), localCaller(r), r.PathValue("id"))
	}}, {"resume", func(r *http.Request) (domain.RelayActionResult, error) {
		return s.Resume(r.Context(), localCaller(r), r.PathValue("id"))
	}}} {
		mux.HandleFunc("POST /api/v1/relay/tasks/{id}/"+a.name, relayCommand(h, func(r *http.Request, _ struct{}) (any, error) { v, e := a.call(r); return h.relayActionResult(r, v), e }))
	}
	mux.HandleFunc("POST /api/v1/relay/tasks/{id}/transition", relayCommand(h, func(r *http.Request, c workspace.RelayTransitionCommand) (any, error) {
		v, e := s.Transition(r.Context(), localCaller(r), r.PathValue("id"), c)
		return relayTaskResult(v), e
	}))
	mux.HandleFunc("POST /api/v1/relay/tasks/{id}/events/reconcile", relayCommand(h, func(r *http.Request, _ struct{}) (any, error) {
		lifecycle, ok := s.(workspace.RelayLifecycle)
		if !ok {
			return nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
		}
		v, e := lifecycle.ReconcileEvents(r.Context(), localCaller(r), r.PathValue("id"))
		return relayTaskResult(v), e
	}))
}

type relayProject struct {
	ID               string   `json:"project_id"`
	Name             string   `json:"name"`
	Root             string   `json:"root"`
	Permission       string   `json:"permission_level"`
	Revision         int      `json:"project_settings_revision"`
	Subdirs          []string `json:"allowed_subdirs"`
	Actions          []string `json:"allowed_actions"`
	AgentControl     bool     `json:"allow_agent_control"`
	AllowDirectory   bool     `json:"allow_directory"`
	ExpandReferences bool     `json:"expand_references"`
}

func (h *handler) relayProjects(r *http.Request) (any, error) {
	if h.services.LocalProjects == nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
	}
	items, err := h.services.LocalProjects.ListProjects(r.Context(), localCaller(r))
	if err != nil {
		return nil, err
	}
	projects := make([]relayProject, 0, len(items))
	for _, p := range items {
		level := "A"
		if p.Settings.AllowDirectory {
			level = "B"
		}
		if p.Settings.ExpandReferences {
			level = "C"
		}
		subs := append([]string{}, p.Settings.AllowedSubdirs...)
		actions := append([]string{}, p.Settings.AllowedActions...)
		projects = append(projects, relayProject{p.ID, p.Name, p.Root, level, p.Settings.Revision, subs, actions, p.Settings.AllowAgentControl, p.Settings.AllowDirectory, p.Settings.ExpandReferences})
	}
	return struct {
		SchemaVersion int            `json:"schema_version"`
		Projects      []relayProject `json:"projects"`
	}{1, projects}, nil
}
