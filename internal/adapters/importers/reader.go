package importers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Reader accepts existing exports. Local files are disabled unless explicit
// allowed roots are supplied by the application's trusted configuration.
type Reader struct {
	Arxiv        *Arxiv
	Summarize    *SummarizeExtractor
	AllowedRoots []string
}

func NewReader(allowedRoots []string) *Reader {
	return &Reader{Arxiv: NewArxiv(), AllowedRoots: append([]string(nil), allowedRoots...)}
}

var _ app.SourceReader = (*Reader)(nil)

func (r *Reader) ReadSource(ctx context.Context, cmd app.ImportMaterialCommand) (app.ImportedSource, error) {
	if err := ctx.Err(); err != nil {
		return app.ImportedSource{}, err
	}
	adapter := cmd.Adapter
	if adapter == "" {
		if cmd.Kind == "paper" && cmd.ExportText == "" && cmd.LocalFileRef == "" {
			adapter = "arxiv"
		} else {
			adapter = "manual"
		}
	}
	var source app.ImportedSource
	switch adapter {
	case "arxiv":
		p, err := r.Arxiv.Read(ctx, cmd.SourceLocator)
		if err != nil {
			return source, sourceError(err)
		}
		digest := sha256.Sum256(p.PDF)
		source = app.ImportedSource{SourceKey: p.SourceKey, SourceLocator: "https://arxiv.org/abs/" + p.ID, Kind: "paper", Title: p.Title, Text: fmt.Sprintf("# %s\n\nSource: https://arxiv.org/abs/%s\n\nOriginal full paper (PDF): %s\nPDF SHA256: %x\n\n## Abstract (metadata only)\n\n%s\n", p.Title, p.ID, p.PDFURL, digest, p.Abstract), Summary: p.Abstract, SourceSpans: []string{"whole document: " + p.PDFURL}, Provenance: app.Provenance{Processor: "arxiv", Version: p.ID, Mode: "official_atom_and_pdf", Source: p.PDFURL}, Attachments: []app.SourceAttachment{{Name: p.ID + ".pdf", MediaType: "application/pdf", Data: p.PDF, SourceLocator: p.PDFURL}, {Name: "metadata.atom.xml", MediaType: "application/atom+xml", Data: p.Metadata, SourceLocator: "https://export.arxiv.org/api/query?id_list=" + url.QueryEscape(p.ID)}}}
		if p.FullText != "" {
			source.Text = fmt.Sprintf("# %s\n\nSource: %s\nOriginal PDF: %s\n\n## Original HTML text (summarize extraction)\n\n%s", p.Title, p.HTMLURL, p.PDFURL, p.FullText)
			source.SourceSpans = append(source.SourceSpans, "whole HTML document: "+p.HTMLURL)
			source.Provenance = app.Provenance{Processor: "arxiv+summarize", Version: p.ID + "; summarize " + p.SummarizeVersion, Mode: "official_atom_pdf_and_html_text", Source: p.HTMLURL}
			source.Attachments = append(source.Attachments, app.SourceAttachment{Name: p.ID + ".html", MediaType: "text/html", Data: p.HTML, SourceLocator: p.HTMLURL}, app.SourceAttachment{Name: "summarize-original.json", MediaType: "application/json", Data: p.TextExport, SourceLocator: p.HTMLURL})
		}
	case "summarize_url":
		if cmd.ExportText != "" || cmd.LocalFileRef != "" || (cmd.Kind != "" && cmd.Kind != "video") {
			return source, invalid("summarize_url requires a video URL without an existing export")
		}
		if err := validateVideoURL(cmd.SourceLocator); err != nil {
			return source, err
		}
		if r.Summarize == nil {
			return source, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "project summarize URL extraction is disabled", RequiredAction: "enable_project_summarize_or_provide_existing_export"}
		}
		e, err := r.Summarize.ExtractVideo(ctx, cmd.SourceLocator)
		if err != nil {
			return source, sourceError(err)
		}
		name := "summarize-cli-original.json"
		if e.Mode == "upstream_media_transcript" {
			name = "summarize-upstream-media.json"
		}
		attachments := []app.SourceAttachment{{Name: name, MediaType: "application/json", Data: e.Original, SourceLocator: e.URL}}
		for _, a := range e.ExtraAttachments {
			attachments = append(attachments, app.SourceAttachment{Name: a.Name, MediaType: a.MediaType, Data: a.Data, SourceLocator: e.URL})
		}
		version := e.Version
		if e.TranscriptSource != "" {
			version += "; " + e.TranscriptSource
		}
		source = app.ImportedSource{SourceKey: canonicalWebKey(e.URL), SourceLocator: e.URL, Kind: "video", Title: e.Title, Text: e.Text, Summary: e.Summary, SourceSpans: exportSpans(e), Provenance: app.Provenance{Processor: "summarize", Version: version, Mode: e.Mode, Source: e.URL}, Attachments: attachments}
	case "summarize", "summarize_json", "summarize_markdown":
		raw, err := r.exportBytes(cmd)
		if err != nil {
			return source, err
		}
		var e Export
		mediaType, name := "text/markdown", "summarize.md"
		mode := "existing_markdown_export"
		if adapter == "summarize_json" || (adapter == "summarize" && strings.HasPrefix(strings.TrimSpace(string(bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf}))), "{")) {
			e, err = ParseSummarizeJSON(raw)
			mediaType, name = "application/json", "summarize.json"
			mode = "existing_json_export"
			if !e.HasOriginal {
				mode = "existing_summary_only_export"
			}
		} else {
			e, err = ParseSummarizeMarkdown(cmd.SourceLocator, cmd.Title, raw)
		}
		if err != nil {
			return source, sourceError(err)
		}
		if cmd.SourceLocator != "" && canonicalWebKey(cmd.SourceLocator) != canonicalWebKey(e.URL) {
			return source, invalid("export source does not match requested source")
		}
		kind := cmd.Kind
		if kind == "" {
			kind = "video"
		}
		spans := exportSpans(e)
		source = app.ImportedSource{SourceKey: exportSourceKey(e.URL, kind), SourceLocator: e.URL, Kind: kind, Title: e.Title, Text: e.Text, Summary: e.Summary, SourceSpans: spans, Provenance: app.Provenance{Processor: "summarize", Version: "0.21.8-format", Mode: mode, Source: e.URL}, Attachments: []app.SourceAttachment{{Name: name, MediaType: mediaType, Data: raw, SourceLocator: e.URL}}}
	case "manual":
		raw, err := r.exportBytes(cmd)
		if err != nil {
			return source, err
		}
		if err = validateWebURL(cmd.SourceLocator); err != nil {
			return source, sourceError(err)
		}
		source = app.ImportedSource{SourceKey: exportSourceKey(cmd.SourceLocator, cmd.Kind), SourceLocator: cmd.SourceLocator, Kind: cmd.Kind, Title: cmd.Title, Text: string(raw), SourceSpans: append([]string{}, cmd.SourceSpans...), Provenance: app.Provenance{Processor: "manual", Version: "1", Mode: "manual", Source: cmd.SourceLocator}}
	default:
		return source, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "source adapter unavailable", RequiredAction: "use_arxiv_or_existing_summarize_export"}
	}
	if cmd.SourceKey != "" && cmd.SourceKey != source.SourceKey {
		return app.ImportedSource{}, invalid("source_key differs from canonical source")
	}
	if cmd.ContentDigest != "" {
		digest := sha256.Sum256([]byte(source.Text))
		if cmd.ContentDigest != hex.EncodeToString(digest[:]) {
			return app.ImportedSource{}, invalid("content_digest differs from imported content")
		}
	}
	return source, nil
}

