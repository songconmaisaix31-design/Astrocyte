package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// ProjectProgressRepository persists a per-project progress record. Loading an
// absent project returns the zero value, not an error, so a cache GET never
// fabricates a stage.
type ProjectProgressRepository interface {
	LoadProjectProgress(context.Context, string) (domain.ProjectProgress, error)
	SaveProjectProgress(context.Context, domain.ProjectProgress, int) error
}

// ProjectProgressService exposes project stage reading and writing. Reading is a
// durable-cache GET and never starts a native Agent or a model turn. Inference
// and human writes are separate human-initiated operations.
type ProjectProgressService interface {
	GetProjectProgress(context.Context, domain.Caller, string) (domain.ProjectProgress, error)
	SetProjectProgress(context.Context, domain.Caller, string, domain.ProgressInput) (domain.ProjectProgress, error)
	InferProjectProgress(context.Context, domain.Caller, string, domain.ProgressCommand) (domain.ProjectProgress, error)
}

type projectProgressService struct {
	projects  LocalProjectRepository
	progress  ProjectProgressRepository
	files     ProjectFiles
	processor TextProcessor
	registry  NativeRegistry
}

func NewProjectProgressService(projects LocalProjectRepository, progress ProjectProgressRepository, files ProjectFiles, processor TextProcessor, registry NativeRegistry) *projectProgressService {
	return &projectProgressService{projects: projects, progress: progress, files: files, processor: processor, registry: registry}
}

// progressProject resolves a project for a human read/write, or for an Agent
// with an active read grant. It never grants start/send/stop or model consent.
func (s *projectProgressService) progressProject(ctx context.Context, c domain.Caller, projectID, action string) (domain.LocalProject, error) {
	p, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return p, err
	}
	if c.Kind == "human" && c.ID != "" {
		return p, nil
	}
	if c.Kind != "agent" || c.ID == "" || c.ProjectID != projectID {
		return p, projectError(apierrors.ScopeDenied, "project progress is outside the caller scope")
	}
	g, err := s.projects.LoadGrant(ctx, projectID, c.ID)
	if err != nil || g.RevokedAt != nil || (g.ExpiresAt != nil && !g.ExpiresAt.After(time.Now())) {
		return p, projectError(apierrors.ApprovalRevoked, "project Agent grant is absent, expired or revoked")
	}
	if !has(g.Actions, action) {
		return p, projectError(apierrors.ScopeDenied, "Agent grant does not allow this action")
	}
	return p, nil
}

func (s *projectProgressService) GetProjectProgress(ctx context.Context, c domain.Caller, projectID string) (domain.ProjectProgress, error) {
	if _, err := s.progressProject(ctx, c, projectID, "read_context"); err != nil {
		return domain.ProjectProgress{}, err
	}
	return s.progress.LoadProjectProgress(ctx, projectID)
}

func (s *projectProgressService) SetProjectProgress(ctx context.Context, c domain.Caller, projectID string, input domain.ProgressInput) (domain.ProjectProgress, error) {
	if err := human(c); err != nil {
		return domain.ProjectProgress{}, err
	}
	if _, err := s.projects.LoadProject(ctx, projectID); err != nil {
		return domain.ProjectProgress{}, err
	}
	if !domain.ValidProgressStage(input.Stage) || !domain.ValidProgressPercent(input.Percent) {
		return domain.ProjectProgress{}, projectError(apierrors.ValidationFailed, "project stage or percent is invalid")
	}
	current, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	record := domain.ProjectProgress{ProjectID: projectID, Stage: input.Stage, Percent: input.Percent, Source: "human", ObservedAt: time.Now().UTC(), Revision: current.Revision + 1}
	if err := s.progress.SaveProjectProgress(ctx, record, current.Revision); err != nil {
		return domain.ProjectProgress{}, err
	}
	return record, nil
}

// progressPrompt frames the approved fixed files as untrusted data and asks for
// a single bounded JSON object. It never embeds executable authority.
func progressPrompt(files []domain.ContextFile) string {
	var b strings.Builder
	b.WriteString("Analyze the supplied approved project task/status documents and infer the current project stage. Respond with exactly one JSON object and nothing else, using this shape: {\"stage\":\"<short label>\",\"percent\":<optional 0..100 integer>}. Base the stage only on the supplied documents. Do not infer completion from activity timestamps or session headers. Omit \"percent\" when it cannot be derived.\nDOCUMENTS:\n")
	for _, f := range files {
		b.WriteString("FILE: ")
		b.WriteString(f.Path)
		b.WriteString("\n")
		b.WriteString(f.Text)
		b.WriteString("\n")
	}
	return b.String()
}

type inferredStage struct {
	Stage   string `json:"stage"`
	Percent *int   `json:"percent"`
}

func (s *projectProgressService) InferProjectProgress(ctx context.Context, c domain.Caller, projectID string, cmd domain.ProgressCommand) (domain.ProjectProgress, error) {
	if err := human(c); err != nil {
		return domain.ProjectProgress{}, err
	}
	p, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	if p.Settings.ExternalModelCLI == "" {
		return domain.ProjectProgress{}, projectError(apierrors.ApprovalRequired, "project external model consent is not configured")
	}
	if _, err := s.registry.Adapter(p.Settings.ExternalModelCLI); err != nil {
		return domain.ProjectProgress{}, err
	}
	if s.processor == nil || s.files == nil {
		return domain.ProjectProgress{}, projectError(apierrors.UnsupportedCapability, "progress inference is not connected")
	}
	if len(cmd.Files) == 0 || len(cmd.Files) > 8 {
		return domain.ProjectProgress{}, projectError(apierrors.ValidationFailed, "progress inference requires 1 to 8 fixed project files")
	}
	files, err := s.files.ReadProjectFiles(ctx, p, cmd.Files)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	prompt := progressPrompt(files)
	result, err := s.processor.ProcessSelectedText(ctx, domain.TextRequest{CLI: p.Settings.ExternalModelCLI, Prompt: prompt})
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	var inferred inferredStage
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Text)), &inferred); err != nil || !domain.ValidProgressStage(inferred.Stage) || !domain.ValidProgressPercent(inferred.Percent) {
		return domain.ProjectProgress{}, projectError(apierrors.EvidenceMissing, "native progress output is not a valid bounded stage")
	}
	basis := make([]domain.ProgressBasis, 0, len(files))
	for _, f := range files {
		basis = append(basis, domain.ProgressBasis{Path: f.Path, Version: f.Version})
	}
	current, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	record := domain.ProjectProgress{ProjectID: projectID, Stage: inferred.Stage, Percent: inferred.Percent, Source: "agent_inferred", Basis: basis, NativeID: result.NativeID, Model: result.Model, ObservedAt: time.Now().UTC(), Revision: current.Revision + 1}
	if err := s.progress.SaveProjectProgress(ctx, record, current.Revision); err != nil {
		return domain.ProjectProgress{}, err
	}
	return record, nil
}
