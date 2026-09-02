#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

bash ./scripts/verify-v0.4-closure.sh

go test ./schemas/imports -run '^TestManagedImportSchemasAreStrictAndMatchRuntimeShapes$' -count=1
go test ./internal/managedimport -run '^(TestManagedImportJSONCommitRollbackAndAudit|TestValidateOnlyCSVMappingCompatibilityAndDuplicateDetection|TestConflictsReferencesCyclesAndGoldenRegressionBlockCommit|TestXLSXFirstSheetParsing|TestManualCriticalResolutionRequiresReviewer|TestUnsupportedSyncAndInvalidSourceFailClosed|TestPostgresManagedImportCommitRollbackAndImmutableAudit)$' -count=1
go test ./internal/httpapp -run '^(TestPublicV05ManagedImportLifecycle|TestPublicV05ManagedImportIsStrictAndRoleGuarded)$' -count=1
go test ./internal/db -run '^TestEmbeddedMigrationsReachRequiredSchemaVersion$' -count=1
go test ./internal/version -run '^(TestCoreMatchesReleaseManifest|TestHistoricalV040ReleaseManifestRemainsImmutable|TestHistoricalV030ReleaseManifestRemainsImmutable|TestHistoricalV020ReleaseManifestRemainsImmutable|TestHistoricalV010ReleaseManifestRemainsImmutable)$' -count=1
