// Package assets exposes immutable build-time assets used by standalone tools.
package assets

import _ "embed"

// FoundationV03 is the canonical standalone Foundation catalogue.
//
//go:embed data/seed/sprach-a-lyzer_foundation_v0.3.json
var FoundationV03 []byte
