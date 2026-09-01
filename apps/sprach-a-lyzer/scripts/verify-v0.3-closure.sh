#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Preserve every v0.2 contract, golden and runtime guarantee.
bash ./scripts/verify-v0.2-closure.sh

go test ./schemas/questions \
  -run '^(TestCanonicalQuestionAndRenderingStaySeparated|TestQuestionSchemasAreStrictAndDoNotContainLegacyDimension|TestQuestionRuntimeTopLevelShapesMatchSchemas|TestApprovedCatalogueUsesCanonicalAndRenderingShapes)$' \
  -count=1

go test ./internal/questions \
  -run '^(TestQuestionAnswerRuntimeGoldenV02|TestAdaptiveSelectionOffersFiveToEightQuestions|TestSessionSupportsHedgedC2AndTemporalC3WithoutC4|TestQuestionCatalogueActivationAndDeactivationSmoke|TestQuestionCatalogueFailsClosed|TestHistoricalQuestionCorpusMismatchIsExplicitlyQuarantined|TestUnknownQuestionAndEmptyAnswerAreRejected|TestQuestionRuntimeRejectsInvalidProfileAndOversizedAnswer|TestNoQuestionOutputContainsTraitOrCausalEffectClaims)$' \
  -count=1

go test ./internal/httpapp \
  -run '^(TestPublicV03QuestionAnswerContracts|TestPublicV03QuestionRoutesAreStrictAndFailClosed)$' \
  -count=1

go test ./internal/version \
  -run '^(TestCoreMatchesReleaseManifest|TestHistoricalV020ReleaseManifestRemainsImmutable|TestHistoricalV010ReleaseManifestRemainsImmutable)$' \
  -count=1
