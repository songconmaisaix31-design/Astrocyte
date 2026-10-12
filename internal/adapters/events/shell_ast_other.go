//go:build !windows

package events

// Non-Windows commands keep the deterministic lexical shell grammar; no
// PowerShell environment is assumed. Parse failures remain unresolved.
func parseShellAST(code string) (shellASTResult, bool) { return shellASTResult{}, false }
