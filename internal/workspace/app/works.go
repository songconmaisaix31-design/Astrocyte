package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// WorksRepository persists the replaceable album observation and the durable
// per-work human metadata. Observations are replaced by refresh/import; metadata
// survives them because it is keyed by work identity.
type WorksRepository interface {
	LoadWorksSnapshot(context.Context) (domain.WorksSnapshot, error)
	SaveWorksSnapshot(context.Context, domain.WorksSnapshot) error
	LoadWorkMetadata(context.Context, string) (domain.WorkHumanMetadata, error)
	SaveWorkMetadata(context.Context, string, domain.WorkHumanMetadata, int) error
}

// WorksScanSource observes only explicitly authorized roots through fixed
// programs. It never scans the whole disk, reads credential files or executes a
// project script. The returned works carry file facts actually read on disk.
type WorksScanSource interface {
	ScanWorks(context.Context, []string) ([]domain.Work, []domain.WorkDiscoveryFailure, error)
}

// WorksDeliverySource parses a bounded "AI 投递 JSON" envelope. It performs no
// filesystem access: the Path in each entry is a stored root for later
// verification, not something read here.
type WorksDeliverySource interface {
	ParseDelivery(context.Context, []byte) ([]domain.Work, error)
}

// WorksEventSink is the chain-1 boundary. Every works activity is emitted here
// and W1 normalizes it into the unified Event (§4.1).
type WorksEventSink interface {
	EmitWorkEvent(context.Context, domain.WorkEvent) error
}

// Works is the album use-case surface. Every mutation is human-only and never
// grants a project, Agent or execution permission.
type Works interface {
	ListWorks(context.Context, domain.Caller) (domain.WorksSnapshot, error)
	RefreshWorks(context.Context, domain.Caller, WorksRefreshCommand) (domain.WorksSnapshot, error)
	ImportWorksDelivery(context.Context, domain.Caller, WorksDeliveryCommand) (domain.WorksSnapshot, error)
	SetWorkAttitude(context.Context, domain.Caller, string, WorkAttitudeCommand) (domain.Work, error)
	SetWorkGroup(context.Context, domain.Caller, string, WorkGroupCommand) (domain.Work, error)
	SetWorkHandoff(context.Context, domain.Caller, string, WorkHandoffCommand) (domain.Work, error)
	SetWorkHidden(context.Context, domain.Caller, string, WorkHiddenCommand) (domain.Work, error)
}

type WorksRefreshCommand struct {
	Roots []string `json:"roots"`
}

type WorksDeliveryCommand struct {
	Content []byte `json:"-"`
}

type WorkAttitudeCommand struct {
	ExpectedRevision int    `json:"expected_revision"`
	Attitude         string `json:"attitude"`
}

type WorkGroupCommand struct {
	ExpectedRevision int    `json:"expected_revision"`
	Group            string `json:"group"`
}

type WorkHandoffCommand struct {
	ExpectedRevision int    `json:"expected_revision"`
	Content          string `json:"content"`
}

type WorkHiddenCommand struct {
	ExpectedRevision int  `json:"expected_revision"`
	Hidden           bool `json:"hidden"`
}

type WorksServiceOptions struct {
	RecentWindow time.Duration
	Clock        func() time.Time
}

type worksService struct {
	repo         WorksRepository
	scan         WorksScanSource
	delivery     WorksDeliverySource
	events       WorksEventSink
	recentWindow time.Duration
	clock        func() time.Time
}

var _ Works = (*worksService)(nil)

// NewWorksService assembles the album service. A nil event sink is allowed so a
// discovery-only assembly cannot imply chain-1 integration; activity is then
// simply not emitted rather than fabricated.
func NewWorksService(repo WorksRepository, scan WorksScanSource, delivery WorksDeliverySource, events WorksEventSink, options WorksServiceOptions) *worksService {
	if options.RecentWindow <= 0 {
		options.RecentWindow = 30 * 24 * time.Hour
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &worksService{repo: repo, scan: scan, delivery: delivery, events: events, recentWindow: options.RecentWindow, clock: options.Clock}
}

func workError(code apierrors.Code, message string) error {
	action := "review_works_album_input"
	if code == apierrors.DeliveryUnknown {
		action = "reload_works_without_replay"
	}
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: action}
}

