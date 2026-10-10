package domain

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// AggregateProjects uses observed repository identity. Names and identical
// commits do not establish identity; independently cloned repositories stay apart.
func AggregateProjects(projects []RegisteredProject) []ProjectSummary {
	board := []ProjectSummary{}
	byKey := map[string]int{}
	for _, p := range projects {
		key := p.ProjectKey
		if key == "" {
			key = "root:" + filepath.Clean(p.Root)
			if runtime.GOOS == "windows" {
				key = strings.ToLower(key)
			}
			if p.RepoID != "" && p.Source != "subproject" {
				key = "orca:" + p.RepoID
			}
		}
		id := fmt.Sprintf("project-%x", sha256.Sum256([]byte(key)))
		i, ok := byKey[id]
		if !ok {
			i = len(board)
			byKey[id] = i
			board = append(board, ProjectSummary{ID: id, Name: p.Name, Roots: []string{}, Observations: []RegisteredProject{}, Contributors: []ProjectContributor{}, Limitations: []string{}})
		}
		item := &board[i]
		duplicate := false
		for _, root := range item.Roots {
			if root == p.Root {
				duplicate = true
			}
		}
		if duplicate {
			continue
		}
		item.Roots = append(item.Roots, p.Root)
		item.Observations = append(item.Observations, p)
		limitations := []string{}
		if p.Git.Reason != "" && (p.Git.Status == "unknown" || p.Git.Status == "unavailable") {
			limitations = append(limitations, "git:"+p.Git.Reason)
		}
		for _, reason := range p.Limitations {
			if strings.HasPrefix(reason, "native_session_sample_bound:") {
				limitations = append(limitations, reason)
			}
		}
		for _, reason := range limitations {
			found := false
			for _, old := range item.Limitations {
				if old == reason {
					found = true
				}
			}
			if !found {
				item.Limitations = append(item.Limitations, reason)
			}
		}
		// Prefer the registered main checkout's name to an arbitrary worktree.
		if p.Source == "orca_registered" {
			item.Name = p.Name
		}
		if stamp := p.Activity.LastActivityAt; stamp != nil && (item.LastActivityAt == nil || stamp.After(*item.LastActivityAt)) {
			item.LastActivityAt = stamp
		}
		// Commit time is independently observed even if a later Git status query
		// fails. Session header creation remains separate from recorded activity.
		if stamp := p.Git.LastCommitAt; stamp != nil && (item.LastActivityAt == nil || stamp.After(*item.LastActivityAt)) {
			item.LastActivityAt = stamp
		}
		for _, c := range p.Contributors {
			found := false
			for _, old := range item.Contributors {
				if old.CLI == c.CLI && old.SessionID == c.SessionID && old.Source == c.Source {
					found = true
				}
			}
			if !found {
				item.Contributors = append(item.Contributors, c)
			}
			if stamp := c.ActivityAt; c.Source != "native_session_header" && stamp != nil && (item.LastActivityAt == nil || stamp.After(*item.LastActivityAt)) {
				item.LastActivityAt = stamp
			}
		}
	}
	for i := range board {
		sort.Strings(board[i].Roots)
	}
	sort.Slice(board, func(i, j int) bool { return board[i].ID < board[j].ID })
	return board
}
