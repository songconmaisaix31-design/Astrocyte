package httpapi

import (
	"net/http"
	"time"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type progressSetCommand struct {
	attentionapp.CommandMeta
	domain.ProgressInput
}

type progressInferCommand struct {
	attentionapp.CommandMeta
	domain.ProgressCommand
}

// progressEvidenceWire 投影领域中的实际文件证据。
type progressEvidenceWire struct {
	SourcePath string `json:"source_path"`
	Kind       string `json:"kind"`
	Version    string `json:"version"`
	Excerpt    string `json:"excerpt,omitempty"`
}

// progressWire 只公开进度字段；内部操作回执保持私有，未观察时间保持空值。
type progressWire struct {
	ProjectID  string                 `json:"project_id"`
	Status     string                 `json:"status"`
	Summary    string                 `json:"summary,omitempty"`
	Percent    *int                   `json:"percent,omitempty"`
	Source     *string                `json:"source"`
	Evidence   []progressEvidenceWire `json:"evidence"`
	NativeID   string                 `json:"native_id,omitempty"`
	Model      *string                `json:"model,omitempty"`
	ObservedAt *string                `json:"observed_at"`
	Warning    string                 `json:"warning,omitempty"`
	Revision   int                    `json:"revision"`
}

// mapProgress 将领域记录映射为 HTTP 字段，未观察状态明确为 unknown。
func mapProgress(p domain.ProjectProgress) progressWire {
	w := progressWire{
		ProjectID: p.ProjectID,
		Status:    p.Status,
		Summary:   p.Summary,
		Percent:   p.Percent,
		NativeID:  p.NativeID,
		Model:     p.Model,
		Warning:   p.Warning,
		Revision:  p.Revision,
		Evidence:  []progressEvidenceWire{},
	}
	if w.Status == "" {
		w.Status = "unknown"
	}
	if p.Source != "" {
		src := p.Source
		w.Source = &src
	}
	if !p.ObservedAt.IsZero() {
		ts := p.ObservedAt.UTC().Format(time.RFC3339)
		w.ObservedAt = &ts
	}
	for _, e := range p.Evidence {
		w.Evidence = append(w.Evidence, progressEvidenceWire{SourcePath: e.SourcePath, Kind: e.Kind, Version: e.Version, Excerpt: e.Excerpt})
	}
	return w
}

// registerProgress 连接缓存读取、人工写入和人工发起的模型推断。
// 缺少服务时返回 501，权限和业务错误由领域服务检查后传递。
func (h *handler) registerProgress(mux *http.ServeMux) {
	s := h.services.Progress
	if s == nil {
		for _, route := range []string{
			"GET /api/v1/local-projects/{id}/progress",
			"PUT /api/v1/local-projects/{id}/progress",
			"POST /api/v1/local-projects/{id}/progress/infer",
		} {
			mux.HandleFunc(route, h.notImplemented("project_progress"))
		}
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects/{id}/progress", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.GetProjectProgress(r.Context(), localCaller(r), r.PathValue("id"))
		if err != nil {
			h.attentionResult(w, r, http.StatusOK, nil, err)
			return
		}
		h.attentionResult(w, r, http.StatusOK, localEnvelope("progress", mapProgress(p)), nil)
	})
	mux.HandleFunc("PUT /api/v1/local-projects/{id}/progress", commandHandler(h, func(c *progressSetCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c progressSetCommand) (any, error) {
		p, err := s.SetProjectProgress(r.Context(), localCaller(r), r.PathValue("id"), c.ProgressInput)
		if err != nil {
			return nil, err
		}
		return localEnvelope("progress", mapProgress(p)), nil
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/progress/infer", commandHandler(h, func(c *progressInferCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c progressInferCommand) (any, error) {
		p, err := s.InferProjectProgress(r.Context(), localCaller(r), r.PathValue("id"), c.ProgressCommand)
		if err != nil {
			return nil, err
		}
		return localEnvelope("progress", mapProgress(p)), nil
	}))
}
