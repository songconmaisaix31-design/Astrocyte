package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type OpenVikingService struct {
	attention *Service
	repo      OpenVikingRepository
	gateway   OpenVikingGateway
}

func NewOpenVikingService(ctx context.Context, attention *Service, repo OpenVikingRepository, gateway OpenVikingGateway) (*OpenVikingService, error) {
	if attention == nil || repo == nil {
		return nil, fmt.Errorf("OpenViking requires the real Attention service and repository")
	}
	if err := repo.RecoverOpenViking(ctx); err != nil {
		return nil, err
	}
	return &OpenVikingService{attention: attention, repo: repo, gateway: gateway}, nil
}

func (s *OpenVikingService) Status(ctx context.Context, p Principal) (OpenVikingStatus, error) {
	if err := authorize(p, false); err != nil {
		return OpenVikingStatus{}, err
	}
	if s.gateway == nil {
		return OpenVikingStatus{SchemaVersion: 1, Message: "OpenViking 未配置；本地资料与核心页面正常可用"}, nil
	}
	return s.gateway.Status(ctx)
}

var openVikingIdentifier = regexp.MustCompile(`^[A-Za-z0-9_-]{1,160}$`)

func (s *OpenVikingService) scope(ctx context.Context, p Principal, spaceID string) (OpenVikingStatus, string, error) {
	if err := authorize(p, true); err != nil {
		return OpenVikingStatus{}, "", err
	}
	if s.gateway == nil {
		return OpenVikingStatus{}, "", serviceError(apierrors.ProviderUnavailable, "OpenViking 未配置", "configure_openviking_on_server")
	}
	if !openVikingIdentifier.MatchString(spaceID) {
		return OpenVikingStatus{}, "", serviceError(apierrors.ValidationFailed, "项目空间标识无效", "select_existing_project_space")
	}
	if _, err := s.attention.GetProjectSpace(ctx, p, spaceID); err != nil {
		return OpenVikingStatus{}, "", err
	}
	status, err := s.gateway.Status(ctx)
	if err != nil {
		return status, "", err
	}
	if !status.Configured || !status.Connected || status.User == "" || status.Account == "" {
		return status, "", serviceError(apierrors.ProviderUnavailable, "OpenViking 健康与身份尚未核实", "verify_openviking_identity")
	}
	root := status.RootURI + "/projects/" + spaceID
	root, err = domain.CanonicalVikingURI(root)
	if err != nil {
		return status, "", serviceError(apierrors.ValidationFailed, "OpenViking 项目根无效", "correct_server_root")
	}
	return status, root, nil
}

func (s *OpenVikingService) List(ctx context.Context, p Principal, spaceID string) ([]domain.OpenVikingSubmission, error) {
	if err := authorize(p, false); err != nil {
		return nil, err
	}
	if _, err := s.attention.GetProjectSpace(ctx, p, spaceID); err != nil {
		return nil, err
	}
	return s.repo.ListOpenViking(ctx, spaceID)
}

