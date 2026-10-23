package agents

import (
	"context"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectfiles"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 旧产品入口保留函数合同，文件权限只由 projectfiles 实现。
func CanonicalRoot(root string) (string, error) { return projectfiles.CanonicalRoot(root) }
func ValidateProjectPath(root, relative string) (string, error) {
	return projectfiles.ValidateProjectPath(root, relative)
}
func ReadProjectFiles(ctx context.Context, project domain.LocalProject, paths []string) ([]domain.ContextFile, error) {
	return projectfiles.ReadProjectFiles(ctx, project, paths)
}
func DiscoverProjects(ctx context.Context, root string) ([]domain.ProjectCandidate, error) {
	return projectfiles.DiscoverProjects(ctx, root)
}
func scopeError(message string) error        { return projectfiles.ScopeError(message) }
func inside(root, path string) bool          { return projectfiles.Inside(root, path) }
func hasRedirectedAncestor(path string) bool { return projectfiles.HasRedirectedAncestor(path) }
func secretComponent(name string) bool       { return projectfiles.SecretComponent(name) }
func pathSpelling(path string) string        { return projectfiles.PathSpelling(path) }
func pathIsRedirect(path string) bool        { return projectfiles.PathIsRedirect(path) }
