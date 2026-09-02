#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Preserve every v0.3 contract, golden and runtime guarantee.
bash ./scripts/verify-v0.3-closure.sh

go test ./schemas/questions \
  -run '^(TestQuestionSchemasAreStrictAndDoNotContainLegacyDimension|TestQuestionRuntimeTopLevelShapesMatchSchemas|TestApprovedV04RenderingCatalogueMatchesV02Shape)$' \
  -count=1

go test ./internal/questions \
  -run '^(TestQuestionRenderingGoldenV01|TestRenderingCatalogueHasCompleteProfileIsolation|TestRenderingPreservesCanonicalCoreAndQuestionSelection|TestCorporateAndPrivateProfilesPreserveCoreScores|TestRenderingActivationDeactivationAndFallbackSmoke|TestRenderingFailsClosedOnProviderOrIntentDrift)$' \
  -count=1

go test ./internal/httpapp \
  -run '^(TestPublicV04ProfileRenderingContract|TestPublicV04RenderingRequestIsStrict)$' \
  -count=1

go test ./internal/presentation \
  -run '^(TestCorporateBundleNeverReturnsCanonicalKey|TestFoundationPresentationBundlesStayProfileIsolated)$' \
  -count=1

go test ./internal/version \
  -run '^(TestCoreMatchesReleaseManifest|TestHistoricalV030ReleaseManifestRemainsImmutable|TestHistoricalV020ReleaseManifestRemainsImmutable|TestHistoricalV010ReleaseManifestRemainsImmutable)$' \
  -count=1
