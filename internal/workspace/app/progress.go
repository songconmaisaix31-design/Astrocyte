package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// ProjectProgressRepository persists a per-project progress record. Loading an
// absent project returns the zero value, not an error, so a cache GET never
// fabricates a status.
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
	mu        sync.Mutex
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

// checkModelConsent verifies the project's current external-model consent and
// that the CLI adapter exists. It is re-run before the model turn and again
// before publication so a mid-flight revocation cannot publish a stale result.
func (s *projectProgressService) checkModelConsent(ctx context.Context, p domain.LocalProject) error {
	if p.Settings.ExternalModelCLI == "" {
		return projectError(apierrors.ApprovalRequired, "project external model consent is not configured")
	}
	if _, err := s.registry.Adapter(p.Settings.ExternalModelCLI); err != nil {
		return err
	}
	return nil
}

func (s *projectProgressService) GetProjectProgress(ctx context.Context, c domain.Caller, projectID string) (domain.ProjectProgress, error) {
	if _, err := s.progressProject(ctx, c, projectID, "read_context"); err != nil {
		return domain.ProjectProgress{}, err
	}
	record, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return record, err
	}
	// An absent project yields the zero record; fill the already-authorized
	// project identity so the transport never returns an empty project_id.
	// Status/source stay empty, revision stays zero, and nothing is written or
	// modeled — this is identity fill, not a fabricated stage.
	if record.ProjectID == "" {
		record.SchemaVersion = 1
		record.ProjectID = projectID
	}
	return record, nil
}

func (s *projectProgressService) SetProjectProgress(ctx context.Context, c domain.Caller, projectID string, input domain.ProgressInput) (domain.ProjectProgress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := human(c); err != nil {
		return domain.ProjectProgress{}, err
	}
	if _, err := s.projects.LoadProject(ctx, projectID); err != nil {
		return domain.ProjectProgress{}, err
	}
	if !domain.ValidProgressStatus(input.Status) || !domain.ValidProgressPercent(input.Percent) || len(input.Summary) > 2000 {
		return domain.ProjectProgress{}, projectError(apierrors.ValidationFailed, "project status, summary or percent is invalid")
	}
	current, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	// A human stage write must not drop the durable operation receipts: clearing
	// Operations (or the pending marker) would let a retry of an already accepted
	// or in-flight inference re-charge the model. Receipts are preserved verbatim.
	record := domain.ProjectProgress{SchemaVersion: 1, ProjectID: projectID, Status: input.Status, Summary: input.Summary, Percent: input.Percent, Source: "human", ObservedAt: time.Now().UTC(), Revision: current.Revision + 1, Operations: current.Operations, PendingOperation: current.PendingOperation}
	if err := s.progress.SaveProjectProgress(ctx, record, current.Revision); err != nil {
		return domain.ProjectProgress{}, err
	}
	return record, nil
}

// progressPrompt frames the approved fixed files as untrusted data and asks for
// a single bounded JSON object. It never embeds executable authority.
func progressPrompt(files []domain.ContextFile) string {
	var b strings.Builder
	b.WriteString("Analyze the supplied approved project task/status documents and infer the current project stage. Respond with exactly one JSON object and nothing else, using this shape: {\"status\":\"<short label>\",\"percent\":<optional 0..100 integer>}. Base the status only on the supplied documents. Do not infer completion from activity timestamps or session headers. Omit \"percent\" when it cannot be derived.\nDOCUMENTS:\n")
	for _, f := range files {
		b.WriteString("FILE: ")
		b.WriteString(f.Path)
		b.WriteString("\n")
		b.WriteString(f.Text)
		b.WriteString("\n")
	}
	return b.String()
}

type inferredStatus struct {
	Status  string `json:"status"`
	Percent *int   `json:"percent"`
}

// settleReceipt transitions a receipt to its terminal status in place. It never
// touches the persisted record; callers persist separately so a save failure can
// be distinguished from the semantic outcome. A nil result clears any previously
// stored snapshot so a terminal failure never leaves a computed result behind.
func (s *projectProgressService) settleReceipt(current *domain.ProjectProgress, op, status string, result *domain.ProgressResult) {
	if current.Operations == nil {
		current.Operations = map[string]domain.ProgressOperation{}
	}
	receipt := current.Operations[op]
	receipt.Status = status
	receipt.Result = result
	now := time.Now().UTC()
	receipt.FinishedAt = &now
	current.Operations[op] = receipt
	current.PendingOperation = ""
}

// failSettle durably marks a rejected or unknown inference. The pending receipt
// was already persisted before the paid turn, so this refinement is best-effort:
// the original typed error is always the return, never a fabricated success. If
// the receipt save also fails, the stale "pending" receipt still blocks a replay.
func (s *projectProgressService) failSettle(ctx context.Context, current domain.ProjectProgress, op, status string, semantic error) (domain.ProjectProgress, error) {
	s.settleReceipt(&current, op, status, nil)
	expected := current.Revision
	current.Revision = expected + 1
	if err := s.progress.SaveProjectProgress(ctx, current, expected); err != nil {
		return current, semantic
	}
	return current, semantic
}

