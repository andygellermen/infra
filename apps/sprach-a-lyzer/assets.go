// Package assets exposes immutable build-time assets used by standalone tools.
package assets

import _ "embed"

// FoundationV03 is the canonical standalone Foundation catalogue.
//
//go:embed data/seed/sprach-a-lyzer_foundation_v0.3.json
var FoundationV03 []byte

// ResolverCatalogueV01 is the canonical deterministic resolver catalogue.
//
//go:embed data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json
var ResolverCatalogueV01 []byte

// ConstructOntologyV02 is the canonical non-scoring construct catalogue.
//
//go:embed data/seed/sprach-a-lyzer_construct-ontology_v0.2.json
var ConstructOntologyV02 []byte
