package httpapi

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// H在既有认证router内注册；各能力可以由不同后端提供。
type RepositoryContextServices struct {
	Map         workspace.RepositoryMap
	Search      workspace.CodeSearch
	Impact      workspace.ImpactQuery
	Indexes     workspace.RepositoryIndexCommands
	MapCommands workspace.RepositoryMapCommands
}

type repositoryContextCommand struct {
	SchemaVersion int                     `json:"schema_version"`
	Purpose       string                  `json:"purpose"`
	Roots         []string                `json:"roots"`
	Budget        domain.RepositoryBudget `json:"budget"`
	Query         string                  `json:"query,omitempty"`
	SymbolID      string                  `json:"symbol_id,omitempty"`
}

type repositoryContextIndexCommand struct {
	repositoryContextCommand
	Action string `json:"action"`
}

func repositoryContextResult(w http.ResponseWriter, r *http.Request, result any, err error) {
	w.Header().Set("Cache-Control", "no-store")
	if err == nil {
		writeJSON(w, http.StatusOK, localEnvelope("repository_context", result))
		return
	}
	var serviceErr *apierrors.ServiceError
	if !errors.As(err, &serviceErr) {
		serviceErr = &apierrors.ServiceError{Code: apierrors.InternalError, Message: "repository context operation failed", RequiredAction: "inspect_local_repository_diagnostics"}
	}
	status := http.StatusInternalServerError
	switch serviceErr.Code {
	case apierrors.NotFound:
		status = http.StatusNotFound
	case apierrors.ValidationFailed, apierrors.EvidenceMissing:
		status = http.StatusBadRequest
	case apierrors.IndexStale, apierrors.ContextStale, apierrors.BudgetExhausted, apierrors.VersionConflict, apierrors.DeliveryUnknown:
		status = http.StatusConflict
	case apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.ApprovalRevoked:
		status = http.StatusForbidden
	case apierrors.UnsupportedCapability:
		status = http.StatusNotImplemented
	case apierrors.ProviderUnavailable:
		status = http.StatusServiceUnavailable
	}
	writeError(w, r, status, serviceErr)
}

func repositoryContextUnavailable() error {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "repository context capability is not configured", RequiredAction: "configure_the_verified_local_repository_adapter"}
}

func repositoryContextQuery(r *http.Request, body repositoryContextCommand) (domain.RepositoryQuery, error) {
	var result domain.RepositoryQuery
	if body.SchemaVersion != 1 || strings.TrimSpace(body.Purpose) == "" || len(body.Purpose) > 512 {
		return result, invalid("schema_version 1 and a purpose of at most 512 UTF-8 bytes are required")
	}
	if len(body.Roots) == 0 || len(body.Roots) > 32 {
		return result, invalid("one to 32 explicit relative roots are required")
	}
	for _, root := range body.Roots {
		if root == "" || len(root) > 4096 || strings.ContainsAny(root, "\\:\x00\r\n") || path.IsAbs(root) || path.Clean(root) != root || root == ".." || strings.HasPrefix(root, "../") {
			return result, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "HTTP repository roots must be safe project-relative paths", RequiredAction: "choose_an_approved_project_relative_root"}
		}
	}
	result = domain.RepositoryQuery{ProjectID: r.PathValue("id"), Purpose: body.Purpose, Roots: body.Roots, Text: body.Query, SymbolID: body.SymbolID, Budget: body.Budget}
	if result.ProjectID == "" {
		return result, invalid("a registered project route is required")
	}
	return result, nil
}

