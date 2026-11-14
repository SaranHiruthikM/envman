// Package migrations ships the SQL files to the server binary via go:embed so
// the server has no filesystem dependency on the `migrations/` directory.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
