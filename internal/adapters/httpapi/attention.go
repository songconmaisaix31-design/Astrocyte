package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func principal(r *http.Request) attentionapp.Principal {
	p, _ := r.Context().Value(principalKey).(attentionapp.Principal)
	return p
}
func invalid(message string) *apierrors.ServiceError {
	return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: message, RequiredAction: "correct_request_and_resubmit"}
}
func (h *handler) attentionResult(w http.ResponseWriter, r *http.Request, status int, result any, err error) {
	if err == nil {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, status, result)
		return
	}
	var serviceErr *apierrors.ServiceError
	if !errors.As(err, &serviceErr) {
		serviceErr = &apierrors.ServiceError{Code: apierrors.InternalError, Message: "Attention service failed", RequiredAction: "contact_support_with_request_id"}
	}
	switch serviceErr.Code {
	case apierrors.NotFound:
		status = http.StatusNotFound
	case apierrors.ValidationFailed, apierrors.EvidenceMissing:
		status = http.StatusBadRequest
	case apierrors.VersionConflict, apierrors.ContextStale, apierrors.DeliveryUnknown:
		status = http.StatusConflict
	case apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.ApprovalRevoked:
		status = http.StatusForbidden
	case apierrors.UnsupportedCapability:
		status = http.StatusNotImplemented
	case apierrors.ProviderUnavailable:
		status = http.StatusServiceUnavailable
	default:
		status = http.StatusInternalServerError
	}
	writeError(w, r, status, serviceErr)
}
func decodeAttention[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var body T
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return body, invalid("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		return body, invalid("invalid JSON command or unsupported fields")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return body, invalid("one JSON command is required")
	}
	return body, nil
}
func commandMeta(r *http.Request, meta *attentionapp.CommandMeta) error {
	meta.IdempotencyKey = r.Header.Get("Idempotency-Key")
	if meta.SchemaVersion != 1 || meta.RequestID == "" || meta.ExpectedVersion < 1 || meta.IdempotencyKey == "" {
		return invalid("schema_version, request_id, expected_version and Idempotency-Key are required")
	}
	return nil
}
func commandHandler[T any](h *handler, meta func(*T) *attentionapp.CommandMeta, status int, call func(*http.Request, T) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeAttention[T](w, r)
		if err == nil {
			err = commandMeta(r, meta(&body))
		}
		var result any
		if err == nil {
			result, err = call(r, body)
		}
		h.attentionResult(w, r, status, result, err)
	}
}
func revision(r *http.Request) (int, error) {
	n, err := strconv.Atoi(r.PathValue("revision"))
	if err != nil || n < 1 {
		return 0, invalid("revision must be a positive integer")
	}
	return n, nil
}
func (h *handler) registerAttention(mux *http.ServeMux) {
	h.registerClassification(mux)
	s := h.services.Attention
	mux.HandleFunc("POST /api/v1/materials/imports", commandHandler(h, func(c *attentionapp.ImportMaterialCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusAccepted, func(r *http.Request, c attentionapp.ImportMaterialCommand) (any, error) {
		return s.ImportMaterial(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("PATCH /api/v1/materials/{id}", commandHandler(h, func(c *attentionapp.UpdateMaterialCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.UpdateMaterialCommand) (any, error) {
		return s.UpdateMaterial(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("GET /api/v1/materials/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetMaterial(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/materials/{id}/revisions/{revision}/content", func(w http.ResponseWriter, r *http.Request) {
		rev, err := revision(r)
		var result attentionapp.ContentResult
		if err == nil {
			result, err = s.GetContent(r.Context(), principal(r), r.PathValue("id"), rev)
		}
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/materials/{id}/revisions/{revision}/attachments/{name}", func(w http.ResponseWriter, r *http.Request) {
		rev, err := revision(r)
		var result attentionapp.AttachmentContent
		if err == nil {
			result, err = s.GetAttachment(r.Context(), principal(r), r.PathValue("id"), rev, r.PathValue("name"))
		}
		if err != nil {
			h.attentionResult(w, r, http.StatusOK, nil, err)
			return
		}
		w.Header().Set("Content-Type", result.MediaType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": result.Name}))
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(result.Data)
	})
	mux.HandleFunc("POST /api/v1/materials/{id}/uses", commandHandler(h, func(c *attentionapp.RecordUseCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.RecordUseCommand) (any, error) {
		return s.RecordUse(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("GET /api/v1/distillations", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.ListDistillations(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/distillations", commandHandler(h, func(c *attentionapp.RecordDistillationCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.RecordDistillationCommand) (any, error) {
		return s.RecordDistillation(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("GET /api/v1/opportunities/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetOpportunity(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/opportunities", commandHandler(h, func(c *attentionapp.OpportunityCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, c attentionapp.OpportunityCommand) (any, error) {
		return s.CreateOpportunity(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("POST /api/v1/opportunities/{id}/revisions", commandHandler(h, func(c *attentionapp.OpportunityCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.OpportunityCommand) (any, error) {
		return s.ReviseOpportunity(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/opportunities/{id}/reviews", commandHandler(h, func(c *attentionapp.ReviewOpportunityCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, c attentionapp.ReviewOpportunityCommand) (any, error) {
		return s.ReviewOpportunity(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.ListJobs(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetJob(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	for _, action := range []string{"retry", "cancel"} {
		mux.HandleFunc("POST /api/v1/jobs/{id}/"+action, commandHandler(h, func(c *attentionapp.CommandMeta) *attentionapp.CommandMeta { return c }, http.StatusOK, func(r *http.Request, c attentionapp.CommandMeta) (any, error) {
			if action == "retry" {
				return s.RetryJob(r.Context(), principal(r), r.PathValue("id"), c)
			}
			return s.CancelJob(r.Context(), principal(r), r.PathValue("id"), c)
		}))
	}
}
