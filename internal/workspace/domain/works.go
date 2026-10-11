package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Works attitudes. These are the only three intents a human can record for a
// work; an unknown label never silently becomes one of them.
const (
	AttitudeContinue = "continue" // 想继续做
	AttitudePark     = "park"     // 暂时搁置
	AttitudeDone     = "done"     // 已经收尾
)

// Work handoff manual provenance.
const (
	HandoffSourceHuman     = "human"
	HandoffSourceGenerated = "generated"
)

// Work import provenance.
const (
	WorkSourceScan       = "scan"
	WorkSourceJSON       = "json"
	WorkSourceRegistered = "registered"
)

// WorkKind distinguishes a folder-rooted work from a single-file work. The
// file facts are verified the same way for both.
const (
	WorkKindFolder = "folder"
	WorkKindFile   = "file"
)

// WorkFileFact records the source path, recent mtime and file count that a
// bounded filesystem read actually observed. Verified is true only when the
// scanner read the directory itself; a JSON delivery alone never sets it, so
// the "从文件核实" badge stays honest about what was checked on this machine.
type WorkFileFact struct {
	SourcePath string     `json:"source_path"`
	Mtime      *time.Time `json:"mtime"`
	FileCount  int        `json:"file_count"`
	Verified   bool       `json:"verified"`
	Reason     string     `json:"reason,omitempty"`
}

// WorkHandoff is the per-work handoff manual (交接手册). Content is human
// text or a fixed generator's output; UpdatedAt records the last generate or
// update so a reader can see how fresh the manual is.
type WorkHandoff struct {
	Content   string     `json:"content"`
	UpdatedAt *time.Time `json:"updated_at"`
	Source    string     `json:"source"`
}

// Work is an archival album entry: a named creative work attributed to one AI
// platform, rooted at a local path. It is a finder entry, never an execution
// binding, an Agent permission or a space admission. Attitude, group, hidden
// and handoff are human metadata kept independently from replaceable file
// observations.
type Work struct {
	ID        string       `json:"id"`
	Platform  string       `json:"platform"`
	Name      string       `json:"name"`
	Root      string       `json:"root"`
	Kind      string       `json:"kind"`
	Source    string       `json:"source"`
	Attitude  string       `json:"attitude"`
	Group     string       `json:"group"`
	Hidden    bool         `json:"hidden"`
	Handoff   WorkHandoff  `json:"handoff"`
	Files     WorkFileFact `json:"files"`
	UpdatedAt *time.Time   `json:"updated_at"`
	Revision  int          `json:"revision"`
}

// WorkDiscoveryFailure keeps one bounded root that could not be observed, so a
// partial scan never reports an empty album as complete.
type WorkDiscoveryFailure struct {
	Root   string `json:"root"`
	Reason string `json:"reason"`
}

// WorksSnapshot is the durable album projection plus per-platform statistics.
type WorksSnapshot struct {
	Works      []Work                 `json:"works"`
	Platforms  []PlatformStat         `json:"platforms"`
	Status     string                 `json:"status"`
	ObservedAt *time.Time             `json:"observed_at"`
	Failures   []WorkDiscoveryFailure `json:"failures"`
}

// PlatformStat is the per-platform aggregation shown at the top of the album:
// work count, folder count, handed-over count (a handoff manual exists) and
// recently-active count (mtime inside the recent window).
type PlatformStat struct {
	Platform   string `json:"platform"`
	Works      int    `json:"works"`
	Folders    int    `json:"folders"`
	HandedOver int    `json:"handed_over"`
	Recent     int    `json:"recent"`
}

// WorkIdentity is the stable key for one platform + root pair. Two independent
// folders never share an ID; two imports of the same folder always do.
func WorkIdentity(platform, root string) string {
	platform = strings.TrimSpace(platform)
	root = filepath.Clean(root)
	if runtime.GOOS == "windows" {
		root = strings.ToLower(root)
	}
	sum := sha256.Sum256([]byte(platform + "\x00" + root))
	return fmt.Sprintf("work-%x", sum)
}

// ValidAttitude accepts only the three supported intents.
func ValidAttitude(attitude string) bool {
	return attitude == AttitudeContinue || attitude == AttitudePark || attitude == AttitudeDone
}

// ValidWorkPlatform bounds the free-form platform label so a hostile import
// cannot grow the stored record without limit. The vocabulary stays open:
// platforms are defined by what the bounded survey actually found.
func ValidWorkPlatform(platform string) bool {
	return len(platform) >= 1 && len(platform) <= 64
}

// ValidWorkName bounds a work name.
func ValidWorkName(name string) bool {
	return len(name) >= 1 && len(name) <= 200
}

// ValidWorkGroup bounds the human grouping label.
func ValidWorkGroup(group string) bool {
	return len(group) <= 200
}

// HandedOver reports whether a handoff manual exists for the work. A manual
// with content is what "接力过" counts; an empty manual is not a handoff.
func HandedOver(handoff WorkHandoff) bool {
	return strings.TrimSpace(handoff.Content) != ""
}

// RecentlyActive reports whether the observed file mtime falls inside the
// recent window ending at now. An unverified work is never "recent".
func (f WorkFileFact) RecentlyActive(now time.Time, window time.Duration) bool {
	if !f.Verified || f.Mtime == nil {
		return false
	}
	return f.Mtime.After(now.Add(-window))
}

