package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode"
)

type OpenVikingSubmission struct {
	ID           string          `json:"id"`
	Version      int             `json:"version"`
	RequestID    string          `json:"request_id"`
	SpaceID      string          `json:"space_id"`
	MaterialID   string          `json:"material_id"`
	Revision     int             `json:"revision"`
	Locator      string          `json:"locator"`
	Title        string          `json:"title"`
	Destination  string          `json:"destination"`
	BaseURL      string          `json:"base_url"`
	Account      string          `json:"account"`
	User         string          `json:"user"`
	SpaceVersion int             `json:"space_version"`
	ActorID      string          `json:"actor_id"`
	ApprovedAt   time.Time       `json:"approved_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Status       string          `json:"status"`
	Stage        string          `json:"stage"`
	TempFileID   string          `json:"temp_file_id"`
	TaskID       string          `json:"task_id"`
	RootURI      string          `json:"root_uri"`
	Error        string          `json:"error"`
	TaskSnapshot json.RawMessage `json:"task_snapshot"`
}

// 规范 URI 保留逐段身份，拒绝会改变目录边界的等价写法。
func CanonicalVikingURI(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "viking" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(raw, "#") || u.Opaque != "" || u.Host == "" || u.Port() != "" {
		return "", fmt.Errorf("invalid canonical Viking URI")
	}
	if u.Host != "~" && u.Host != "user" && u.Host != "resources" {
		return "", fmt.Errorf("unsupported Viking namespace")
	}
	clean := path.Clean(u.Path)
	if u.Path != clean || clean == "." || strings.HasSuffix(raw, "/") || strings.ContainsAny(u.Path, "\r\n\t") {
		return "", fmt.Errorf("noncanonical Viking path")
	}
	for _, encoded := range strings.Split(strings.TrimPrefix(u.EscapedPath(), "/"), "/") {
		segment, err := url.PathUnescape(encoded)
		if err != nil || segment == "" || segment == "." || segment == ".." || strings.ContainsAny(segment, "/\\%{}") || strings.IndexFunc(segment, unicode.IsControl) >= 0 {
			return "", fmt.Errorf("invalid Viking path segment")
		}
	}
	u.RawPath = ""
	return u.String(), nil
}

func VikingWithin(raw, root string) error {
	canonicalRoot, err := CanonicalVikingURI(root)
	if err != nil {
		return err
	}
	canonicalRaw, err := CanonicalVikingURI(raw)
	if err != nil {
		return err
	}
	if canonicalRaw != canonicalRoot && !strings.HasPrefix(canonicalRaw, canonicalRoot+"/") {
		return fmt.Errorf("Viking URI is outside the authorized project root")
	}
	return nil
}

// 上游资源首次固定后只能核实同一规范 URI。
func BindVikingResource(projectRoot, boundURI, observedURI string) (string, error) {
	root, err := CanonicalVikingURI(projectRoot)
	if err != nil {
		return "", err
	}
	strict := func(raw string) (string, error) {
		canonical, err := CanonicalVikingURI(raw)
		if err != nil {
			return "", err
		}
		if canonical == root {
			return "", fmt.Errorf("resource must be strictly below the project root")
		}
		if err := VikingWithin(canonical, root); err != nil {
			return "", err
		}
		return canonical, nil
	}
	bound := ""
	if boundURI != "" {
		bound, err = strict(boundURI)
		if err != nil {
			return "", err
		}
	}
	if observedURI == "" {
		return bound, nil
	}
	observed, err := strict(observedURI)
	if err != nil {
		return "", err
	}
	if bound != "" && bound != observed {
		return "", fmt.Errorf("task resource differs from the fixed submission resource")
	}
	return observed, nil
}
