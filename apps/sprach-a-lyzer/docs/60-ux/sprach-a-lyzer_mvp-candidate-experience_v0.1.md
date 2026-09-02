# Sprach-A-Lyzer – MVP Candidate Experience v0.1

**Status:** APPROVED
**Version:** 0.1
**Stand:** 3. September 2026
**Owner:** Product & Engineering

## Ziel

Der MVP Candidate macht den bisherigen deterministischen Core als
zusammenhängendes Produkt sichtbar. Private Nutzung heißt **MeineSprache**,
berufliche Nutzung **Sprachkompass**. Beide Profile verwenden dieselbe
Analyse; nur Zugang, Labels und freigegebene Fragen unterscheiden sich.

## Produktfluss

```text
Privat / Berufsleben
→ Text + Kontext + Sprachstufe
→ transiente Core-Analyse ohne KI
→ sechs Dimensionen mit Assessability
→ Textmuster + Contribution-Erklärung
→ Reflexionsimpuls + freigegebene Alternativen
→ freiwillige adaptive Frage
→ progressive C0–C3-Session im Browserzustand
```

## Privacy Defaults

- Analyse- und Antworttexte werden nicht persistiert.
- Antworten verbleiben ausschließlich im Arbeitsspeicher des Browser-Tabs.
- Feedback auf Impulse oder Alternativen bleibt lokal in dieser Sitzung.
- Kein externer Request und keine generative KI sind beteiligt.
- HTTP-Antworten sind `no-store`; CSP, Frame-, Referrer- und Permissions-
  Policies begrenzen den Browserkontext.
- Nicht ausreichende Evidenz bleibt `null` und wird als „offen“ gezeigt.

Der Server darf den Text zur synchronen Verarbeitung empfangen und im
Response an denselben Client zurückgeben. Er schreibt ihn weder in PostgreSQL
noch in Logs oder Analysehistorien.

## Ergebnisgrenzen

Die sichtbare Erklärung bezieht sich immer auf den vorliegenden Text. Sie
diagnostiziert keine Person, erzeugt keine Trait-Aussage und erhebt keinen
Anspruch auf psychologische oder berufliche Eignung. Corporate zeigt
`VOLITION` als „Handlungsspielraum“, Private als „Freier Wille“; der
kanonische Score bleibt identisch.

## Admin-Basis

`/admin` beginnt read-only mit Betriebsstatus, Knowledge-Operations-Hinweisen,
Test-Lab-Einstieg und Audit-/Privacy-Grenzen. Publish und Rollback erscheinen
nicht in dieser Oberfläche. Eine schreibende öffentliche Admin-UI bleibt bis
zur Anbindung einer vertrauenswürdigen Authentisierung fail-closed.

## Acceptance

1. `/` liefert die responsive Produktoberfläche ohne externe Assets.
2. `POST /api/v6/experience/analyze` liefert Core-Ergebnis, Erklärung,
   Alternativen, fünf Fragen und einen maschinenlesbaren Privacy Receipt.
3. Private und Corporate verändern keinen Core-Score.
4. Easy verändert nur freigegebene Fragenrenderings.
5. Der bestehende Session-Vertrag komponiert Antworten ohne Persistenz.
6. `/admin` enthält keine Commit- oder Rollback-Aktion.
7. Die drei Experience-Golden-Fälle sowie alle v0.5-Gates bleiben grün.
