package agents

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// RelayOwnedRoots reuses existing project canonicalization and redirect checks.
// It reads directory identity only, never session files or private home roots.
type RelayOwnedRoots struct{}

var _ app.RelayOwnedRoots = RelayOwnedRoots{}

func (RelayOwnedRoots) CanonicalRoot(root string) (string, error) {
	if hasRedirectedAncestor(root) {
		return "", scopeError("registered relay root is redirected")
	}
	return CanonicalRoot(root)
}
