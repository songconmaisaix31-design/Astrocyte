package app

import (
	"context"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type TasteRubric struct {
	ID           string `json:"id"`
	QuestionType string `json:"question_type"`
	Criteria     string `json:"criteria"`
}

type TasteExample struct {
	ID            string `json:"id"`
	RubricID      string `json:"rubric_id"`
	MaterialID    string `json:"material_id"`
	Revision      int    `json:"revision"`
	ContentDigest string `json:"content_digest"`
	Expected      string `json:"expected"`
}

// 委托仅固定原 S5 规则引用，当前授权由 S5 在原事务中重新核验。
type TasteDelegation struct {
	RuleID     string `json:"rule_id"`
	Revision   int    `json:"revision"`
	WorkflowID string `json:"workflow_id"`
}

type TasteProfile struct {
	ID           string            `json:"id"`
	Version      int               `json:"version"`
	Enabled      bool              `json:"enabled"`
	Rubrics      []TasteRubric     `json:"rubrics"`
	Examples     []TasteExample    `json:"examples"`
	Delegations  []TasteDelegation `json:"delegations"`
	CreatedAt    time.Time         `json:"created_at"`
	ConfiguredBy string            `json:"configured_by"`
}

type TasteProfileDetail struct {
	SchemaVersion int           `json:"schema_version"`
	Configured    bool          `json:"configured"`
	Profile       *TasteProfile `json:"profile"`
}

type TasteProfileSnapshot struct {
	SchemaVersion int           `json:"schema_version"`
	Configured    bool          `json:"configured"`
	Current       *TasteProfile `json:"current"`
	Fixed         *TasteProfile `json:"fixed"`
}

type UpdateTasteProfileCommand struct {
	CommandMeta
	Enabled     bool              `json:"enabled"`
	Rubrics     []TasteRubric     `json:"rubrics"`
	Examples    []TasteExample    `json:"examples"`
	Delegations []TasteDelegation `json:"delegations"`
}

type TasteProfileTx interface {
	LoadTasteProfile() (TasteProfileDetail, error)
	LoadTasteProfileRevision(string, int) (TasteProfile, error)
	SaveTasteProfile(TasteProfile, int) error
}

type TasteProfileService interface {
	GetTasteProfile(context.Context, Principal) (TasteProfileDetail, error)
	UpdateTasteProfile(context.Context, Principal, UpdateTasteProfileCommand) (TasteProfileDetail, error)
	ReadTasteProfile(context.Context, Principal, string, int) (TasteProfileSnapshot, error)
}

var _ TasteProfileService = (*Service)(nil)

func tasteProfilePort(tx AttentionTx) (TasteProfileTx, error) {
	port, ok := tx.(TasteProfileTx)
	if !ok {
		return nil, serviceError(apierrors.UnsupportedCapability, "Taste profile storage is not configured", "configure_taste_profile_storage")
	}
	return port, nil
}

func (s *Service) GetTasteProfile(ctx context.Context, p Principal) (TasteProfileDetail, error) {
	result := TasteProfileDetail{SchemaVersion: 1}
	if err := authorize(p, false); err != nil {
		return result, err
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, err := tasteProfilePort(tx)
		if err != nil {
			return err
		}
		result, err = port.LoadTasteProfile()
		return err
	})
	return result, mapError(err, "")
}

func (s *Service) UpdateTasteProfile(ctx context.Context, p Principal, c UpdateTasteProfileCommand) (TasteProfileDetail, error) {
	return command(s, ctx, p, c.CommandMeta, "UpdateTasteProfile", c, func(tx AttentionTx) (TasteProfileDetail, error) {
		port, err := tasteProfilePort(tx)
		if err != nil {
			return TasteProfileDetail{}, err
		}
		current, err := port.LoadTasteProfile()
		if err != nil {
			return TasteProfileDetail{}, err
		}
		expected := 0
		if current.Profile != nil {
			expected = current.Profile.Version
		}
		if (expected == 0 && c.ExpectedVersion != 1) || (expected > 0 && c.ExpectedVersion != expected) {
			return TasteProfileDetail{}, domain.ErrVersion
		}
		profile := TasteProfile{ID: "attention", Version: expected + 1, Enabled: c.Enabled, Rubrics: c.Rubrics, Examples: c.Examples,
			Delegations: c.Delegations, CreatedAt: s.options.Clock().UTC(), ConfiguredBy: p.ID}
		if err := ValidateTasteProfile(profile); err != nil {
			return TasteProfileDetail{}, err
		}
		if err := validateTasteExamples(tx, profile); err != nil {
			return TasteProfileDetail{}, err
		}
		if err := port.SaveTasteProfile(profile, expected); err != nil {
			return TasteProfileDetail{}, err
		}
		if err := s.event(tx, "taste_profile_updated", profile.ID, profile.Version, c.CommandMeta,
			map[string]any{"profile_id": profile.ID, "profile_version": profile.Version, "enabled": profile.Enabled}); err != nil {
			return TasteProfileDetail{}, err
		}
		return TasteProfileDetail{SchemaVersion: 1, Configured: true, Profile: &profile}, nil
	})
}

