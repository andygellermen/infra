package migrations

import "embed"

// Files contains the immutable PostgreSQL migrations.
//
//go:embed *.sql
var Files embed.FS
