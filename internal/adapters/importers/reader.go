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
)

// Reader accepts existing exports. Local files are disabled unless explicit
// allowed roots are supplied by the application's trusted configuration.
type Reader struct {
	Arxiv        *Arxiv
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
		spans := []string{}
		for _, seg := range e.Segments {
			span := fmt.Sprintf("%s [%.3fs", e.URL, seg.StartMS/1000)
			if seg.EndMS != nil {
				span += fmt.Sprintf("–%.3fs", *seg.EndMS/1000)
			}
			spans = append(spans, span+"]: "+seg.Text)
		}
		source = app.ImportedSource{SourceKey: canonicalWebKey(e.URL), SourceLocator: e.URL, Kind: kind, Title: e.Title, Text: e.Text, Summary: e.Summary, SourceSpans: spans, Provenance: app.Provenance{Processor: "summarize", Version: "0.21.8-format", Mode: mode, Source: e.URL}, Attachments: []app.SourceAttachment{{Name: name, MediaType: mediaType, Data: raw, SourceLocator: e.URL}}}
	case "manual":
		raw, err := r.exportBytes(cmd)
		if err != nil {
			return source, err
		}
		if err = validateWebURL(cmd.SourceLocator); err != nil {
			return source, sourceError(err)
		}
		source = app.ImportedSource{SourceKey: canonicalWebKey(cmd.SourceLocator), SourceLocator: cmd.SourceLocator, Kind: cmd.Kind, Title: cmd.Title, Text: string(raw), SourceSpans: append([]string{}, cmd.SourceSpans...), Provenance: app.Provenance{Processor: "manual", Version: "1", Mode: "manual", Source: cmd.SourceLocator}}
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

func canonicalWebKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Fragment = ""
	host := strings.ToLower(u.Hostname())
	if host == "youtu.be" {
		return "youtube:" + strings.Trim(u.Path, "/")
	}
	if host == "youtube.com" || host == "www.youtube.com" || host == "m.youtube.com" {
		if id := u.Query().Get("v"); id != "" {
			return "youtube:" + id
		}
		for _, prefix := range []string{"/shorts/", "/embed/"} {
			if strings.HasPrefix(u.Path, prefix) {
				return "youtube:" + strings.TrimPrefix(u.Path, prefix)
			}
		}
	}
	u.Host = strings.ToLower(u.Host)
	return u.String()
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
