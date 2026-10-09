// Package migrations embeds SQL migration files for use at runtime.
package migrations

import "embed"

// FS contains all embedded SQL migration files.
//
//go:embed *.sql
var FS embed.FS
