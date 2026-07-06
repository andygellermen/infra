# Easy-Event-Planner – Ghost/CMS Snippet Model

## Ziel

Veranstalter erzeugen im Admin kopierbare Snippets und fügen diese in Ghost-Content, Ghost Code Injection oder später WordPress/andere CMS ein.

## Grundsatz

Das Snippet ist reine Darstellung. Es darf keine Kapazität final entscheiden, keine sensiblen Daten enthalten und keine Zahlungswahrheit erzeugen.

## Beispiele

```html
<script src="https://events.geller.men/customerxyz/include.js?events=all&view=10&layout=table" defer></script>
<script src="https://events.geller.men/customerxyz/include.js?view=cards&limit=6" defer></script>
<script src="https://events.geller.men/customerxyz/include.js?series=angst-workshop&register=true" defer></script>
<script src="https://events.geller.men/customerxyz/register.js?event=angst-workshop-2026-09-18" defer></script>
```

## Parameter

`view=cards|list|table|minimal|button`, `limit`, `events=all|upcoming`, `series`, `event`, `target`, `register`, `include_past`, `theme`.

## Gespeicherte Config

```html
<script src="https://events.geller.men/customerxyz/include.js?config=footer-upcoming" defer></script>
```

Hinweis (Haertung): In produktiven Setups sollte nur die config-basierte Form verwendet werden. Bei `config=...` werden keine weiteren Query-Parameter akzeptiert.

## Rendering

`include.js` lädt Config, lädt optional CSS, findet/erzeugt Container, ruft Public API auf, rendert HTML und bindet Eventhandler.

## CSS

Präfix `eep-`, optional Shadow DOM, kleine CSS-Datei, keine globalen Theme-Eingriffe.

## Live-Empfehlung

- `include.js` zeigt Event-Listen auf Ghost-/CMS-Seiten und verlinkt auf die gepflegte Detailseite.
- `register.js` bettet das echte EEP-Anmeldeformular direkt auf einer Ghost-/HTML-Seite ein.
- `event_detail_base_url` sollte auf die redaktionelle Detailseite zeigen, z. B. `https://www.geller.men/events`.
- `allowed_embed_origins` muss die einbettende Seite enthalten; fuer bewusst universelle Einbettung ist auch `["*"]` moeglich.

## Akzeptanzkriterien

```text
[x] Admin kann Code kopieren
[x] Snippet zeigt kommende Events
[x] list und cards funktionieren
[x] Limit ist parametrisierbar
[x] Zielcontainer funktioniert
[x] keine sensiblen Daten
```