// persistComputed durably stores the computed result inside a "computed" receipt
// before publication. This is the crash-safe step: if the process dies or the
// publication save keeps conflicting, a later retry finds the computed snapshot
// and re-publishes it without running another model turn. The pending-operation
// marker stays in place because publication is still outstanding.
func (s *projectProgressService) persistComputed(ctx context.Context, projectID, op string, result domain.ProgressResult) (domain.ProjectProgress, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		current, err := s.progress.LoadProjectProgress(ctx, projectID)
		if err != nil {
			return current, err
		}
		current.SchemaVersion = 1
		current.ProjectID = projectID
		if current.Operations == nil {
			current.Operations = map[string]domain.ProgressOperation{}
		}
		receipt := current.Operations[op]
		receipt.Status = "computed"
		receipt.Result = &result
		current.Operations[op] = receipt
		expected := current.Revision
		current.Revision = expected + 1
		if err := s.progress.SaveProjectProgress(ctx, current, expected); err != nil {
			var conflict *apierrors.ServiceError
			if errors.As(err, &conflict) && conflict.Code == apierrors.VersionConflict {
				lastErr = err
				continue
			}
			return current, err
		}
		return current, nil
	}
	return domain.ProjectProgress{}, lastErr
}

// publishAccepted re-verifies model consent and the permission revision the
// result was computed under, then durably publishes the computed result and its
// accepted receipt. On a CAS conflict the already-durable computed snapshot is
// re-published; the model turn is never re-run.
func (s *projectProgressService) publishAccepted(ctx context.Context, projectID, op string, result domain.ProgressResult) (domain.ProjectProgress, error) {
	before, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return s.unknownComputed(ctx, projectID, op, projectError(apierrors.DeliveryUnknown, "project could not be re-verified before publication"))
	}
	if before.Settings.Revision != result.SettingsRevision {
		return s.unknownComputed(ctx, projectID, op, projectError(apierrors.ApprovalRevoked, "project settings changed during inference"))
	}
	if err := s.checkModelConsent(ctx, before); err != nil {
		return s.unknownComputed(ctx, projectID, op, err)
	}
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		current, err := s.progress.LoadProjectProgress(ctx, projectID)
		if err != nil {
			return current, err
		}
		current.SchemaVersion = 1
		current.ProjectID = projectID
		if current.Operations == nil {
			current.Operations = map[string]domain.ProgressOperation{}
		}
		s.settleReceipt(&current, op, "accepted", &result)
		current.Status = result.Status
		current.Summary = result.Summary
		current.Percent = result.Percent
		current.Source = "agent_inferred"
		current.Evidence = result.Evidence
		current.NativeID = result.NativeID
		current.Model = result.Model
		current.ObservedAt = result.ObservedAt
		expected := current.Revision
		current.Revision = expected + 1
		if err := s.progress.SaveProjectProgress(ctx, current, expected); err != nil {
			var conflict *apierrors.ServiceError
			if errors.As(err, &conflict) && conflict.Code == apierrors.VersionConflict {
				lastErr = err
				continue
			}
			return current, err
		}
		return current, nil
	}
	return domain.ProjectProgress{}, lastErr
}

// unknownComputed discards a durably-computed result whose consent was revoked
// before publication. The receipt flips to "unknown" (result cleared) so a retry
// neither re-publishes nor re-runs the model.
func (s *projectProgressService) unknownComputed(ctx context.Context, projectID, op string, semantic error) (domain.ProjectProgress, error) {
	current, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return current, semantic
	}
	return s.failSettle(ctx, current, op, "unknown", semantic)
}

// restoreAccepted rebuilds the advisory record from an accepted receipt's
// immutable snapshot so a retry returns the original result, not a later human
// write that replaced the live fields.
func (s *projectProgressService) restoreAccepted(current domain.ProjectProgress, r domain.ProgressResult) domain.ProjectProgress {
	current.Status = r.Status
	current.Summary = r.Summary
	current.Percent = r.Percent
	current.Source = "agent_inferred"
	current.Evidence = r.Evidence
	current.NativeID = r.NativeID
	current.Model = r.Model
	current.ObservedAt = r.ObservedAt
	return current
}