func exportSpans(e Export) []string {
	spans := []string{}
	for _, seg := range e.Segments {
		span := fmt.Sprintf("%s [%.3fs", e.URL, seg.StartMS/1000)
		if seg.EndMS != nil {
			span += fmt.Sprintf("–%.3fs", *seg.EndMS/1000)
		}
		spans = append(spans, span+"]: "+seg.Text)
	}
	return spans
}

func (r *Reader) exportBytes(cmd app.ImportMaterialCommand) ([]byte, error) {
	if cmd.ExportText != "" && cmd.LocalFileRef != "" {
		return nil, invalid("supply export_text or local_file_ref, not both")
	}
	if cmd.LocalFileRef == "" {
		if strings.TrimSpace(cmd.ExportText) == "" {
			return nil, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "existing export is required; no extraction backend was selected", RequiredAction: "provide_summarize_json_or_markdown_export"}
		}
		if len(cmd.ExportText) > 16<<20 {
			return nil, invalid("export exceeds size limit")
		}
		return []byte(cmd.ExportText), nil
	}
	resolved, err := filepath.EvalSymlinks(cmd.LocalFileRef)
	if err != nil {
		return nil, invalid("local export is unavailable")
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return nil, err
	}
	allowed := false
	for _, root := range r.AllowedRoots {
		root, err = filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		root, err = filepath.Abs(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(root, resolved)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "export path is outside configured import roots", RequiredAction: "use_export_text_or_register_import_root"}
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return nil, invalid("local export must be a regular file within the size limit")
	}
	return os.ReadFile(resolved)
}

func canonicalWebKey(raw string) string { return domain.CanonicalWebKey(raw) }

// Official paper locators retain their exact original version/URL in the
// export and provenance, while sharing the existing arXiv material identity.
// Neither video URLs, arbitrary domains nor invalid paper IDs are collapsed.
func exportSourceKey(locator, kind string) string {
	if kind == "paper" {
		if id, err := NormalizeArxivID(locator); err == nil {
			return "arxiv:" + revisionSuffix.ReplaceAllString(id, "")
		}
	}
	return canonicalWebKey(locator)
}

func invalid(message string) *apierrors.ServiceError {
	return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: message, RequiredAction: "correct_source_input"}
}

func sourceError(err error) error {
	var service *apierrors.ServiceError
	if errors.As(err, &service) {
		return err
	}
	if strings.Contains(err.Error(), "unavailable") || strings.Contains(err.Error(), "HTTP") {
		return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "official source is unavailable", Retryable: true, RequiredAction: "retry_or_provide_existing_export"}
	}
	return invalid(err.Error())
}