// ListWorks is a pure cache read: it never scans or emits. Statistics are
// recomputed in memory so an upgrade needs no disk pass.
func (s *worksService) ListWorks(ctx context.Context, c domain.Caller) (domain.WorksSnapshot, error) {
	if err := human(c); err != nil {
		return domain.WorksSnapshot{}, err
	}
	snapshot, err := s.repo.LoadWorksSnapshot(ctx)
	if err != nil {
		return snapshot, err
	}
	return s.withMetadata(ctx, snapshot)
}

func (s *worksService) RefreshWorks(ctx context.Context, c domain.Caller, cmd WorksRefreshCommand) (domain.WorksSnapshot, error) {
	return worksRequest(ctx, s, c, "RefreshWorks", cmd, func() (domain.WorksSnapshot, error) { return s.refreshWorks(ctx, c, cmd) })
}

func (s *worksService) refreshWorks(ctx context.Context, c domain.Caller, cmd WorksRefreshCommand) (domain.WorksSnapshot, error) {
	if err := human(c); err != nil {
		return domain.WorksSnapshot{}, err
	}
	if s.scan == nil {
		return domain.WorksSnapshot{}, workError(apierrors.UnsupportedCapability, "works scan source is not connected")
	}
	if len(cmd.Roots) > 64 {
		return domain.WorksSnapshot{}, workError(apierrors.ValidationFailed, "too many authorized roots")
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	old, err := s.repo.LoadWorksSnapshot(ctx)
	if err != nil {
		return old, err
	}
	works, failures, err := s.scan.ScanWorks(ctx, cmd.Roots)
	if err != nil {
		if old.ObservedAt != nil {
			old.Status = "stale"
		} else {
			old.Status = "unknown"
		}
		if len(failures) == 0 {
			old.Failures = append(old.Failures, domain.WorkDiscoveryFailure{Root: "", Reason: "works_scan_source_unavailable"})
		}
		if ctx.Err() != nil {
			return old, ctx.Err()
		}
		if saveErr := s.repo.SaveWorksSnapshot(ctx, old); saveErr != nil {
			return old, saveErr
		}
		return s.withMetadata(ctx, old)
	}
	works = retainAttributedWorks(old.Works, works)
	snapshot := s.mergeObservations(old, works, failures)
	if err := s.repo.SaveWorksSnapshot(ctx, snapshot); err != nil {
		return snapshot, err
	}
	if err := s.emit(ctx, domain.WorksSnapshot{Works: works}, domain.WorkEventSynced); err != nil {
		return snapshot, err
	}
	return s.withMetadata(ctx, snapshot)
}

func (s *worksService) ImportWorksDelivery(ctx context.Context, c domain.Caller, cmd WorksDeliveryCommand) (domain.WorksSnapshot, error) {
	return worksRequest(ctx, s, c, "ImportWorksDelivery", string(cmd.Content), func() (domain.WorksSnapshot, error) { return s.importWorksDelivery(ctx, c, cmd) })
}

func (s *worksService) importWorksDelivery(ctx context.Context, c domain.Caller, cmd WorksDeliveryCommand) (domain.WorksSnapshot, error) {
	if err := human(c); err != nil {
		return domain.WorksSnapshot{}, err
	}
	if s.delivery == nil {
		return domain.WorksSnapshot{}, workError(apierrors.UnsupportedCapability, "works delivery source is not connected")
	}
	if len(cmd.Content) == 0 || len(cmd.Content) > 1<<20 {
		return domain.WorksSnapshot{}, workError(apierrors.ValidationFailed, "delivery JSON is empty or over 1 MiB")
	}
	works, err := s.delivery.ParseDelivery(ctx, cmd.Content)
	if err != nil {
		return domain.WorksSnapshot{}, workError(apierrors.ValidationFailed, "delivery JSON could not be parsed: "+err.Error())
	}
	old, err := s.repo.LoadWorksSnapshot(ctx)
	if err != nil {
		return old, err
	}
	snapshot := s.mergeObservations(old, works, nil)
	if err := s.repo.SaveWorksSnapshot(ctx, snapshot); err != nil {
		return snapshot, err
	}
	if err := s.seedDeliveryMetadata(ctx, works); err != nil {
		return snapshot, workError(apierrors.DeliveryUnknown, "works observation saved but delivery metadata outcome is unknown; reload without replay")
	}
	if err := s.emit(ctx, domain.WorksSnapshot{Works: works}, domain.WorkEventImported); err != nil {
		return snapshot, err
	}
	return s.withMetadata(ctx, snapshot)
}

// seedDeliveryMetadata captures AI-delivered hints (attitude, group, handoff) as
// initial metadata only when no human metadata exists yet. A human override is
// never overwritten, and a delivered handoff is marked generated.
func (s *worksService) seedDeliveryMetadata(ctx context.Context, works []domain.Work) error {
	for _, work := range works {
		if work.Attitude == "" && work.Group == "" && work.Handoff.Content == "" {
			continue
		}
		existing, err := s.repo.LoadWorkMetadata(ctx, work.ID)
		if err != nil {
			return err
		}
		if existing.Revision > 0 {
			continue
		}
		stamp := s.clock().UTC()
		metadata := domain.WorkHumanMetadata{Attitude: work.Attitude, Group: work.Group, UpdatedAt: &stamp, Revision: 1}
		if work.Handoff.Content != "" {
			metadata.Handoff = domain.WorkHandoff{Content: work.Handoff.Content, UpdatedAt: &stamp, Source: domain.HandoffSourceGenerated}
		}
		if err := s.repo.SaveWorkMetadata(ctx, work.ID, metadata, 0); err != nil {
			return err
		}
	}
	return nil
}

// mergeObservations replaces observations for the same identity while keeping
// the album's existing entries for works the scan/delivery did not touch. It
// never touches human metadata, which lives in its own table.
func (s *worksService) mergeObservations(old domain.WorksSnapshot, works []domain.Work, failures []domain.WorkDiscoveryFailure) domain.WorksSnapshot {
	byID := map[string]domain.Work{}
	order := []string{}
	for _, work := range old.Works {
		if _, ok := byID[work.ID]; !ok {
			order = append(order, work.ID)
		}
		byID[work.ID] = work
	}
	for id, work := range byID {
		for _, failure := range failures {
			left, right := filepath.Clean(work.Root), filepath.Clean(failure.Root)
			same := left == right
			if runtime.GOOS == "windows" {
				same = strings.EqualFold(left, right)
			}
			if same {
				work.Files.Verified = false
				work.Files.Reason = failure.Reason
				byID[id] = work
			}
		}
	}
	for _, work := range works {
		if _, ok := byID[work.ID]; !ok {
			order = append(order, work.ID)
		}
		byID[work.ID] = work
	}
	merged := make([]domain.Work, 0, len(order))
	for _, id := range order {
		merged = append(merged, byID[id])
	}
	now := s.clock().UTC()
	return domain.WorksSnapshot{
		Works:      merged,
		Status:     statusForFailures(failures),
		ObservedAt: &now,
		Failures:   append([]domain.WorkDiscoveryFailure{}, failures...),
	}
}

func statusForFailures(failures []domain.WorkDiscoveryFailure) string {
	if len(failures) == 0 {
		return "complete"
	}
	return "partial"
}

// A scan with no platform marker verifies existing attributed cards at this
// root. It does not invent a second work or overwrite delivery attribution.
func retainAttributedWorks(cached, scanned []domain.Work) []domain.Work {
	result := []domain.Work{}
	for _, observation := range scanned {
		matched := false
		if observation.Platform == "unknown" {
			for _, work := range cached {
				left, right := filepath.Clean(work.Root), filepath.Clean(observation.Root)
				same := left == right
				if runtime.GOOS == "windows" {
					same = strings.EqualFold(left, right)
				}
				if !same {
					continue
				}
				work.Files = observation.Files
				work.Kind = observation.Kind
				result = append(result, work)
				matched = true
			}
		}
		if !matched {
			result = append(result, observation)
		}
	}
	return result
}

// withMetadata overlays durable human annotations and recomputes statistics.
func (s *worksService) withMetadata(ctx context.Context, snapshot domain.WorksSnapshot) (domain.WorksSnapshot, error) {
	for i := range snapshot.Works {
		metadata, err := s.repo.LoadWorkMetadata(ctx, snapshot.Works[i].ID)
		if err != nil {
			return snapshot, err
		}
		metadata.Apply(&snapshot.Works[i])
	}
	snapshot.Platforms = domain.AggregateWorksStats(snapshot.Works, s.clock(), s.recentWindow)
	return snapshot, nil
}

func (s *worksService) loadWork(ctx context.Context, id string) (domain.Work, error) {
	snapshot, err := s.repo.LoadWorksSnapshot(ctx)
	if err != nil {
		return domain.Work{}, err
	}
	for _, work := range snapshot.Works {
		if work.ID != id {
			continue
		}
		metadata, err := s.repo.LoadWorkMetadata(ctx, id)
		if err != nil {
			return domain.Work{}, err
		}
		metadata.Apply(&work)
		return work, nil
	}
	return domain.Work{}, workError(apierrors.NotFound, "work is absent from the album")
}

func (s *worksService) saveMetadata(ctx context.Context, c domain.Caller, id string, expectedRevision int, mutate func(domain.WorkHumanMetadata) domain.WorkHumanMetadata, eventKind string) (domain.Work, error) {
	work, err := s.loadWork(ctx, id)
	if err != nil {
		return domain.Work{}, err
	}
	if work.Revision != expectedRevision {
		return domain.Work{}, workError(apierrors.VersionConflict, "work metadata changed; reload before saving")
	}
	metadata := domain.WorkHumanMetadata{
		Attitude: work.Attitude, Group: work.Group, Hidden: work.Hidden,
		Handoff: work.Handoff, UpdatedAt: work.UpdatedAt, Revision: work.Revision,
	}
	next := mutate(metadata)
	next.Revision = work.Revision + 1
	stamp := s.clock().UTC()
	next.UpdatedAt = &stamp
	if err := s.repo.SaveWorkMetadata(ctx, id, next, work.Revision); err != nil {
		return domain.Work{}, err
	}
	next.Apply(&work)
	if err := s.emitWork(ctx, work, eventKind); err != nil {
		return work, err
	}
	return work, nil
}

func (s *worksService) SetWorkAttitude(ctx context.Context, c domain.Caller, id string, cmd WorkAttitudeCommand) (domain.Work, error) {
	return worksRequest(ctx, s, c, "SetWorkAttitude", struct {
		ID      string
		Command WorkAttitudeCommand
	}{id, cmd}, func() (domain.Work, error) { return s.setWorkAttitude(ctx, c, id, cmd) })
}

func (s *worksService) setWorkAttitude(ctx context.Context, c domain.Caller, id string, cmd WorkAttitudeCommand) (domain.Work, error) {
	if err := human(c); err != nil {
		return domain.Work{}, err
	}
	if !domain.ValidAttitude(cmd.Attitude) {
		return domain.Work{}, workError(apierrors.ValidationFailed, "attitude must be continue, park or done")
	}
	return s.saveMetadata(ctx, c, id, cmd.ExpectedRevision, func(m domain.WorkHumanMetadata) domain.WorkHumanMetadata {
		m.Attitude = cmd.Attitude
		return m
	}, domain.WorkEventAttitudeSet)
}

func (s *worksService) SetWorkGroup(ctx context.Context, c domain.Caller, id string, cmd WorkGroupCommand) (domain.Work, error) {
	return worksRequest(ctx, s, c, "SetWorkGroup", struct {
		ID      string
		Command WorkGroupCommand
	}{id, cmd}, func() (domain.Work, error) { return s.setWorkGroup(ctx, c, id, cmd) })
}

func (s *worksService) setWorkGroup(ctx context.Context, c domain.Caller, id string, cmd WorkGroupCommand) (domain.Work, error) {
	if err := human(c); err != nil {
		return domain.Work{}, err
	}
	if !domain.ValidWorkGroup(cmd.Group) {
		return domain.Work{}, workError(apierrors.ValidationFailed, "group label is empty or over 200 characters")
	}
	return s.saveMetadata(ctx, c, id, cmd.ExpectedRevision, func(m domain.WorkHumanMetadata) domain.WorkHumanMetadata {
		m.Group = cmd.Group
		return m
	}, domain.WorkEventGroupSet)
}

func (s *worksService) SetWorkHandoff(ctx context.Context, c domain.Caller, id string, cmd WorkHandoffCommand) (domain.Work, error) {
	return worksRequest(ctx, s, c, "SetWorkHandoff", struct {
		ID      string
		Command WorkHandoffCommand
	}{id, cmd}, func() (domain.Work, error) { return s.setWorkHandoff(ctx, c, id, cmd) })
}

func (s *worksService) setWorkHandoff(ctx context.Context, c domain.Caller, id string, cmd WorkHandoffCommand) (domain.Work, error) {
	if err := human(c); err != nil {
		return domain.Work{}, err
	}
	if !utf8.ValidString(cmd.Content) || len(cmd.Content) > 64*1024 {
		return domain.Work{}, workError(apierrors.ValidationFailed, "handoff manual is invalid or over 64 KiB")
	}
	return s.saveMetadata(ctx, c, id, cmd.ExpectedRevision, func(m domain.WorkHumanMetadata) domain.WorkHumanMetadata {
		stamp := s.clock().UTC()
		m.Handoff = domain.WorkHandoff{Content: cmd.Content, UpdatedAt: &stamp, Source: domain.HandoffSourceHuman}
		return m
	}, domain.WorkEventHandoffSet)
}

func (s *worksService) SetWorkHidden(ctx context.Context, c domain.Caller, id string, cmd WorkHiddenCommand) (domain.Work, error) {
	return worksRequest(ctx, s, c, "SetWorkHidden", struct {
		ID      string
		Command WorkHiddenCommand
	}{id, cmd}, func() (domain.Work, error) { return s.setWorkHidden(ctx, c, id, cmd) })
}

func (s *worksService) setWorkHidden(ctx context.Context, c domain.Caller, id string, cmd WorkHiddenCommand) (domain.Work, error) {
	if err := human(c); err != nil {
		return domain.Work{}, err
	}
	return s.saveMetadata(ctx, c, id, cmd.ExpectedRevision, func(m domain.WorkHumanMetadata) domain.WorkHumanMetadata {
		m.Hidden = cmd.Hidden
		return m
	}, domain.WorkEventHidden)
}

func (s *worksService) emit(ctx context.Context, snapshot domain.WorksSnapshot, kind string) error {
	for _, work := range snapshot.Works {
		if err := s.emitWork(ctx, work, kind); err != nil {
			return err
		}
	}
	return nil
}

func (s *worksService) emitWork(ctx context.Context, work domain.Work, kind string) error {
	if s.events == nil {
		return nil
	}
	payload, _ := json.Marshal(work)
	event := domain.WorkEvent{
		WorkID: work.ID, Kind: kind, Source: domain.WorkEventSourceManual,
		Scope: work.Root, At: s.clock().UTC(), Payload: payload,
	}
	if kind == domain.WorkEventSynced || kind == domain.WorkEventVerified {
		event.Source = domain.WorkEventSourceMachine
	}
	if err := s.events.EmitWorkEvent(ctx, event); err != nil {
		return workError(apierrors.DeliveryUnknown, "work saved but event delivery outcome is unknown; reload without replay")
	}
	return nil
}
