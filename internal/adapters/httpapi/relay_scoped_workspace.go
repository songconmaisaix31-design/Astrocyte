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
	if len(parts) == 3 {
		switch parts[2] {
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
		case "proposals":
			return r.Method == http.MethodGet || r.Method == http.MethodPut
		case "goals", "approvals", "context-packets":
			return r.Method == http.MethodGet
		}
	}
	return len(parts) == 5 && parts[2] == "proposals" && parts[3] != "" && parts[4] == "submit" && r.Method == http.MethodPost
}
