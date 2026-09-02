#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

bash ./scripts/verify-v0.5-closure.sh

go test ./schemas/experience -run '^TestMVPExperienceSchemasAreStrictAndMatchRuntimeShapes$' -count=1
go test ./internal/experience -run '^(TestMVPExperienceGoldenV01|TestMVPExperienceComposesNoAIPrivateResultWithoutPersistence|TestMVPExperienceKeepsProfilesIsolatedAndSupportsEasyQuestions|TestMVPExperienceFailsClosed)$' -count=1
go test ./internal/httpapp -run '^(TestPublicV06MVPExperienceIsTransientNoAIAndExplainable|TestPublicV06ExperienceContractIsStrictAndFailClosed|TestMVPProductAndReadOnlyAdminShellAreEmbedded)$' -count=1
go test ./internal/architecture -run '^(TestFeatureModulesDoNotImportEachOther|TestAnalysisInternalsAreOnlyUsedThroughFacade)$' -count=1
go test ./internal/version -run '^(TestCoreMatchesReleaseManifest|TestHistoricalV050ReleaseManifestRemainsImmutable|TestHistoricalV040ReleaseManifestRemainsImmutable|TestHistoricalV030ReleaseManifestRemainsImmutable|TestHistoricalV020ReleaseManifestRemainsImmutable|TestHistoricalV010ReleaseManifestRemainsImmutable)$' -count=1
