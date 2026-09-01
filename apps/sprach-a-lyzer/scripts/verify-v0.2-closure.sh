#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

go test ./schemas/analysis \
  -run '^(TestGoContractsMatchJSONSchemaObjectShapes|TestEngineOutputAndDerivedTraceRespectContractShapes|TestResolverResultEnumsMatchCanonicalPolicyIDs)$' \
  -count=1

go test ./internal/httpapp \
  -run '^(TestAnalyzeUsesCoreEngine|TestPublicV02ResolverAndTraceContracts|TestPublicV02RoutesUseStrictSharedRequestContract)$' \
  -count=1

go test ./internal/golden \
  -run '^(TestVerticalSliceGolden|TestPropositionTraceGolden|TestConstructCompositionRuntimeGolden)$' \
  -count=1

go test ./internal/resolver \
  -run '^(TestContextPropositionGolden|TestRelationsAndScopeGoldenV03|TestResolverCatalogueSenseActivationAndReactivation)$' \
  -count=1

go test ./internal/engine \
  -run '^(TestRuntimeCatalogueActivationAndDeactivationSmoke|TestOntologyCompositionActivationAndDeactivationSmoke)$' \
  -count=1

go test ./internal/db \
  -run '^TestEmbeddedMigrationsReachRequiredSchemaVersion$' \
  -count=1

go test ./internal/version \
  -run '^(TestCoreMatchesReleaseManifest|TestHistoricalV010ReleaseManifestRemainsImmutable)$' \
  -count=1
