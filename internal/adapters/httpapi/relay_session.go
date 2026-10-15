package httpapi

import (
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func (h *handler) handleRelaySessionEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if principal(r).Kind != "human" || principal(r).ID == "" {
		writeError(w, r, 403, relaySafeError(&apierrors.ServiceError{Code: apierrors.ScopeDenied}))
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, r, 405, relaySafeError(&apierrors.ServiceError{Code: apierrors.UnsupportedCapability}))
		return
	}
	values, err := url.ParseQuery(r.URL.RawQuery)
	q := workbench.SessionEventsQuery{Page: 1, PageSize: 50}
	bad := err != nil
	for key, value := range values {
		if len(value) != 1 || value[0] == "" {
			bad = true
			continue
		}
		switch key {
		case "task_id":
			u, e := uuid.Parse(value[0])
			if e != nil || u == uuid.Nil || u.String() != value[0] {
				bad = true
			}
			q.TaskID = value[0]
		case "start", "end":
			if _, e := time.Parse(time.RFC3339Nano, value[0]); e != nil {
				bad = true
			}
			if key == "start" {
				q.Start = value[0]
			} else {
				q.End = value[0]
			}
		case "page", "page_size":
			n, e := strconv.Atoi(value[0])
			if e != nil || n < 1 || (key == "page_size" && n > 100) || n > 1000000 {
				bad = true
			}
			if key == "page" {
				q.Page = n
			} else {
				q.PageSize = n
			}
		default:
			bad = true
		}
	}
	if q.Start != "" && q.End != "" {
		a, _ := time.Parse(time.RFC3339Nano, q.Start)
		b, _ := time.Parse(time.RFC3339Nano, q.End)
		if a.After(b) {
			bad = true
		}
	}
	if bad {
		writeError(w, r, 400, relaySafeError(invalid("invalid session event query")))
		return
	}
	if h.services.RelaySessionEvents == nil {
		writeError(w, r, 503, relaySafeError(&apierrors.ServiceError{Code: apierrors.ProviderUnavailable}))
		return
	}
	v, e := h.services.RelaySessionEvents.ReadSessionEvents(r.Context(), q)
	h.relayResult(w, r, v, e)
}
