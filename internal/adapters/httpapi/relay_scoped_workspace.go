package httpapi

import (
	"net/http"
	"strings"
)

// 此分支仅放行已鉴权的项目 caller；领域服务逐项重核目标项目和 grant。
func relayScopedWorkspaceRoute(r *http.Request, projectID string) bool {
	if projectID == "" {
		return false
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "v1" {
		return false
	}
	if len(parts) >= 5 && parts[2] == "local-projects" && parts[3] == projectID && parts[4] == "judgments" {
		return len(parts) == 5 && r.Method == http.MethodPost || len(parts) == 6 && r.Method == http.MethodGet || len(parts) == 7 && parts[6] == "feedback" && r.Method == http.MethodGet
	}
	if len(parts) == 6 && parts[2] == "local-projects" && parts[3] == projectID && parts[4] == "repository-context" {
		switch parts[5] {
		case "map":
			return r.Method == http.MethodGet
		case "search", "impact":
			return r.Method == http.MethodPost
		}
		return false
	}
	if len(parts) == 3 {
		switch parts[2] {
		case "missions":
			return r.Method == http.MethodGet
		case "goals":
			return r.Method == http.MethodGet
		case "proposals":
			return r.Method == http.MethodGet || r.Method == http.MethodPost
		case "context-packets", "session-bindings":
			return r.Method == http.MethodPost
		}
	}
	if len(parts) == 4 && parts[3] != "" {
		switch parts[2] {
		case "missions", "artifacts":
			return r.Method == http.MethodGet
		case "proposals":
			return r.Method == http.MethodGet || r.Method == http.MethodPut
		case "goals", "approvals", "context-packets":
			return r.Method == http.MethodGet
		}
	}
	if len(parts) == 5 && parts[3] != "" {
		switch parts[2] {
		case "missions":
			return parts[4] == "work-items" && (r.Method == http.MethodGet || r.Method == http.MethodPost) || parts[4] == "timeline" && r.Method == http.MethodGet
		case "work-items":
			return (parts[4] == "claims" || parts[4] == "artifacts") && r.Method == http.MethodPost
		case "artifacts":
			return parts[4] == "diff" && r.Method == http.MethodGet
		}
	}
	return len(parts) == 5 && parts[2] == "proposals" && parts[3] != "" && parts[4] == "submit" && r.Method == http.MethodPost
}