// AggregateWorksStats produces per-platform counts, sorted by platform label so
// the projection is deterministic.
func AggregateWorksStats(works []Work, now time.Time, recentWindow time.Duration) []PlatformStat {
	if recentWindow <= 0 {
		recentWindow = 30 * 24 * time.Hour
	}
	index := map[string]int{}
	result := []PlatformStat{}
	for _, work := range works {
		i, ok := index[work.Platform]
		if !ok {
			i = len(result)
			index[work.Platform] = i
			result = append(result, PlatformStat{Platform: work.Platform})
		}
		stat := &result[i]
		stat.Works++
		if work.Kind == WorkKindFolder {
			stat.Folders++
		}
		if HandedOver(work.Handoff) {
			stat.HandedOver++
		}
		if work.Files.RecentlyActive(now, recentWindow) {
			stat.Recent++
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Platform < result[j].Platform })
	return result
}

// WorkFilter narrows the album view. Zero fields mean no filtering; Hidden
// selects whether hidden works are included at all.
type WorkFilter struct {
	Platform      string
	Group         string
	Attitude      string
	IncludeHidden bool
}

// FilterWorks applies the view filter, preserving the input order.
func FilterWorks(works []Work, filter WorkFilter) []Work {
	result := make([]Work, 0, len(works))
	for _, work := range works {
		if work.Hidden && !filter.IncludeHidden {
			continue
		}
		if filter.Platform != "" && work.Platform != filter.Platform {
			continue
		}
		if filter.Group != "" && work.Group != filter.Group {
			continue
		}
		if filter.Attitude != "" && work.Attitude != filter.Attitude {
			continue
		}
		result = append(result, work)
	}
	return result
}

// SortWorks orders the album by a fixed key. "recent" sorts by observed file
// mtime, "updated" by the last human/record update, and anything else by name.
func SortWorks(works []Work, key string) []Work {
	result := append([]Work{}, works...)
	sort.SliceStable(result, func(i, j int) bool {
		switch key {
		case "recent":
			return fileMtime(result[i]).After(fileMtime(result[j]))
		case "updated":
			return updatedAt(result[i]).After(updatedAt(result[j]))
		default:
			return result[i].Name < result[j].Name
		}
	})
	return result
}

func fileMtime(work Work) time.Time {
	if work.Files.Mtime != nil {
		return *work.Files.Mtime
	}
	return time.Time{}
}

func updatedAt(work Work) time.Time {
	if work.UpdatedAt != nil {
		return *work.UpdatedAt
	}
	return time.Time{}
}

// WorkEvent is the raw works-activity event emitted to the chain-1 event
// source. W1 normalizes it into the unified Event (§4.1): Source carries the
// actor (manual|mcp), Kind carries the activity, Scope the work root, Topic a
// platform-derived seed, and Payload the audit body. Tokens are absent here
// because works activities are not token-metered.
type WorkEvent struct {
	WorkID  string          `json:"work_id"`
	Kind    string          `json:"kind"`
	Source  string          `json:"source"`
	Scope   string          `json:"scope"`
	Topic   string          `json:"topic"`
	At      time.Time       `json:"at"`
	Payload json.RawMessage `json:"payload"`
}

// Works activity kinds. These are the works layer's own vocabulary; the W1
// events adapter maps them onto the §4.1 kind set in the handoff contract.
const (
	WorkEventImported    = "work_imported"
	WorkEventAttitudeSet = "work_attitude_set"
	WorkEventHandoffSet  = "work_handoff_set"
	WorkEventHidden      = "work_hidden"
	WorkEventVerified    = "work_verified"
	WorkEventSynced      = "work_synced"
)

// WorkEventSource is the chain-1 event source label for works activity.
const WorkEventSourceManual = "manual"

// WorkHumanMetadata holds the durable human annotations for one work. It is
// stored independently from the replaceable scan observation, so a refresh or
// re-import never overwrites attitude, group, hidden or the handoff manual.
type WorkHumanMetadata struct {
	Attitude  string      `json:"attitude"`
	Group     string      `json:"group"`
	Hidden    bool        `json:"hidden"`
	Handoff   WorkHandoff `json:"handoff"`
	UpdatedAt *time.Time  `json:"updated_at"`
	Revision  int         `json:"revision"`
}

// Apply merges human metadata into a work observation, producing the album view.
func (m WorkHumanMetadata) Apply(work *Work) {
	work.Attitude = m.Attitude
	work.Group = m.Group
	work.Hidden = m.Hidden
	work.Handoff = m.Handoff
	work.UpdatedAt = m.UpdatedAt
	work.Revision = m.Revision
}

// WorkDeliveryEntry is one entry in an "AI 投递 JSON" file. Path is the local
// root the work points at; the platform label is required so the album can
// attribute the work without guessing.
type WorkDeliveryEntry struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Attitude string `json:"attitude"`
	Group    string `json:"group"`
	Handoff  string `json:"handoff"`
}

// WorkDelivery is the bounded "AI 投递 JSON" envelope.
type WorkDelivery struct {
	Works []WorkDeliveryEntry `json:"works"`
}

// NormalizeWorkKind maps an empty or unknown kind to the folder default.
func NormalizeWorkKind(kind string) string {
	if kind == WorkKindFile {
		return WorkKindFile
	}
	return WorkKindFolder
}
