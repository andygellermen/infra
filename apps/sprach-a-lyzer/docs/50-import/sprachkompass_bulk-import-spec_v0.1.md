# Sprachkompass Bulk Import v0.1

## Empfohlene Formate
- Pflege/Review: XLSX oder Google Sheets
- Kanonischer Austausch/API: JSON
- Große flache Bulk-Jobs: CSV pro Entity
- DB-Import: transaktionaler UPSERT via Natural Keys

## Natural Keys
Keine UUIDs im redaktionellen Sheet. Nutze stabile Keys wie `lx_muessen`, `sn_muessen_internal_pressure`, `ph_ich_muss`.

## Importablauf
1. Schema-Version validieren
2. Tabellen/JSON syntaktisch prüfen
3. Referenzen auflösen
4. Guardrails prüfen
5. Transaktionaler Import in Reihenfolge Lexemes → Senses → Phrases → Relations → Contributions → Sources
6. Smoke-/Golden-Tests
7. Commit oder Rollback

## Google Sheets
Für den MVP: Google Sheet pflegen → XLSX oder CSV exportieren → Admin-Import hochladen.
Später: Google Apps Script/Sync-Service → JSON → `POST /api/v1/admin/imports`.

## API-Idee
- `POST /api/v1/admin/imports/validate`
- `POST /api/v1/admin/imports`
- `GET /api/v1/admin/imports/{batch_id}`

Empfehlung: zunächst immer `VALIDATE_ONLY`, danach bewusster Commit/Publish.

## Wichtig
PresentationBundles Private/Corporate separat importieren. Sie gehören nicht in die fachlichen Basisdaten.
