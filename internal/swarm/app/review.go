package app

import (
	"context"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func (s *Service) RejectArtifact(ctx context.Context, p domain.Principal, c RejectCommand) (domain.Artifact, error) {
	var out domain.Artifact
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if err := human(p); err != nil {
		return out, err
	}
	if strings.TrimSpace(c.Reason) == "" {
		return out, domain.Failure(apierrors.ValidationFailed, "artifact review reason required")
	}
	err := s.authorized(ctx, p, c.MissionID, "reject_artifact", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.Artifact](tx, p, "reject_artifact", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			out, err = tx.Artifact(c.ArtifactID)
			if err != nil {
				return err
			}
			if err = checkVersion(out.Version, c.Meta.ExpectedVersion); err != nil {
				return err
			}
			w, err := tx.WorkItem(out.WorkItemID)
			if err != nil {
				return err
			}
			if w.MissionID != c.MissionID {
				return domain.Failure(apierrors.ScopeDenied, "artifact belongs to another mission")
			}
			if out.State != "submitted" || w.State != "submitted" || w.ArtifactID != out.ID {
				return domain.Failure(apierrors.VersionConflict, "only the current unaccepted submission can be returned for revision")
			}
			now := time.Now().UTC()
			oldA := out.Version
			out.Version++
			out.State = "revision_needed"
			out.ReviewReason = c.Reason
			out.ReviewedBy = p.ID
			out.ReviewedAt = &now
			oldW := w.Version
			w.Version++
			w.State = "revision_needed"
			w.Reason = c.Reason
			m, err := tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			oldM := m.Version
			m.Version++
			m.NoProgress++
			m.Reason = c.Reason
			if m.StopReached(now) {
				m.State = "blocked"
				m.Reason = "mission stop condition reached after artifact review"
			}
			if err = tx.SaveArtifact(out, oldA); err != nil {
				return err
			}
			if err = tx.SaveWorkItem(w, oldW); err != nil {
				return err
			}
			if err = tx.SaveMission(m, oldM); err != nil {
				return err
			}
			return finish(tx, p, "reject_artifact", c.Meta, d, out, "artifact_revision_requested", out.ID, out.Version)
		})
	})
	return out, err
}

func (s *Service) ReadArtifactDiff(ctx context.Context, p domain.Principal, missionID, id string, versions ...int) (ArtifactDiff, error) {
	var out ArtifactDiff
	if len(versions) > 1 {
		return out, domain.Failure(apierrors.ValidationFailed, "one exact artifact version is required")
	}
	version := 0
	if len(versions) == 1 {
		version = versions[0]
	}
	a, err := s.ReadArtifactVersion(ctx, p, missionID, id, version)
	if err != nil {
		return out, err
	}
	reader, ok := s.evidence.(ArtifactDiffReader)
	if !ok {
		return out, domain.Failure(apierrors.UnsupportedCapability, "actual verified artifact diff reader is unavailable")
	}
	diff, err := reader.ReadVerifiedDiff(ctx, a)
	if err != nil {
		return out, err
	}
	return ArtifactDiff{SchemaVersion: 1, ArtifactID: a.ID, Commit: a.Commit, BaseCommit: a.BaseCommit, Diff: diff}, nil
}