func repositoryMapQuery(r *http.Request) (domain.RepositoryQuery, error) {
	params := r.URL.Query()
	for key := range params {
		switch key {
		case "purpose", "root", "max_files", "max_bytes", "max_results", "max_depth":
		default:
			return domain.RepositoryQuery{}, invalid("unsupported repository map query parameter")
		}
		if key != "root" && len(params[key]) != 1 {
			return domain.RepositoryQuery{}, invalid("repository map scalar parameters must occur exactly once")
		}
	}
	parse := func(key string) (int64, error) {
		values := params[key]
		if len(values) != 1 {
			return 0, invalid("explicit repository query budgets are required")
		}
		value, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || value < 1 {
			return 0, invalid("repository budgets must be positive integers")
		}
		return value, nil
	}
	files, err := parse("max_files")
	if err != nil {
		return domain.RepositoryQuery{}, err
	}
	bytes, err := parse("max_bytes")
	if err != nil {
		return domain.RepositoryQuery{}, err
	}
	results, err := parse("max_results")
	if err != nil {
		return domain.RepositoryQuery{}, err
	}
	depth, err := parse("max_depth")
	if err != nil {
		return domain.RepositoryQuery{}, err
	}
	if files > 20000 || results > 200 || depth > 5 || bytes > 128*1024*1024 {
		return domain.RepositoryQuery{}, &apierrors.ServiceError{Code: apierrors.BudgetExhausted, Message: "repository query budget exceeds the supported limit", RequiredAction: "reduce_the_explicit_repository_budget"}
	}
	body := repositoryContextCommand{SchemaVersion: 1, Purpose: params.Get("purpose"), Roots: params["root"], Budget: domain.RepositoryBudget{MaxFiles: int(files), MaxBytes: bytes, MaxResults: int(results), MaxDepth: int(depth)}}
	return repositoryContextQuery(r, body)
}

func repositoryContextHuman(r *http.Request) error {
	c := localCaller(r)
	if c.Kind != "human" || c.ID == "" {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "repository index maintenance requires the human control plane", RequiredAction: "use_the_current_human_project_session"}
	}
	return nil
}

// 注册生产路由，不创建认证会话、索引器或独立HTTP服务器。
func RegisterRepositoryContextRoutes(mux *http.ServeMux, s RepositoryContextServices) {
	const prefix = "/api/v1/local-projects/{id}/repository-context"
	mux.HandleFunc("GET "+prefix+"/map", func(w http.ResponseWriter, r *http.Request) {
		q, err := repositoryMapQuery(r)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		if s.Map == nil {
			repositoryContextResult(w, r, nil, repositoryContextUnavailable())
			return
		}
		result, err := s.Map.ReadRepositoryMap(r.Context(), localCaller(r), q)
		repositoryContextResult(w, r, result, err)
	})
	readQuery := func(w http.ResponseWriter, r *http.Request) (domain.RepositoryQuery, error) {
		body, err := decodeAttention[repositoryContextCommand](w, r)
		if err != nil {
			return domain.RepositoryQuery{}, err
		}
		return repositoryContextQuery(r, body)
	}
	mux.HandleFunc("POST "+prefix+"/search", func(w http.ResponseWriter, r *http.Request) {
		q, err := readQuery(w, r)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		if s.Search == nil {
			repositoryContextResult(w, r, nil, repositoryContextUnavailable())
			return
		}
		result, err := s.Search.SearchRepositoryCode(r.Context(), localCaller(r), q)
		repositoryContextResult(w, r, result, err)
	})
	mux.HandleFunc("POST "+prefix+"/impact", func(w http.ResponseWriter, r *http.Request) {
		q, err := readQuery(w, r)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		if s.Impact == nil {
			repositoryContextResult(w, r, nil, repositoryContextUnavailable())
			return
		}
		result, err := s.Impact.QueryRepositoryImpact(r.Context(), localCaller(r), q)
		repositoryContextResult(w, r, result, err)
	})
	mux.HandleFunc("POST "+prefix+"/index", func(w http.ResponseWriter, r *http.Request) {
		if err := repositoryContextHuman(r); err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		body, err := decodeAttention[repositoryContextIndexCommand](w, r)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		q, err := repositoryContextQuery(r, body.repositoryContextCommand)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		if s.Indexes == nil {
			repositoryContextResult(w, r, nil, repositoryContextUnavailable())
			return
		}
		result, err := s.Indexes.UpdateRepositoryIndex(r.Context(), localCaller(r), q, body.Action)
		repositoryContextResult(w, r, result, err)
	})
	mux.HandleFunc("POST "+prefix+"/map/refresh", func(w http.ResponseWriter, r *http.Request) {
		if err := repositoryContextHuman(r); err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		q, err := readQuery(w, r)
		if err != nil {
			repositoryContextResult(w, r, nil, err)
			return
		}
		if s.MapCommands == nil {
			repositoryContextResult(w, r, nil, repositoryContextUnavailable())
			return
		}
		result, err := s.MapCommands.RefreshRepositoryMap(r.Context(), localCaller(r), q)
		repositoryContextResult(w, r, result, err)
	})
}
