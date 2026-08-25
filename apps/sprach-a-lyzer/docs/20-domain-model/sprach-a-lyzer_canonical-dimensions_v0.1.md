# Sprach-A-Lyzer – Canonical Dimensions v0.1

**Status:** APPROVED  
**Version:** 0.1  
**Stand:** 24. August 2026  
**Owner:** Product & Engineering

## Kanonischer Vertrag

Der Core besitzt exakt sechs Dimensionen in stabiler Reihenfolge:

| ID | Positiver Anzeigepol | Gegenpol | Fachlicher Fokus |
|---|---|---|---|
| `AGENCY` | Wirksamkeit | Ohnmacht | sprachlich sichtbarer Handlungs- und Einflussraum |
| `CONNECTION` | Verbindung | Trennung | sprachlich sichtbare Bezogenheit |
| `APPRECIATION` | Wertschätzung | Abwertung | würdigende oder abwertende Evidenz |
| `CLARITY` | Klarheit | Unklarheit | Konkretheit und strukturelle Bestimmtheit |
| `VOLITION` | Freier Wille / Gestaltungsspielraum | Zwang | sprachlich sichtbare Wahl, Zustimmung und Selbstbestimmung |
| `OPENNESS` | Offenheit | Begrenzung | sprachlich sichtbarer Möglichkeits- und Lernraum |

Die IDs sind technische, profilunabhängige Schlüssel. Sichtbare Bezeichnungen
kommen aus dem jeweiligen Private- oder Corporate-Präsentationsbundle.

## Schutzregel

Dimensionen bewerten Evidenz im analysierten Text, niemals Eigenschaften eines
Menschen. Fehlende Evidenz erzeugt keinen neutralen 50-Punkte-Wert, sondern
`NOT_ASSESSABLE` mit `score: null`.

## Legacy-Kompatibilität

Einziger zulässiger Legacy-Alias:

```text
FREE_WILL → VOLITION
```

Die Kompatibilitätsschicht akzeptiert den Alias bei:

- typisierten JSON-Feldern und JSON-Map-Schlüsseln
- Dimensionslisten
- kommaseparierten CSV-Feldern
- Richtungsausdrücken wie `AGENCY:+;FREE_WILL:-`
- älteren Construct-Feldern, in denen `FREE_WILL` als Dimensionskonstrukt diente

Nach der Importgrenze existiert ausschließlich `VOLITION`. API-Antworten,
neue Seed-Artefakte, Golden Files und PostgreSQL dürfen `FREE_WILL` nicht neu
erzeugen oder speichern.

## Historische Artefakte

Versionierte Quelldateien werden nicht nachträglich verändert. Sie bleiben als
fachliche Herkunft erhalten und werden beim Lesen in eine kanonische
Laufzeitrepräsentation überführt. Jeder tatsächlich angewendete Alias wird mit
Quellpfad im Compatibility Report aufgeführt; Seed-Importe schreiben dafür ein
Audit Event.

Enthält ein Objekt gleichzeitig `FREE_WILL` und `VOLITION` als Schlüssel,
wird der Import wegen einer Alias-Kollision abgelehnt. Es findet keine stille
Zusammenführung potenziell widersprüchlicher Werte statt.

## Prüfwerkzeug

Die Normalisierung schreibt standardmäßig nur nach stdout und verändert die
historische Eingabedatei nicht:

```bash
go run ./cmd/normalize-dimensions \
  -input data/seed/sprachkompass_coaching-question-pool_v0.1.json \
  > /tmp/question-pool-canonical.json
```

Der Compatibility Report wird getrennt nach stderr geschrieben.
