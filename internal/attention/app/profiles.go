package app

import (
	"context"
	"maps"
	"sort"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

var _ RankingProfileService = (*Service)(nil)

func (s *Service) GetRankingProfile(ctx context.Context, p Principal) (RankingProfileDetail, error) {
	if err := authorize(p, false); err != nil {
		return RankingProfileDetail{}, err
	}
	result := RankingProfileDetail{SchemaVersion: 1, Versions: []RankingProfile{}}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, ok := tx.(RankingProfileTx)
		if !ok {
			return serviceError(apierrors.UnsupportedCapability, "Ranking profile storage is not configured", "configure_ranking_profile_storage")
		}
		row, err := port.LoadRankingProfile()
		if isMissing(err) {
			return nil
		}
		if err != nil {
			return err
		}
		result = row
		result.Versions = nonNil(result.Versions)
		return nil
	})
	return result, mapError(err, "")
}

func (s *Service) UpdateRankingProfile(ctx context.Context, p Principal, c UpdateRankingProfileCommand) (RankingProfileDetail, error) {
	return command(s, ctx, p, c.CommandMeta, "UpdateRankingProfile", c, func(tx AttentionTx) (RankingProfileDetail, error) {
		port, ok := tx.(RankingProfileTx)
		if !ok {
			return RankingProfileDetail{}, serviceError(apierrors.UnsupportedCapability, "Ranking profile storage is not configured", "configure_ranking_profile_storage")
		}
		row, err := port.LoadRankingProfile()
		if err != nil && !isMissing(err) {
			return RankingProfileDetail{}, err
		}
		base := domain.TasteProfile{}
		if row.Profile != nil {
			base = domain.TasteProfile{Version: row.Profile.Version, Enabled: row.Profile.Enabled, Weights: row.Profile.Weights}
		}
		if (base.Version == 0 && c.ExpectedVersion != 1) || (base.Version > 0 && c.ExpectedVersion != base.Version) {
			return RankingProfileDetail{}, domain.ErrVersion
		}
		next, err := base.WithWeights(c.Enabled, c.Weights)
		if err != nil {
			return RankingProfileDetail{}, err
		}
		profile := RankingProfile{ID: "attention", Version: next.Version, Enabled: next.Enabled, Weights: next.Weights, CreatedAt: s.options.Clock()}
		row.SchemaVersion = 1
		row.Configured = true
		row.Profile = &profile
		row.Versions = append(nonNil(row.Versions), profile)
		if err = port.SaveRankingProfile(row, base.Version); err != nil {
			return RankingProfileDetail{}, err
		}
		if err = s.event(tx, "ranking_profile_updated", profile.ID, profile.Version, c.CommandMeta, map[string]any{"profile_id": profile.ID, "profile_version": profile.Version, "enabled": profile.Enabled}); err != nil {
			return RankingProfileDetail{}, err
		}
		return row, nil
	})
}

func projectOpportunityRanking(o Opportunity, profile *RankingProfile) Opportunity {
	o.CompositeScore = nil
	o.RankingProfileVersion = nil
	o.RankingStrategy = "manual"
	o.RankingReason = "未配置完整且启用的四维 profile，保留人工顺序"
	if profile == nil || !profile.Enabled || !domain.CompleteWeights(profile.Weights) {
		return o
	}
	version := profile.Version
	o.RankingProfileVersion = &version
	o.CompositeScore = domain.CompositeScore(dimensions(o.Dimensions), profile.Weights)
	if o.CompositeScore == nil {
		o.RankingReason = "存在 unknown 维度，保留人工顺序；未知未填零"
		return o
	}
	o.RankingStrategy = "taste-profile-v1"
	o.RankingReason = "按用户完整四维权重综合排序；goal_progress 权重首要，各维非零"
	return o
}

func rankingProfile(tx AttentionTx) (*RankingProfile, error) {
	port, ok := tx.(RankingProfileTx)
	if !ok {
		return nil, nil
	}
	row, err := port.LoadRankingProfile()
	if isMissing(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if row.Profile == nil {
		return nil, nil
	}
	copy := *row.Profile
	copy.Weights = maps.Clone(row.Profile.Weights)
	return &copy, nil
}

func rankOpportunities(items []Opportunity, profile *RankingProfile) []Opportunity {
	for i := range items {
		items[i] = projectOpportunityRanking(items[i], profile)
	}
	// Keep candidates with unknown dimensions in their original manual slots;
	// rank only the known subset, rather than demoting missing values to zero.
	positions := []int{}
	known := []Opportunity{}
	for i, o := range items {
		if o.CompositeScore != nil {
			positions = append(positions, i)
			known = append(known, o)
		}
	}
	sort.SliceStable(known, func(i, j int) bool { return *known[i].CompositeScore > *known[j].CompositeScore })
	for i, position := range positions {
		items[position] = known[i]
	}
	return items
}
