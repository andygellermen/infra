// Package assets exposes immutable build-time assets used by standalone tools.
package assets

import _ "embed"

// FoundationV02 is the canonical standalone Foundation catalogue.
//
//go:embed data/seed/sprach-a-lyzer_foundation_v0.2.json
var FoundationV02 []byte