func (s *OpenVikingService) Submit(ctx context.Context, p Principal, spaceID string, c OpenVikingSubmitCommand) (domain.OpenVikingSubmission, error) {
	var empty domain.OpenVikingSubmission
	if err := authorize(p, true); err != nil {
		return empty, err
	}
	if c.SchemaVersion != 1 || c.ExpectedVersion == nil || *c.ExpectedVersion != 0 || c.ExpectedSpaceVersion < 1 || strings.TrimSpace(c.RequestID) == "" || c.RequestID != c.IdempotencyKey || c.MaterialID == "" || c.Revision < 1 {
		return empty, serviceError(apierrors.ValidationFailed, "提交身份和固定资料版本必须完整且一致", "review_fixed_revision")
	}
	status, root, err := s.scope(ctx, p, spaceID)
	if err != nil {
		return empty, err
	}
	if c.Destination != root || c.BaseURL != status.BaseURL || c.Account != status.Account || c.User != status.User {
		return empty, serviceError(apierrors.ContextStale, "OpenViking 目的地与当前配置不同", "review_current_destination")
	}
	space, err := s.attention.GetProjectSpace(ctx, p, spaceID)
	if err != nil {
		return empty, err
	}
	if space.Space.Version != c.ExpectedSpaceVersion {
		return empty, serviceError(apierrors.ContextStale, "项目引用版本已变化", "reload_project_references")
	}
	ref := SourceRef{MaterialID: c.MaterialID, Revision: c.Revision}
	snapshot, err := s.attention.ReadProjectReference(ctx, p, spaceID, ref, false)
	if err != nil {
		return empty, err
	}
	if strings.TrimSpace(snapshot.Text) == "" || len(snapshot.Text) > 10<<20 {
		return empty, serviceError(apierrors.ValidationFailed, "资料正文为空或超过 10 MiB", "select_supported_fixed_text")
	}
	now := time.Now().UTC()
	row := domain.OpenVikingSubmission{ID: rand.Text(), Version: 1, RequestID: c.RequestID, SpaceID: spaceID, SpaceVersion: c.ExpectedSpaceVersion, MaterialID: c.MaterialID, Revision: c.Revision, Locator: snapshot.Ref.Locator, Title: snapshot.Title, Destination: root, BaseURL: status.BaseURL, Account: status.Account, User: status.User, ActorID: p.ID, ApprovedAt: now, UpdatedAt: now, Status: "submitting", Stage: "upload"}
	row, created, err := s.repo.ReserveOpenViking(ctx, row)
	if err != nil {
		return row, err
	}
	if !created {
		if row.Status == "submitting" || row.Status == "unknown" {
			return row, serviceError(apierrors.DeliveryUnknown, "原提交仍在处理或结果未知，请核实原记录", "inspect_original_submission")
		}
		return row, nil
	}
	if err := ctx.Err(); err != nil {
		return s.fail(row, serviceError(apierrors.ProviderUnavailable, "请求在上传前结束，未发出上传", "review_submission"))
	}
	if err := s.currentReference(ctx, ref, spaceID); err != nil {
		return s.fail(row, err)
	}
	tempID, err := s.gateway.Upload(ctx, snapshot.Title, snapshot.Text)
	if err != nil {
		return s.fail(row, err)
	}
	row.TempFileID = tempID
	row.Stage = "uploaded"
	if err := s.save(&row); err != nil {
		return row, serviceError(apierrors.DeliveryUnknown, "临时上传成功，保存结果失败；禁止重发", "inspect_original_submission")
	}
	current, currentRoot, err := s.scope(ctx, p, spaceID)
	if err != nil {
		return s.fail(row, err)
	}
	if current.BaseURL != row.BaseURL || currentRoot != row.Destination || current.Account != row.Account || current.User != row.User {
		return s.fail(row, serviceError(apierrors.ContextStale, "目的地已变化，资源未提交", "review_original_destination"))
	}
	if err := s.currentReference(ctx, ref, spaceID); err != nil {
		return s.fail(row, err)
	}
	if err := ctx.Err(); err != nil {
		return s.fail(row, serviceError(apierrors.ProviderUnavailable, "请求在资源提交前结束，未发出资源提交", "review_submission"))
	}
	row.Stage = "resource_submit"
	if err := s.save(&row); err != nil {
		return row, err
	}
	taskID, uri, err := s.gateway.Submit(ctx, tempID, root, snapshot.Title)
	if err != nil {
		return s.fail(row, err)
	}
	row.TaskID = taskID
	row.RootURI = uri
	row.Status = "submitted"
	row.Stage = "task"
	if err := s.save(&row); err != nil {
		return row, serviceError(apierrors.DeliveryUnknown, "资源已提交，保存任务标识失败；禁止重发", "inspect_original_submission")
	}
	return row, nil
}

func (s *OpenVikingService) currentReference(ctx context.Context, ref SourceRef, spaceID string) error {
	return mapError(s.attention.repo.WithTx(ctx, func(tx AttentionTx) error { return projectReferenceAllowed(tx, spaceID, ref, false) }), "")
}

func (s *OpenVikingService) save(row *domain.OpenVikingSubmission) error {
	expected := row.Version
	row.Version++
	row.UpdatedAt = time.Now().UTC()
	// 上游结果必须在 HTTP 客户端断开后仍能持久化。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.repo.SaveOpenViking(ctx, *row, expected)
}

func (s *OpenVikingService) fail(row domain.OpenVikingSubmission, cause error) (domain.OpenVikingSubmission, error) {
	row.Status = "failed"
	var known *apierrors.ServiceError
	if !errors.As(cause, &known) || known.Code == apierrors.DeliveryUnknown {
		row.Status = "unknown"
	}
	row.Error = cause.Error()
	if err := s.save(&row); err != nil {
		return row, serviceError(apierrors.DeliveryUnknown, "外部操作后记录保存失败；禁止重发", "inspect_original_submission")
	}
	return row, cause
}

