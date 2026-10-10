package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerTracking(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/source-collections/discover", commandHandler(h, func(c *attentionapp.DiscoverSourceCollectionsCommand) *attentionapp.CommandMeta {
		return &c.CommandMeta
	}, http.StatusOK, func(r *http.Request, c attentionapp.DiscoverSourceCollectionsCommand) (any, error) {
		s, ok := h.services.Tracking.(attentionapp.SourceCollectionDiscoveryService)
		if !ok {
			return nil, apierrors.NewUnsupported("source_collection_discovery")
		}
		return s.DiscoverSourceCollections(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("GET /api/v1/source-collections", func(w http.ResponseWriter, r *http.Request) {
		s, ok := h.services.Tracking.(attentionapp.SourceCollectionService)
		if !ok {
			h.notImplemented("public_source_collections")(w, r)
			return
		}
		result, err := s.ListSourceCollections(r.Context(), principal(r), r.URL.Query().Get("platform"), r.URL.Query().Get("owner_id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	if h.services.Tracking == nil {
		for _, route := range []string{"GET /api/v1/tracking-sources", "POST /api/v1/tracking-sources", "GET /api/v1/tracking-sources/{id}", "POST /api/v1/tracking-sources/{id}/sync", "POST /api/v1/tracking-sources/{id}/recommend", "POST /api/v1/tracking-sources/{id}/select"} {
			mux.HandleFunc(route, h.notImplemented("public_source_tracking"))
		}
		return
	}
	s := h.services.Tracking
	mux.HandleFunc("GET /api/v1/tracking-sources", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.ListTrackingSources(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/tracking-sources/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetTrackingSource(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/tracking-sources", commandHandler(h, func(c *attentionapp.BindTrackingSourceCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.BindTrackingSourceCommand) (any, error) {
		return s.BindTrackingSource(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("POST /api/v1/tracking-sources/{id}/sync", commandHandler(h, func(c *attentionapp.SyncTrackingSourceCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.SyncTrackingSourceCommand) (any, error) {
		return s.SyncTrackingSource(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/tracking-sources/{id}/recommend", commandHandler(h, func(c *attentionapp.RecommendSourceItemsCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.RecommendSourceItemsCommand) (any, error) {
		return s.RecommendSourceItems(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/tracking-sources/{id}/select", commandHandler(h, func(c *attentionapp.SelectSourceItemsCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.SelectSourceItemsCommand) (any, error) {
		return s.SelectSourceItems(r.Context(), principal(r), r.PathValue("id"), c)
	}))
}