// 当前版本与固定版本共用原事务，不能凭历史快照绕过已撤回资料的权限。
func (s *Service) ReadTasteProfile(ctx context.Context, p Principal, id string, version int) (TasteProfileSnapshot, error) {
	result := TasteProfileSnapshot{SchemaVersion: 1}
	if err := authorize(p, false); err != nil {
		return result, err
	}
	if id != "attention" || version < 1 {
		return result, mapError(domain.ErrInvalid, "")
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, err := tasteProfilePort(tx)
		if err != nil {
			return err
		}
		current, err := port.LoadTasteProfile()
		if err != nil {
			return err
		}
		if !current.Configured {
			return nil
		}
		fixed, err := port.LoadTasteProfileRevision(id, version)
		if err != nil {
			return err
		}
		if err := validateTasteExamples(tx, fixed); err != nil {
			return err
		}
		if err := validateTasteExamples(tx, *current.Profile); err != nil {
			return err
		}
		result.Configured, result.Current, result.Fixed = true, current.Profile, &fixed
		return nil
	})
	return result, mapError(err, "")
}

// 验证明确输入的配置结构，不生成品味、示例或委托规则。
func ValidateTasteProfile(profile TasteProfile) error {
	if profile.ID != "attention" || profile.Version < 1 || profile.CreatedAt.IsZero() || strings.TrimSpace(profile.ConfiguredBy) == "" ||
		profile.Rubrics == nil || profile.Examples == nil || profile.Delegations == nil || len(profile.Rubrics) > 256 || len(profile.Examples) > 256 || len(profile.Delegations) > 256 {
		return domain.ErrInvalid
	}
	if profile.Enabled && (len(profile.Rubrics) == 0 || len(profile.Examples) == 0) {
		return domain.ErrEvidence
	}
	rubrics := map[string]bool{}
	for _, rubric := range profile.Rubrics {
		if strings.TrimSpace(rubric.ID) == "" || strings.TrimSpace(rubric.QuestionType) == "" || strings.TrimSpace(rubric.Criteria) == "" || rubrics[rubric.ID] {
			return domain.ErrInvalid
		}
		rubrics[rubric.ID] = true
	}
	examples := map[string]bool{}
	for _, example := range profile.Examples {
		if strings.TrimSpace(example.ID) == "" || examples[example.ID] || !rubrics[example.RubricID] || example.MaterialID == "" ||
			example.Revision < 1 || strings.TrimSpace(example.ContentDigest) == "" || strings.TrimSpace(example.Expected) == "" {
			return domain.ErrInvalid
		}
		examples[example.ID] = true
	}
	delegations := map[string]bool{}
	for _, delegation := range profile.Delegations {
		if strings.TrimSpace(delegation.RuleID) == "" || delegation.Revision < 1 || strings.TrimSpace(delegation.WorkflowID) == "" || delegations[delegation.RuleID] {
			return domain.ErrInvalid
		}
		delegations[delegation.RuleID] = true
	}
	return nil
}

func validateTasteExamples(tx AttentionTx, profile TasteProfile) error {
	for _, example := range profile.Examples {
		material, err := tx.LoadMaterial(example.MaterialID)
		if err != nil {
			return err
		}
		if material.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Taste example material was withdrawn", "choose_active_material")
		}
		found := false
		for _, revision := range material.Revisions {
			if revision.Revision == example.Revision {
				if revision.ContentDigest != example.ContentDigest {
					return serviceError(apierrors.VersionConflict, "Taste example differs from the fixed material digest", "reload_fixed_material_revision")
				}
				found = true
				break
			}
		}
		if !found {
			return apierrors.NewNotFound("material_revision", example.MaterialID)
		}
	}
	return nil
}