func (s *projectProgressService) InferProjectProgress(ctx context.Context, c domain.Caller, projectID string, cmd domain.ProgressCommand) (domain.ProjectProgress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := human(c); err != nil {
		return domain.ProjectProgress{}, err
	}
	// The operation identity arrives from the transport (Idempotency-Key header),
	// never from the body, matching the native command identity discipline.
	op := c.OperationID
	if _, err := uuid.Parse(op); err != nil {
		return domain.ProjectProgress{}, projectError(apierrors.ValidationFailed, "a verified UUID operation identity is required")
	}
	p, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	if err := s.checkModelConsent(ctx, p); err != nil {
		return domain.ProjectProgress{}, err
	}
	if s.processor == nil || s.files == nil {
		return domain.ProjectProgress{}, projectError(apierrors.UnsupportedCapability, "progress inference is not connected")
	}
	if len(cmd.Files) == 0 || len(cmd.Files) > 8 {
		return domain.ProjectProgress{}, projectError(apierrors.ValidationFailed, "progress inference requires 1 to 8 fixed project files")
	}
	current, err := s.progress.LoadProjectProgress(ctx, projectID)
	if err != nil {
		return domain.ProjectProgress{}, err
	}
	if receipt, ok := current.Operations[op]; ok {
		switch receipt.Status {
		case "accepted":
			// Return the immutable accepted snapshot, not the live advisory fields
			// a later human write may have replaced.
			if receipt.Result != nil {
				return s.restoreAccepted(current, *receipt.Result), nil
			}
			return current, nil
		case "computed":
			// The result was durably computed but publication did not complete.
			// Re-publish the stored snapshot; no new model turn is run.
			if receipt.Result == nil {
				return current, projectError(apierrors.DeliveryUnknown, "computed progress receipt is missing its durable result")
			}
			return s.publishAccepted(ctx, projectID, op, *receipt.Result)
		case "failed":
			return current, projectError(apierrors.VersionConflict, "original progress inference failed; use a new operation identity")
		default:
			return current, projectError(apierrors.DeliveryUnknown, "progress inference remains pending or unknown; do not replay")
		}
	}
	// Persist the pending receipt before any paid model turn. The record must
	// carry its project identity before the first save so it is stored under the
	// correct key.
	current.SchemaVersion = 1
	current.ProjectID = projectID
	current.PendingOperation = op
	if current.Operations == nil {
		current.Operations = map[string]domain.ProgressOperation{}
	}
	current.Operations[op] = domain.ProgressOperation{Action: "infer", Status: "pending", CreatedAt: time.Now().UTC()}
	expected := current.Revision
	current.Revision = expected + 1
	if err := s.progress.SaveProjectProgress(ctx, current, expected); err != nil {
		return current, err
	}
	// Re-check consent immediately before the paid turn (permission revision).
	before, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return s.failSettle(ctx, current, op, "unknown", projectError(apierrors.DeliveryUnknown, "project could not be re-verified before inference"))
	}
	if before.Settings.Revision != p.Settings.Revision {
		return s.failSettle(ctx, current, op, "unknown", projectError(apierrors.ApprovalRevoked, "project settings changed during inference"))
	}
	if err := s.checkModelConsent(ctx, before); err != nil {
		return s.failSettle(ctx, current, op, "unknown", err)
	}
	files, err := s.files.ReadProjectFiles(ctx, before, cmd.Files)
	if err != nil {
		return s.failSettle(ctx, current, op, "failed", err)
	}
	result, err := s.processor.ProcessSelectedText(ctx, domain.TextRequest{CLI: before.Settings.ExternalModelCLI, Prompt: progressPrompt(files)})
	if err != nil {
		// A definite prelaunch rejection is "failed"; an unknown native delivery
		// is "unknown" and must never be automatically replayed.
		status := "unknown"
		var service *apierrors.ServiceError
		if errors.As(err, &service) && (service.Code == apierrors.ValidationFailed || service.Code == apierrors.UnsupportedCapability || service.Code == apierrors.ScopeDenied || service.Code == apierrors.ApprovalRevoked) {
			status = "failed"
		}
		return s.failSettle(ctx, current, op, status, err)
	}
	var inferred inferredStatus
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Text)), &inferred); err != nil || !domain.ValidProgressStatus(inferred.Status) || !domain.ValidProgressPercent(inferred.Percent) {
		return s.failSettle(ctx, current, op, "failed", projectError(apierrors.EvidenceMissing, "native progress output is not a valid bounded status"))
	}
	evidence := make([]domain.ProgressEvidence, 0, len(files))
	for _, f := range files {
		kind := "project_file"
		if f.Path == "TASK.md" || strings.HasSuffix(f.Path, "/TASK.md") {
			kind = "task"
		} else if f.Path == "STATUS.md" || strings.HasSuffix(f.Path, "/STATUS.md") {
			kind = "status"
		}
		evidence = append(evidence, domain.ProgressEvidence{SourcePath: f.Path, Kind: kind, Version: f.Version})
	}
	snapshot := domain.ProgressResult{Status: inferred.Status, Percent: inferred.Percent, Evidence: evidence, NativeID: result.NativeID, Model: result.Model, ObservedAt: time.Now().UTC(), SettingsRevision: before.Settings.Revision}
	// Durably persist the computed result before publication so a crash or a
	// publication conflict can be recovered without re-running the model turn.
	if _, err := s.persistComputed(ctx, projectID, op, snapshot); err != nil {
		return domain.ProjectProgress{}, err
	}
	return s.publishAccepted(ctx, projectID, op, snapshot)
}
