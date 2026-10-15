package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func swarmCaller(r *http.Request) domain.Principal {
	p := localCaller(r)
	return domain.Principal{ID: p.ID, Kind: p.Kind, ProjectID: p.ProjectID}
}

func swarmError(err error) *apierrors.ServiceError {
	var service *apierrors.ServiceError
	if errors.As(err, &service) {
		return service
	}
	var rule *domain.RuleError
	if errors.As(err, &rule) {
		return &apierrors.ServiceError{Code: apierrors.Code(rule.Code), Message: rule.Message, RequiredAction: "inspect_current_state_and_permissions"}
	}
	return &apierrors.ServiceError{Code: apierrors.InternalError, Message: "Swarm operation failed", RequiredAction: "contact_support_with_request_id"}
}

func (h *handler) swarmResult(w http.ResponseWriter, r *http.Request, result any, err error) {
	if err == nil {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, result)
		return
	}
	e := swarmError(err)
	status := http.StatusInternalServerError
	switch e.Code {
	case apierrors.NotFound:
		status = http.StatusNotFound
	case apierrors.ValidationFailed, apierrors.EvidenceMissing:
		status = http.StatusBadRequest
	case apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.ApprovalRevoked:
		status = http.StatusForbidden
	case apierrors.VersionConflict, apierrors.ContextStale, apierrors.IndexStale, apierrors.DeliveryUnknown, apierrors.LeaseLost, apierrors.BudgetExhausted:
		status = http.StatusConflict
	case apierrors.UnsupportedCapability:
		status = http.StatusNotImplemented
	case apierrors.ProviderUnavailable:
		status = http.StatusServiceUnavailable
	}
	writeError(w, r, status, e)
}

func swarmCommand[T any](h *handler, metadata func(*T) *domain.Meta, call func(*http.Request, T) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := decodeAttention[T](w, r)
		if err == nil {
			m := metadata(&c)
			m.IdempotencyKey = r.Header.Get("Idempotency-Key")
			if m.SchemaVersion != 1 || m.RequestID == "" || m.ExpectedVersion < 0 || m.IdempotencyKey == "" {
				err = invalid("Swarm command requires schema_version, request_id, expected_version and Idempotency-Key")
			}
		}
		var result any
		if err == nil {
			result, err = call(r, c)
		}
		h.swarmResult(w, r, result, err)
	}
}

type swarmDetailEnvelope struct {
	SchemaVersion int `json:"schema_version"`
	swarm.MissionDetail
}

func swarmArtifactVersion(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("version")
	if raw == "" {
		return 0, nil
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 1 {
		return 0, invalid("artifact version must be a positive integer")
	}
	return version, nil
}

// H 在同一个已认证主 router 注册，移除相同路径的 S0 占位项。
func (h *handler) registerSwarmS3(mux *http.ServeMux, s *swarm.Service) {
	if s == nil {
		return
	}
	mux.HandleFunc("GET /api/v1/missions", func(w http.ResponseWriter, r *http.Request) {
		rows, err := s.ListMissionsFor(r.Context(), swarmCaller(r))
		h.swarmResult(w, r, localList(rows), err)
	})
	mux.HandleFunc("GET /api/v1/missions/{id}", func(w http.ResponseWriter, r *http.Request) {
		detail, err := s.ReadMission(r.Context(), swarmCaller(r), r.PathValue("id"))
		h.swarmResult(w, r, swarmDetailEnvelope{SchemaVersion: 1, MissionDetail: detail}, err)
	})
	mux.HandleFunc("GET /api/v1/missions/{id}/work-items", func(w http.ResponseWriter, r *http.Request) {
		detail, err := s.ReadMission(r.Context(), swarmCaller(r), r.PathValue("id"))
		h.swarmResult(w, r, localList(detail.WorkItems), err)
	})
	mux.HandleFunc("GET /api/v1/missions/{id}/timeline", func(w http.ResponseWriter, r *http.Request) {
		detail, err := s.ReadMission(r.Context(), swarmCaller(r), r.PathValue("id"))
		h.swarmResult(w, r, localList(detail.Timeline), err)
	})
	mux.HandleFunc("GET /api/v1/artifacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("mission_id")
		var err error
		var a domain.Artifact
		if id == "" {
			err = invalid("mission_id is required")
		} else {
			var version int
			version, err = swarmArtifactVersion(r)
			if err == nil {
				a, err = s.ReadArtifactVersion(r.Context(), swarmCaller(r), id, r.PathValue("id"), version)
			}
		}
		h.swarmResult(w, r, localEnvelope("artifact", a), err)
	})
	mux.HandleFunc("POST /api/v1/missions", swarmCommand(h, func(c *swarm.CreateMissionCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.CreateMissionCommand) (any, error) {
		if c.EventID != "" {
			return nil, invalid("event_id belongs to the trusted event consumer")
		}
		m, err := s.CreateMission(r.Context(), swarmCaller(r), c)
		return localEnvelope("mission", m), err
	}))
	mux.HandleFunc("POST /api/v1/missions/{id}/work-items", swarmCommand(h, func(c *swarm.AddWorkItemCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.AddWorkItemCommand) (any, error) {
		if c.MissionID != "" && c.MissionID != r.PathValue("id") {
			return nil, invalid("mission_id differs from route")
		}
		c.MissionID = r.PathValue("id")
		w, err := s.AddWorkItem(r.Context(), swarmCaller(r), c)
		return localEnvelope("work_item", w), err
	}))
	for _, route := range []struct {
		name string
		call func(*http.Request, swarm.TransitionCommand) (domain.Mission, error)
	}{
		{"pause", func(r *http.Request, c swarm.TransitionCommand) (domain.Mission, error) {
			return s.PauseMission(r.Context(), swarmCaller(r), c)
		}},
		{"cancel", func(r *http.Request, c swarm.TransitionCommand) (domain.Mission, error) {
			return s.CancelMission(r.Context(), swarmCaller(r), c)
		}},
		{"recover", func(r *http.Request, c swarm.TransitionCommand) (domain.Mission, error) {
			return s.RecoverMission(r.Context(), swarmCaller(r), c)
		}},
	} {
		mux.HandleFunc("POST /api/v1/missions/{id}/"+route.name, swarmCommand(h, func(c *swarm.TransitionCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.TransitionCommand) (any, error) {
			if c.MissionID != "" && c.MissionID != r.PathValue("id") {
				return nil, invalid("mission_id differs from route")
			}
			c.MissionID = r.PathValue("id")
			m, err := route.call(r, c)
			return localEnvelope("mission", m), err
		}))
	}
	h.registerSwarmWorkItems(mux, s)
}
