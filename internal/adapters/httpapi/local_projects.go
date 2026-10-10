package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const workspaceCallerKey contextKey = "workspace_caller"

func localCaller(r *http.Request) domain.Caller {
	if c, ok := r.Context().Value(workspaceCallerKey).(domain.Caller); ok {
		c.OperationID = r.Header.Get("Idempotency-Key")
		return c
	}
	p := principal(r)
	return domain.Caller{ID: p.ID, Kind: p.Kind, OperationID: r.Header.Get("Idempotency-Key")}
}

func localList[T any](items []T) apierrors.ListResult {
	result := apierrors.EmptyList()
	for _, item := range items {
		result.Items = append(result.Items, item)
	}
	return result
}
func localEnvelope(key string, value any) any { return map[string]any{"schema_version": 1, key: value} }

type registerLocalCommand struct {
	attentionapp.CommandMeta
	domain.RegisterProjectCommand
}
type discoverLocalCommand struct {
	attentionapp.CommandMeta
	Root string `json:"root"`
}
type localSettingsCommand struct {
	attentionapp.CommandMeta
	domain.SettingsCommand
}
type localGrantCommand struct {
	attentionapp.CommandMeta
	AgentID string   `json:"agent_id"`
	Actions []string `json:"actions"`
}
type localAgentCommand struct {
	attentionapp.CommandMeta
	AgentID string `json:"agent_id"`
}
type localContextCommand struct {
	attentionapp.CommandMeta
	domain.ContextRequest
}
type localNativeCommand struct {
	attentionapp.CommandMeta
	domain.NativeCommand
}
type localMessageCommand struct {
	attentionapp.CommandMeta
	Message string `json:"message"`
}
type localDiscoverSessionCommand struct {
	attentionapp.CommandMeta
	CLI string `json:"cli"`
}

func (h *handler) registerLocalProjects(mux *http.ServeMux) {
	s := h.services.LocalProjects
	if s == nil {
		for _, route := range []string{
			"GET /api/v1/local-projects", "POST /api/v1/local-projects", "POST /api/v1/local-projects/discover",
			"PUT /api/v1/local-projects/{id}/settings", "POST /api/v1/local-projects/{id}/grants", "POST /api/v1/local-projects/{id}/grants/revoke", "POST /api/v1/local-projects/{id}/agent-token",
			"POST /api/v1/local-projects/{id}/context", "GET /api/v1/local-projects/{id}/sessions", "POST /api/v1/local-projects/{id}/sessions",
			"POST /api/v1/local-projects/{id}/sessions/discover", "GET /api/v1/local-projects/{id}/sessions/{session_id}/context",
			"POST /api/v1/local-projects/{id}/sessions/{session_id}/resume", "POST /api/v1/local-projects/{id}/sessions/{session_id}/send", "POST /api/v1/local-projects/{id}/sessions/{session_id}/stop", "GET /api/v1/local-projects/{id}/sessions/{session_id}/observe",
		} {
			mux.HandleFunc(route, h.notImplemented("local_project_control"))
		}
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects", func(w http.ResponseWriter, r *http.Request) {
		projects, err := s.ListProjects(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localList(projects), err)
	})
	mux.HandleFunc("POST /api/v1/local-projects", commandHandler(h, func(c *registerLocalCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c registerLocalCommand) (any, error) {
		p, err := s.RegisterProject(r.Context(), localCaller(r), c.RegisterProjectCommand)
		return localEnvelope("project", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/discover", commandHandler(h, func(c *discoverLocalCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c discoverLocalCommand) (any, error) {
		p, err := s.DiscoverProjects(r.Context(), localCaller(r), c.Root)
		return localList(p), err
	}))
	mux.HandleFunc("PUT /api/v1/local-projects/{id}/settings", commandHandler(h, func(c *localSettingsCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localSettingsCommand) (any, error) {
		p, err := s.SetProjectSettings(r.Context(), localCaller(r), r.PathValue("id"), c.SettingsCommand)
		return localEnvelope("project", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/grants", commandHandler(h, func(c *localGrantCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localGrantCommand) (any, error) {
		p, err := s.GrantProjectAgent(r.Context(), localCaller(r), r.PathValue("id"), domain.ProjectGrant{AgentID: c.AgentID, Actions: c.Actions})
		return localEnvelope("grant", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/grants/revoke", commandHandler(h, func(c *localAgentCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localAgentCommand) (any, error) {
		p, err := s.RevokeProjectAgent(r.Context(), localCaller(r), r.PathValue("id"), c.AgentID)
		return localEnvelope("grant", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/agent-token", commandHandler(h, func(c *localAgentCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localAgentCommand) (any, error) {
		p, err := s.IssueProjectAgentToken(r.Context(), localCaller(r), r.PathValue("id"), c.AgentID)
		return localEnvelope("credential", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/context", commandHandler(h, func(c *localContextCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localContextCommand) (any, error) {
		return s.ReadProjectContext(r.Context(), localCaller(r), r.PathValue("id"), c.ContextRequest)
	}))
	mux.HandleFunc("GET /api/v1/local-projects/{id}/sessions", func(w http.ResponseWriter, r *http.Request) {
		sessions, err := s.ListNativeSessions(r.Context(), localCaller(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, localList(sessions), err)
	})
	mux.HandleFunc("POST /api/v1/local-projects/{id}/sessions/discover", commandHandler(h, func(c *localDiscoverSessionCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localDiscoverSessionCommand) (any, error) {
		p, err := s.DiscoverNativeSessions(r.Context(), localCaller(r), r.PathValue("id"), c.CLI)
		return localList(p), err
	}))
	mux.HandleFunc("GET /api/v1/local-projects/{id}/sessions/{session_id}/context", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.ReadNativeContext(r.Context(), localCaller(r), r.PathValue("id"), r.PathValue("session_id"))
		h.attentionResult(w, r, http.StatusOK, localList(p), err)
	})
	mux.HandleFunc("POST /api/v1/local-projects/{id}/sessions", commandHandler(h, func(c *localNativeCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localNativeCommand) (any, error) {
		c.NativeCommand.OperationID = c.IdempotencyKey
		p, err := s.StartNativeSession(r.Context(), localCaller(r), r.PathValue("id"), c.NativeCommand)
		return localEnvelope("session", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/sessions/{session_id}/resume", commandHandler(h, func(c *localNativeCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localNativeCommand) (any, error) {
		c.NativeCommand.OperationID = c.IdempotencyKey
		p, err := s.ResumeNativeSession(r.Context(), localCaller(r), r.PathValue("id"), r.PathValue("session_id"), c.NativeCommand)
		return localEnvelope("session", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/sessions/{session_id}/send", commandHandler(h, func(c *localMessageCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c localMessageCommand) (any, error) {
		p, err := s.SendNativeMessage(r.Context(), localCaller(r), r.PathValue("id"), r.PathValue("session_id"), c.Message)
		return localEnvelope("observation", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/sessions/{session_id}/stop", commandHandler(h, func(c *attentionapp.CommandMeta) *attentionapp.CommandMeta { return c }, http.StatusOK, func(r *http.Request, c attentionapp.CommandMeta) (any, error) {
		p, err := s.StopNativeSession(r.Context(), localCaller(r), r.PathValue("id"), r.PathValue("session_id"))
		return localEnvelope("observation", p), err
	}))
	mux.HandleFunc("GET /api/v1/local-projects/{id}/sessions/{session_id}/observe", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.ObserveNativeSession(r.Context(), localCaller(r), r.PathValue("id"), r.PathValue("session_id"))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("observation", p), err)
	})
}
