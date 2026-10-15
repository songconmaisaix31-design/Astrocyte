package importers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// paperSnapshotSource 解析人工审阅的浏览器快照，原 JSON bytes 完整保留。
// 来源由观察到的 DOI/arXiv/URL 重新计算，正文保留实际访问状态及截断标记。
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
		// 原始快照 JSON 保留实际观察来源。
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
		// 当前观察只有元数据，不能将未经核实的正文作为论文全文。
		source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract (metadata only)\n\n%s", snap.Title, snap.SourceURL, snap.Abstract)
		source.Provenance.Mode = "browser_snapshot_abstract_only"
	case "paywall", "restricted":
		return source, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "Access restriction detected; no entitlement bypass", RequiredAction: "choose_accessible_public_paper_or_provide_existing_export"}
	default:
		return source, invalid("paper snapshot has an unsupported content_state")
	}
	return source, nil
}

// paperSnapshotSourceKey 按既有 arXiv/ACL/DOI 顺序重新计算来源身份，
// 让同一论文的快照、PDF 和页面导入共用资料身份。
func paperSnapshotSourceKey(snap PaperSnapshot) (string, error) {
	if strings.TrimSpace(snap.ArxivID) != "" {
		id, err := NormalizeArxivID(snap.ArxivID)
		if err != nil {
			return "", invalid("paper snapshot arxiv_id is invalid")
		}
		return "arxiv:" + revisionSuffix.ReplaceAllString(id, ""), nil
	}
	// ACL Anthology URL 内含 DOI 身份，复用 paper_pdf 的相同计算方式。
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
