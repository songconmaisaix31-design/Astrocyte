package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

var _ SourceCollectionDiscoveryService = (*Service)(nil)

func (s *Service) catalogIdentity(platform, owner, mode string) (string, string, error) {
	owner = strings.TrimSpace(owner)
	mode = normalizedSourceAccess(mode)
	if platform == "douyin" && owner == "self" && mode == "browser_selected" {
		// A local accessor resolves the handed-off identity without page reads.
		configured, ok := s.options.CollectionReader.(interface{ ConfiguredSelectedOwner() string })
		if !ok || configured.ConfiguredSelectedOwner() == "" {
			return "", mode, serviceError(apierrors.ProviderUnavailable, "The selected browser account is not configured", "connect_selected_douyin_browser")
		}
		owner = configured.ConfiguredSelectedOwner()
	}
	if platform == "bilibili" && mode == "public" && publicDecimalID.MatchString(owner) {
		return owner, mode, nil
	}
	if platform == "douyin" && mode == "browser_selected" && owner != "self" && publicDouyinID.MatchString(owner) {
		return owner, mode, nil
	}
	return "", mode, serviceError(apierrors.ScopeDenied, "Choose public Bilibili or an explicitly selected Douyin browser account", "choose_supported_source_scope")
}

// GET is cache-only. Not-yet-discovered is explicit, not a provider-empty claim.
func (s *Service) ListSourceCollections(ctx context.Context, p Principal, platform, owner string) (apierrors.ListResult, error) {
	if err := authorize(p, true); err != nil {
		return apierrors.EmptyList(), err
	}
	mode := "public"
	if platform == "douyin" {
		mode = "browser_selected"
	}
	owner, mode, err := s.catalogIdentity(platform, owner, mode)
	if err != nil {
		return apierrors.EmptyList(), mapError(err, "")
	}
	var rows []SourceCollection
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, ok := tx.(CatalogTx)
		if !ok {
			return serviceError(apierrors.UnsupportedCapability, "Source catalog storage is unavailable", "configure_source_catalog_storage")
		}
		var err error
		rows, err = port.LoadSourceCatalog(platform, owner, mode)
		if isMissing(err) {
			return serviceError(apierrors.EvidenceMissing, "This source catalog has not been discovered", "discover_source_collections")
		}
		return err
	})
	return listResult(rows), mapError(err, "")
}

func (s *Service) DiscoverSourceCollections(ctx context.Context, p Principal, c DiscoverSourceCollectionsCommand) (apierrors.ListResult, error) {
	if err := authorize(p, true); err != nil {
		return apierrors.EmptyList(), mapError(err, c.RequestID)
	}
	if s.configurationError != nil {
		return apierrors.EmptyList(), mapError(s.configurationError, c.RequestID)
	}
	if err := validateMeta(c.CommandMeta); err != nil {
		return apierrors.EmptyList(), mapError(err, c.RequestID)
	}
	if c.ExpectedVersion != 1 {
		return apierrors.EmptyList(), mapError(domain.ErrVersion, c.RequestID)
	}
	owner, mode, err := s.catalogIdentity(c.Platform, c.OwnerID, c.AccessMode)
	if err != nil {
		return apierrors.EmptyList(), mapError(err, c.RequestID)
	}
	c.OwnerID, c.AccessMode = owner, mode
	// Check the ordinary receipt before external read, then command checks it
	// again when committing. No SQLite write lock is held during browser I/O.
	digest, err := semanticDigest(c)
	if err != nil {
		return apierrors.EmptyList(), mapError(err, c.RequestID)
	}
	var replay apierrors.ListResult
	found := false
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		receipt, err := tx.LoadReceipt(p.Kind+":"+p.ID, "DiscoverSourceCollections", c.IdempotencyKey)
		if isMissing(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if receipt.Digest != digest {
			return domain.ErrVersion
		}
		found = true
		return json.Unmarshal(receipt.Result, &replay)
	})
	if err != nil || found {
		return replay, mapError(err, c.RequestID)
	}
	if s.options.CollectionReader == nil {
		return apierrors.EmptyList(), serviceError(apierrors.ProviderUnavailable, "No source collection reader is configured", "configure_source_collection_reader")
	}
	rows, err := s.options.CollectionReader.ListCollections(ctx, c.Platform, owner)
	if err != nil {
		return apierrors.EmptyList(), mapError(err, c.RequestID)
	}
	seen := map[string]bool{}
	for i := range rows {
		row := &rows[i]
		if row.OwnerID != owner || row.ExternalID == "" || row.Title == "" || row.Locator == "" || seen[row.ExternalID] || (row.ItemCount != nil && *row.ItemCount < 0) {
			return apierrors.EmptyList(), serviceError(apierrors.EvidenceMissing, "Source catalog identity or metadata could not be verified", "inspect_source_collection_catalog")
		}
		row.AccessMode = mode
		seen[row.ExternalID] = true
	}
	return command(s, ctx, p, c.CommandMeta, "DiscoverSourceCollections", c, func(tx AttentionTx) (apierrors.ListResult, error) {
		port, ok := tx.(CatalogTx)
		if !ok {
			return apierrors.EmptyList(), serviceError(apierrors.UnsupportedCapability, "Source catalog storage is unavailable", "configure_source_catalog_storage")
		}
		if err := port.SaveSourceCatalog(c.Platform, owner, mode, nonNil(rows)); err != nil {
			return apierrors.EmptyList(), err
		}
		if err := s.event(tx, "source_catalog_discovered", c.Platform+":"+owner, 1, c.CommandMeta, map[string]any{"platform": c.Platform, "owner_id": owner, "access_mode": mode, "folder_count": len(rows)}); err != nil {
			return apierrors.EmptyList(), err
		}
		return listResult(rows), nil
	})
}
