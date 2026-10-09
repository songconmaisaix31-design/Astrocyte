// Package importers adapts existing exports and official arXiv originals.
// It never downloads video or starts transcription tools.
package importers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
)

// Export is local adapter data, mapped to application ports after parsing.
type Export struct {
	URL, Title, Text, Summary, SourceRevision string
	Segments                                  []Segment
	Original                                  []byte
}

type Segment struct {
	StartMS float64  `json:"startMs"`
	EndMS   *float64 `json:"endMs"`
	Text    string   `json:"text"`
}

// ParseSummarizeJSON accepts the installed 0.21.8 CLI output (extract or
// summary). Original JSON is retained; unknown usage/cost is not filled in.
func ParseSummarizeJSON(raw []byte) (Export, error) {
	var v struct {
		Input struct {
			URL string `json:"url"`
		} `json:"input"`
		Extracted struct {
			URL                string `json:"url"`
			Title              string `json:"title"`
			Content            string `json:"content"`
			TranscriptSegments []struct {
				StartMS *float64 `json:"startMs"`
				EndMS   *float64 `json:"endMs"`
				Text    string   `json:"text"`
			} `json:"transcriptSegments"`
		} `json:"extracted"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return Export{}, fmt.Errorf("summarize export JSON: %w", err)
	}
	locator := v.Input.URL
	if locator == "" {
		locator = v.Extracted.URL
	}
	if err := validateWebURL(locator); err != nil {
		return Export{}, err
	}
	if v.Input.URL != "" && v.Extracted.URL != "" && v.Input.URL != v.Extracted.URL {
		// Redirects are legitimate. Both values remain in the immutable export;
		// the explicit input URL continues to identify the requested source.
		if err := validateWebURL(v.Extracted.URL); err != nil {
			return Export{}, err
		}
	}
	text := v.Extracted.Content
	if strings.TrimSpace(text) == "" {
		text = v.Summary
	}
	if strings.TrimSpace(text) == "" {
		return Export{}, errors.New("summarize export has no readable content")
	}
	r := Export{URL: locator, Title: v.Extracted.Title, Text: text, Summary: v.Summary, Original: append([]byte(nil), raw...)}
	for _, seg := range v.Extracted.TranscriptSegments {
		if seg.StartMS == nil || math.IsNaN(*seg.StartMS) || math.IsInf(*seg.StartMS, 0) || *seg.StartMS < 0 {
			return Export{}, errors.New("invalid transcript startMs")
		}
		if seg.EndMS != nil && (math.IsNaN(*seg.EndMS) || math.IsInf(*seg.EndMS, 0) || *seg.EndMS < *seg.StartMS) {
			return Export{}, errors.New("invalid transcript endMs")
		}
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		r.Segments = append(r.Segments, Segment{StartMS: *seg.StartMS, EndMS: seg.EndMS, Text: seg.Text})
	}
	return r, nil
}

// Plain markdown exports require an explicit source URL. No timestamps are
// inferred from prose or total video duration.
func ParseSummarizeMarkdown(locator, title string, raw []byte) (Export, error) {
	if err := validateWebURL(locator); err != nil {
		return Export{}, err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return Export{}, errors.New("empty summarize markdown")
	}
	return Export{URL: locator, Title: title, Text: string(raw), Original: append([]byte(nil), raw...)}, nil
}

func validateWebURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
		return errors.New("source requires an HTTP(S) URL without credentials")
	}
	return nil
}