func (s *OpenVikingService) Refresh(ctx context.Context, p Principal, id string) (domain.OpenVikingSubmission, *OpenVikingTask, error) {
	var empty domain.OpenVikingSubmission
	if err := authorize(p, true); err != nil {
		return empty, nil, err
	}
	row, err := s.repo.GetOpenViking(ctx, id)
	if err != nil {
		return row, nil, err
	}
	status, root, err := s.scope(ctx, p, row.SpaceID)
	if err != nil {
		return row, nil, err
	}
	if row.BaseURL != status.BaseURL || row.Destination != root || row.Account != status.Account || row.User != status.User {
		return row, nil, serviceError(apierrors.ContextStale, "提交使用的目的地与当前配置不同", "restore_original_destination")
	}
	if row.TaskID == "" {
		return row, nil, serviceError(apierrors.DeliveryUnknown, "原记录没有可查询的任务标识，禁止重发", "reconcile_original_operation")
	}
	if err := s.currentReference(ctx, SourceRef{MaterialID: row.MaterialID, Revision: row.Revision}, row.SpaceID); err != nil {
		return row, nil, err
	}
	task, err := s.gateway.Task(ctx, row.TaskID)
	if err != nil {
		return row, nil, err
	}
	if task.TaskID != row.TaskID || !slices.Contains([]string{"pending", "running", "cancelling", "completed", "failed", "cancelled"}, task.Status) {
		return row, nil, serviceError(apierrors.ProviderUnavailable, "上游任务身份或状态无效", "inspect_upstream_task")
	}
	boundURI, err := domain.BindVikingResource(root, row.RootURI, task.ResourceID)
	if err != nil {
		return row, nil, serviceError(apierrors.ScopeDenied, "上游任务资源与原提交的固定范围不一致", "inspect_original_submission_resource")
	}
	row.RootURI = boundURI
	if task.Status == "completed" && row.RootURI == "" {
		return row, nil, serviceError(apierrors.ProviderUnavailable, "已完成任务缺少资源 URI", "inspect_upstream_task")
	}
	row.Status = task.Status
	row.TaskSnapshot, err = json.Marshal(task)
	if err != nil {
		return row, nil, err
	}
	if task.Stage != nil {
		row.Stage = *task.Stage
	}
	if task.Error != nil {
		row.Error = *task.Error
	}
	if err := s.save(&row); err != nil {
		return row, nil, err
	}
	return row, &task, nil
}

func (s *OpenVikingService) permittedURI(ctx context.Context, spaceID, uri, root string, status OpenVikingStatus) error {
	if err := domain.VikingWithin(uri, root); err != nil {
		return serviceError(apierrors.ScopeDenied, "资源 URI 不在当前项目范围", "select_project_resource")
	}
	rows, err := s.repo.ListOpenViking(ctx, spaceID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.BaseURL != status.BaseURL || row.Account != status.Account || row.User != status.User || row.Destination != root || row.Status != "completed" || row.RootURI == "" {
			continue
		}
		bound, err := domain.BindVikingResource(root, row.RootURI, "")
		if err != nil {
			return serviceError(apierrors.ContextStale, "已保存资源范围无效，需要核实原提交", "inspect_original_submission_resource")
		}
		if err := domain.VikingWithin(uri, bound); err == nil {
			return s.currentReference(ctx, SourceRef{MaterialID: row.MaterialID, Revision: row.Revision}, spaceID)
		}
	}
	return serviceError(apierrors.ScopeDenied, "资源没有当前项目的已完成外发记录", "select_completed_project_submission")
}

func (s *OpenVikingService) Search(ctx context.Context, p Principal, spaceID string, c OpenVikingSearchCommand) ([]OpenVikingHit, error) {
	if err := authorize(p, true); err != nil {
		return nil, err
	}
	if strings.TrimSpace(c.Query) == "" || len(c.Query) > 8000 {
		return nil, serviceError(apierrors.ValidationFailed, "检索内容为空或过长", "provide_bounded_query")
	}
	status, root, err := s.scope(ctx, p, spaceID)
	if err != nil {
		return nil, err
	}
	hits, err := s.gateway.Find(ctx, root, c.Query)
	if err != nil {
		return nil, err
	}
	for _, hit := range hits {
		if hit.ContextType != "resource" || hit.Level < 0 || hit.Level > 2 || math.IsNaN(hit.Score) || math.IsInf(hit.Score, 0) {
			return nil, serviceError(apierrors.ProviderUnavailable, "上游检索结果字段无效", "inspect_upstream_results")
		}
		if err := s.permittedURI(ctx, spaceID, hit.URI, root, status); err != nil {
			return nil, err
		}
	}
	return hits, nil
}

func (s *OpenVikingService) Content(ctx context.Context, p Principal, spaceID string, c OpenVikingContentCommand) (string, error) {
	if err := authorize(p, true); err != nil {
		return "", err
	}
	if c.Level == nil || *c.Level < 0 || *c.Level > 2 {
		return "", serviceError(apierrors.ValidationFailed, "内容层级必须是 0、1 或 2", "choose_supported_level")
	}
	canonical, err := domain.CanonicalVikingURI(c.URI)
	if err != nil {
		return "", serviceError(apierrors.ValidationFailed, "内容 URI 无效", "select_project_resource")
	}
	c.URI = canonical
	level := *c.Level
	status, root, err := s.scope(ctx, p, spaceID)
	if err != nil {
		return "", err
	}
	if err := s.permittedURI(ctx, spaceID, c.URI, root, status); err != nil {
		return "", err
	}
	uri := c.URI
	if level < 2 && (strings.HasSuffix(uri, "/.abstract.md") || strings.HasSuffix(uri, "/.overview.md")) {
		uri = uri[:strings.LastIndex(uri, "/")]
	}
	content, err := s.gateway.Read(ctx, uri, level)
	if err != nil {
		return "", err
	}
	if err := s.permittedURI(ctx, spaceID, c.URI, root, status); err != nil {
		return "", err
	}
	return content, nil
}
