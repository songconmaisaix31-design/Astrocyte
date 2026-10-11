package httpapi

import (
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type worksRefreshRequest struct {
	attentionapp.CommandMeta
	Roots []string `json:"roots"`
}
type worksDeliveryRequest struct {
	attentionapp.CommandMeta
	Content string `json:"content"`
}
type workMetadataRequest struct {
	attentionapp.CommandMeta
	ExpectedRevision *int `json:"expected_revision"`
}
type workAttitudeRequest struct {
	workMetadataRequest
	Attitude string `json:"attitude"`
}
type workGroupRequest struct {
	workMetadataRequest
	Group *string `json:"group"`
}
type workHandoffRequest struct {
	workMetadataRequest
	Content *string `json:"content"`
}
type workHiddenRequest struct {
	workMetadataRequest
	Hidden *bool `json:"hidden"`
}

func workRevision(c workMetadataRequest) (int, error) {
	if c.ExpectedRevision == nil || *c.ExpectedRevision < 0 {
		return 0, invalid("expected_revision must be a non-negative integer")
	}
	return *c.ExpectedRevision, nil
}

func worksRootKey(root string) string {
	root = filepath.Clean(root)
	if runtime.GOOS == "windows" {
		root = strings.ToLower(root)
	}
	return root
}

// Resolve roots from the existing cached registry; request roots can select
// registered entries but cannot grant a new filesystem boundary.
func (h *handler) worksRoots(r *http.Request, requested []string) ([]string, error) {
	if h.services.RegisteredProjects == nil {
		return nil, apierrors.NewUnsupported("registered_works_roots")
	}
	snapshot, err := h.services.RegisteredProjects.ListRegisteredProjects(r.Context(), localCaller(r))
	if err != nil {
		return nil, err
	}
	allowed := map[string]string{}
	var defaults []string
	for _, p := range snapshot.Projects {
		if !filepath.IsAbs(p.Root) {
			continue
		}
		key := worksRootKey(p.Root)
		if _, ok := allowed[key]; !ok {
			allowed[key] = p.Root
			defaults = append(defaults, p.Root)
		}
	}
	if len(requested) == 0 {
		requested = defaults
	}
	if len(requested) > 64 {
		return nil, invalid("at most 64 registered roots may be selected")
	}
	roots := []string{}
	seen := map[string]bool{}
	for _, root := range requested {
		key := worksRootKey(root)
		registered, ok := allowed[key]
		if !filepath.IsAbs(root) || !ok {
			return nil, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Works root is outside registered metadata roots", RequiredAction: "select_registered_project_root"}
		}
		if !seen[key] {
			roots = append(roots, registered)
			seen[key] = true
		}
	}
	if len(roots) == 0 {
		return nil, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "No registered metadata roots are available", RequiredAction: "refresh_registered_project_metadata"}
	}
	return roots, nil
}

func (h *handler) registerWorks(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/works", func(w http.ResponseWriter, r *http.Request) {
		if h.services.Works == nil {
			h.attentionResult(w, r, http.StatusNotImplemented, nil, apierrors.NewUnsupported("works_album"))
			return
		}
		result, err := h.services.Works.ListWorks(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("snapshot", result), err)
	})
	mux.HandleFunc("POST /api/v1/works/refresh", commandHandler(h, func(c *worksRefreshRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c worksRefreshRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_refresh")
		}
		roots, err := h.worksRoots(r, c.Roots)
		if err != nil {
			return nil, err
		}
		snapshot, err := h.services.Works.RefreshWorks(r.Context(), localCaller(r), workspaceapp.WorksRefreshCommand{Roots: roots})
		return localEnvelope("snapshot", snapshot), err
	}))
	mux.HandleFunc("POST /api/v1/works/delivery", commandHandler(h, func(c *worksDeliveryRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c worksDeliveryRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_delivery")
		}
		if len(c.Content) == 0 || len(c.Content) > 1<<20 {
			return nil, invalid("delivery content must be between 1 byte and 1 MiB")
		}
		snapshot, err := h.services.Works.ImportWorksDelivery(r.Context(), localCaller(r), workspaceapp.WorksDeliveryCommand{Content: []byte(c.Content)})
		return localEnvelope("snapshot", snapshot), err
	}))
	mux.HandleFunc("PUT /api/v1/works/{work_id}/attitude", commandHandler(h, func(c *workAttitudeRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c workAttitudeRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_attitude")
		}
		revision, err := workRevision(c.workMetadataRequest)
		if err != nil {
			return nil, err
		}
		work, err := h.services.Works.SetWorkAttitude(r.Context(), localCaller(r), r.PathValue("work_id"), workspaceapp.WorkAttitudeCommand{ExpectedRevision: revision, Attitude: c.Attitude})
		return localEnvelope("work", work), err
	}))
	mux.HandleFunc("PUT /api/v1/works/{work_id}/group", commandHandler(h, func(c *workGroupRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c workGroupRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_group")
		}
		revision, err := workRevision(c.workMetadataRequest)
		if err != nil {
			return nil, err
		}
		if c.Group == nil {
			return nil, invalid("group is required")
		}
		work, err := h.services.Works.SetWorkGroup(r.Context(), localCaller(r), r.PathValue("work_id"), workspaceapp.WorkGroupCommand{ExpectedRevision: revision, Group: *c.Group})
		return localEnvelope("work", work), err
	}))
	mux.HandleFunc("PUT /api/v1/works/{work_id}/handoff", commandHandler(h, func(c *workHandoffRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c workHandoffRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_handoff")
		}
		revision, err := workRevision(c.workMetadataRequest)
		if err != nil {
			return nil, err
		}
		if c.Content == nil {
			return nil, invalid("content is required")
		}
		work, err := h.services.Works.SetWorkHandoff(r.Context(), localCaller(r), r.PathValue("work_id"), workspaceapp.WorkHandoffCommand{ExpectedRevision: revision, Content: *c.Content})
		return localEnvelope("work", work), err
	}))
	mux.HandleFunc("PUT /api/v1/works/{work_id}/hidden", commandHandler(h, func(c *workHiddenRequest) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c workHiddenRequest) (any, error) {
		if h.services.Works == nil {
			return nil, apierrors.NewUnsupported("works_hidden")
		}
		revision, err := workRevision(c.workMetadataRequest)
		if err != nil {
			return nil, err
		}
		if c.Hidden == nil {
			return nil, invalid("hidden is required")
		}
		work, err := h.services.Works.SetWorkHidden(r.Context(), localCaller(r), r.PathValue("work_id"), workspaceapp.WorkHiddenCommand{ExpectedRevision: revision, Hidden: *c.Hidden})
		return localEnvelope("work", work), err
	}))
}
