package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

func (h *handler) registerWorkbench(mux *http.ServeMux) {
	for _, name := range []string{"health", "today", "focus", "topics/evidence", "agents", "events", "filters"} {
		mux.HandleFunc("GET /api/v1/cognition/"+name, h.handleWorkbench)
	}
	// Go's GET pattern also matches HEAD; handleWorkbench explicitly rejects it.
	mux.HandleFunc("/api/v1/cognition/", func(w http.ResponseWriter, r *http.Request) {
		workbenchError(w, r, 404, apierrors.NotFound, "Cognition endpoint unavailable", false)
	})
}
func workbenchError(w http.ResponseWriter, r *http.Request, status int, code apierrors.Code, message string, retry bool) {
	writeError(w, r, status, &apierrors.ServiceError{Code: code, Message: message, Retryable: retry, RequiredAction: "check_filters_or_refresh_local_snapshot"})
}
func (h *handler) handleWorkbench(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		workbenchError(w, r, 405, apierrors.UnsupportedCapability, "Cognition views require GET", false)
		return
	}
	if h.services.Workbench == nil {
		workbenchError(w, r, 503, apierrors.ProviderUnavailable, "Workbench snapshot unavailable", true)
		return
	}
	values, e := url.ParseQuery(r.URL.RawQuery)
	if e != nil {
		workbenchError(w, r, 400, apierrors.ValidationFailed, "Invalid cognition query", false)
		return
	}
	q := workbench.Query{Topic: values.Get("topic"), Kind: values.Get("kind"), Source: values.Get("source"), Start: values.Get("start"), End: values.Get("end"), Page: 1, PageSize: 50}
	for _, p := range []struct {
		name string
		dest *int
	}{{"page", &q.Page}, {"page_size", &q.PageSize}} {
		if x := values.Get(p.name); x != "" {
			n, e := strconv.Atoi(x)
			if e != nil || n < 1 {
				workbenchError(w, r, 400, apierrors.ValidationFailed, "Invalid cognition query", false)
				return
			}
			*p.dest = n
		}
	}
	v, e := h.services.Workbench.Read(r.Context(), strings.TrimPrefix(r.URL.Path, "/api/v1/cognition/"), q)
	if e != nil {
		if errors.Is(e, workbench.ErrInvalidQuery) {
			workbenchError(w, r, 400, apierrors.ValidationFailed, "Invalid cognition query", false)
		} else {
			workbenchError(w, r, 503, apierrors.ProviderUnavailable, "Workbench snapshot unavailable", true)
		}
		return
	}
	w.Header().Set("X-Snapshot-At", v.SnapshotAt)
	switch {
	case v.Health != nil:
		writeJSON(w, 200, v.Health)
	case v.Today != nil:
		writeJSON(w, 200, v.Today)
	case v.Focus != nil:
		writeJSON(w, 200, v.Focus)
	case v.Agents != nil:
		writeJSON(w, 200, v.Agents)
	case v.Events != nil:
		writeJSON(w, 200, v.Events)
	case v.Evidence != nil:
		writeJSON(w, 200, v.Evidence)
	case v.Filters != nil:
		writeJSON(w, 200, v.Filters)
	default:
		workbenchError(w, r, 503, apierrors.ProviderUnavailable, "Workbench snapshot unavailable", true)
	}
}
