# Sprach-A-Lyzer – Foundation Rule Migration v0.1

- **Status:** APPROVED
- **Version:** 0.1
- **Stand:** 25. August 2026
- **Owner:** Product & Engineering
- **Basiert auf:** Sprint 0A Canonical Contracts v0.1

## Ergebnis

Die sechs Foundation-Regeln wurden ohne fachliche Neukalibrierung aus freien
JSON-Fragmenten in einen strikt typisierten Regelkatalog migriert:

- Foundation und Rule Set: `0.2`
- Rule Contract: `0.3`
- Policy Registry: `0.2`
- Golden Suite: unverändert `0.2`

Rule v0.2 und Policy Registry v0.1 bleiben unverändert. Die Foundation benötigt
zusätzlich eine explizite Sinnauswahl und einen reinen Resonanzhinweis. Deshalb
führt Rule v0.3 additiv `SELECT_SENSE` und `ADD_RESONANCE_HINT` ein, statt den
bereits veröffentlichten v0.2-Vertrag still zu verändern.

## Kanonische Artefakte

- [`sprach-a-lyzer_foundation_v0.2.json`](../../data/seed/sprach-a-lyzer_foundation_v0.2.json)
- [`sprach-a-lyzer_rule_v0.3.json`](../../schemas/rules/sprach-a-lyzer_rule_v0.3.json)
- [`sprach-a-lyzer_policy-registry_v0.2.json`](../../data/seed/sprach-a-lyzer_policy-registry_v0.2.json)
- [`sprach-a-lyzer_policy-registry_v0.2.json`](../../schemas/rules/sprach-a-lyzer_policy-registry_v0.2.json)

## Guardrails

`ADD_RESONANCE_HINT` verlangt `semantic_score: false`. Damit können Homophone
einen erklärenden Hinweis, aber weder eine Sinnübertragung noch einen
Dimensionsbeitrag erzeugen. `SELECT_SENSE` wählt nur einen expliziten Sinn; ein
Score entsteht ausschließlich durch eine separate registrierte Score-Aktion.

Die Seed-Grenze verwirft unbekannte Felder, Legacy-Aktionsformen, ungültige
Conditions, nicht registrierte Aktionen und unzulässige Wertebereiche. Die
Datenbank speichert Vertragsversion, Evidenzklasse und stabile Source Keys.

## Nicht Bestandteil

Die deterministische Engine liest den Katalog noch nicht als Laufzeitprogramm.
Ihre Golden-gesicherte Semantik bleibt unverändert. Die Engine-Anbindung und die
fachliche Kalibrierung sind getrennte Folgeschritte.
