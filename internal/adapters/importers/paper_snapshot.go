package importers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// paperSnapshotSource parses a human-reviewed browser-plugin snapshot and maps
// it to an importable source. It performs no network access, no paywall bypass
// and no model call: the bytes are the exact JSON the human reviewed and pasted.
//
// Provenance is preserved as a browser_snapshot (not a publisher-verified
// network export): the material source key is re-derived from the observed
// DOI/arXiv/URL identity rather than trusted from the snapshot's own source_key,
// so a hand-edited or conflicting key cannot merge different texts. A summary or
// truncated observation is never presented as full text.
func paperSnapshotSource(raw []byte, locator string) (app.ImportedSource, error) {
	var source app.ImportedSource
	var snap PaperSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return source, invalid("paper snapshot JSON is malformed")
	}
	if snap.SchemaVersion != 1 {
		return source, invalid("unsupported paper snapshot schema version")
	}
	if snap.SourceURL == "" || snap.SourceURL != locator {
		return source, invalid("paper snapshot source_url must match the requested source")
	}
	if u, err := url.Parse(snap.SourceURL); err != nil || validatePublicPaperURL(u) != nil {
		return source, invalid("paper snapshot source_url must be a public HTTPS URL")
	}
	key, err := paperSnapshotSourceKey(snap)
	if err != nil {
		return source, err
	}
	version := snap.Provenance.Version
	if strings.TrimSpace(version) == "" {
		version = "browser snapshot"
	}
	if observed := strings.TrimSpace(snap.ObservedVersion); observed != "" {
		version += "; observed " + observed
	}
	source = app.ImportedSource{
		SourceKey:     key,
		SourceLocator: snap.SourceURL,
		Kind:          "paper",
		Title:         snap.Title,
		Summary:       snap.Abstract,
		SourceSpans:   []string{"browser snapshot: " + snap.SourceURL},
		Provenance:    app.Provenance{Processor: "paper_snapshot", Version: version, Source: snap.SourceURL},
		// The original snapshot JSON is the provenance truth; keep it verbatim.
		Attachments: []app.SourceAttachment{{Name: "paper-snapshot.json", MediaType: "application/json", Data: append([]byte(nil), raw...), SourceLocator: snap.SourceURL}},
	}
	switch snap.ContentState {
	case "readable_fulltext":
		if strings.TrimSpace(snap.Text) == "" {
			return source, invalid("paper snapshot claims full text but body is empty")
		}
		if snap.Truncated {
			source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract\n\n%s\n\n## Full text (browser snapshot, truncated)\n\n%s", snap.Title, snap.SourceURL, snap.Abstract, snap.Text)
			source.Provenance.Mode = "browser_snapshot_truncated"
		} else {
			source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract\n\n%s\n\n## Full text (browser snapshot)\n\n%s", snap.Title, snap.SourceURL, snap.Abstract, snap.Text)
			source.Provenance.Mode = "browser_snapshot_fulltext"
		}
	case "abstract_only":
		// Metadata only. Any captured body is deliberately dropped here because
		// it was never established as readable full-paper structure.
		source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract (metadata only)\n\n%s", snap.Title, snap.SourceURL, snap.Abstract)
		source.Provenance.Mode = "browser_snapshot_abstract_only"
	case "paywall", "restricted":
		return source, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "Access restriction detected; no entitlement bypass", RequiredAction: "choose_accessible_public_paper_or_provide_existing_export"}
	default:
		return source, invalid("paper snapshot has an unsupported content_state")
	}
	return source, nil
}

// paperSnapshotSourceKey re-derives the stable material identity from the
// snapshot's observed fields, never from its self-reported source_key. Priority
// matches the existing arXiv/ACL/DOI importers so a snapshot and a paper_pdf or
// paper_url import of the same paper share one material identity.
func paperSnapshotSourceKey(snap PaperSnapshot) (string, error) {
	if strings.TrimSpace(snap.ArxivID) != "" {
		id, err := NormalizeArxivID(snap.ArxivID)
		if err != nil {
			return "", invalid("paper snapshot arxiv_id is invalid")
		}
		return "arxiv:" + revisionSuffix.ReplaceAllString(id, ""), nil
	}
	// The ACL Anthology URL embeds its DOI identity; use it exactly as the
	// paper_pdf adapter does so the two paths correlate without case drift.
	if key, ok := aclAnthologySourceKey(snap.SourceURL); ok {
		return key, nil
	}
	if strings.TrimSpace(snap.DOI) != "" {
		doi := normalizeDOIRaw(snap.DOI)
		if doi == "" {
			return "", invalid("paper snapshot DOI is malformed")
		}
		return "doi:" + doi, nil
	}
	return canonicalWebKey(snap.SourceURL), nil
}
