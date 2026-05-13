package migrations

import "embed"

// Files contains all SQL migrations for easy-event-planner.
//
//go:embed *.sql
var Files embed.FS
