package events

import (
	"context"
	"encoding/json"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// These observations remain local. Only Topic/Status belong in safe audits.
type RepositoryObservation struct {
	CWD       string `json:"cwd"`
	Root      string `json:"root"`
	CommonDir string `json:"common_dir"`
	Remote    string `json:"remote"`
	Topic     string `json:"topic"`
	Status    string `json:"status"`
}

func repositoryIdentity(value string) string {
	value = strings.TrimSuffix(strings.TrimRight(strings.ReplaceAll(value, "\\", "/"), "/"), ".git")
	if u, err := url.Parse(value); err == nil && u.Host != "" {
		value = u.Path
	}
	return strings.ToLower(filepath.Base(value))
}

func repositoryTopic(identity string) string {
	switch identity {
	case "astrocyte":
		return "Astrocyte/Attention Engine"
	case "morphogenesis":
		return "Morphogenesis"
	case "maimai-monster":
		return "McPet"
	}
	return ""
}

// ObserveRepository executes only fixed, read-only Git introspection argv at
// the observed cwd. It never executes historical tool content or captures log.
func ObserveRepository(cwd string) RepositoryObservation {
	o := RepositoryObservation{CWD: cwd, Status: "not_repository"}
	info, err := os.Stat(cwd)
	if err != nil || !info.IsDir() {
		o.Status = "missing_cwd"
		return o
	}
	run := func(args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", cwd}, args...)...)
		b, e := cmd.Output()
		return strings.TrimSpace(string(b)), e
	}
	o.Root, err = run("rev-parse", "--show-toplevel")
	if err != nil {
		return o
	}
	o.CommonDir, err = run("rev-parse", "--git-common-dir")
	if err != nil {
		o.Status = "common_dir_unavailable"
		return o
	}
	if !filepath.IsAbs(o.CommonDir) {
		o.CommonDir = filepath.Clean(filepath.Join(cwd, o.CommonDir))
	}
	o.Remote, _ = run("config", "--get", "remote.origin.url")
	common := o.CommonDir
	if strings.EqualFold(filepath.Base(common), ".git") {
		common = filepath.Dir(common)
	}
	a, b := repositoryTopic(repositoryIdentity(common)), repositoryTopic(repositoryIdentity(o.Remote))
	if a != "" && b != "" && a != b {
		o.Status = "repository_identity_conflict"
		return o
	}
	o.Topic = a
	if o.Topic == "" {
		o.Topic = b
	}
	o.Status = "observed_unrecognized"
	if o.Topic != "" {
		o.Status = "observed_recognized"
	}
	return o
}

// Scope classification uses observed repository identity only. Keyword topic
// associations on Git facts remain the separate existing ClassifyGitTopics API.
type NativeSemanticMapper struct {
	repositories map[string]RepositoryObservation
}

func NewNativeSemanticMapper() *NativeSemanticMapper {
	return &NativeSemanticMapper{repositories: map[string]RepositoryObservation{}}
}

// A mapper instance freezes read-only repository observation once per cwd.
// Use the same instance for the full input; do not reinterpret source spelling.
func (m *NativeSemanticMapper) Repository(cwd string) RepositoryObservation {
	if o, ok := m.repositories[cwd]; ok {
		return o
	}
	o := ObserveRepository(cwd)
	m.repositories[cwd] = o
	return o
}

func (m *NativeSemanticMapper) MapTool(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) ([]app.StoredEvent, []SemanticTool, error) {
	rows, semantic, err := mapNativeSemanticTools(source, origin, project, raw, receivedAt, association, block)
	if err != nil {
		return nil, nil, err
	}
	for i := range rows {
		if rows[i].Event != nil {
			rows[i].Event.Topic = m.Repository(rows[i].Observation.Scope).Topic
		}
	}
	return rows, semantic, nil
}

func (m *NativeSemanticMapper) MapGit(project string, raw json.RawMessage, receivedAt int64) (app.StoredEvent, error) {
	row, err := MapGitLogRecord(project, raw, receivedAt)
	if err == nil && row.Event != nil {
		row.Event.Topic = m.Repository(row.Observation.Scope).Topic
	}
	return row, err
}
