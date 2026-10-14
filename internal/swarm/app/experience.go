package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

// 受信任的既有 outbox consumer 使用此端口；HTTP 不提供人工构造事件入口。
func (s *Service) ReadAcceptedArtifactEvent(ctx context.Context, eventID string) (AcceptedArtifactEvent, error) {
	reader, ok := s.repo.(AcceptedArtifactEventReader)
	if !ok {
		return AcceptedArtifactEvent{}, domain.Failure(apierrors.UnsupportedCapability, "actual accepted artifact event reader is unavailable")
	}
	return reader.ReadAcceptedArtifactEvent(ctx, eventID)
}

func (s *Service) LinkAcceptedExperience(ctx context.Context, reader ExperienceCandidateReader, eventID, candidateID string) (domain.Artifact, error) {
	var out domain.Artifact
	if reader == nil {
		return out, domain.Failure(apierrors.UnsupportedCapability, "actual experience candidate source reader required")
	}
	event, err := s.ReadAcceptedArtifactEvent(ctx, eventID)
	if err != nil {
		return out, err
	}
	source, err := reader.ReadArtifactCandidateSource(ctx, candidateID)
	if err != nil {
		return out, err
	}
	if source.ID != candidateID || source.Version < 1 || source.OwnerProjectID != event.ProjectID || source.AcceptanceEventID != event.EventID || source.ArtifactID != event.Artifact.ID {
		return out, domain.Failure(apierrors.EvidenceMissing, "experience candidate does not resolve to the actual accepted artifact event")
	}
	err = s.repo.Within(ctx, func(tx Transaction) error {
		var err error
		out, err = tx.Artifact(event.Artifact.ID)
		if err != nil {
			return err
		}
		if out.State != "accepted" || out.AcceptanceID != event.Artifact.AcceptanceID {
			return domain.Failure(apierrors.VersionConflict, "accepted artifact identity changed")
		}
		if out.ExperienceID != "" {
			if out.ExperienceID != candidateID {
				return domain.Failure(apierrors.VersionConflict, "acceptance already references a different experience candidate")
			}
			return nil
		}
		fresh, err := tx.Consume("swarm.accepted_experience_projection", eventID)
		if err != nil {
			return err
		}
		if !fresh {
			return domain.Failure(apierrors.EvidenceMissing, "experience projection receipt exists without its artifact reference")
		}
		old := out.Version
		out.Version++
		out.ExperienceID = candidateID
		out.ExperienceStatus = "candidate_created"
		if err = tx.SaveArtifact(out, old); err != nil {
			return err
		}
		return tx.Event(Event{ID: uuid.NewString(), Type: "swarm.artifact_experience_linked", AggregateID: out.ID, Version: out.Version, At: time.Now().UTC(), Payload: out})
	})
	return out, err
}
