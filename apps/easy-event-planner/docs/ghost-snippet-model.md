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
```

## Parameter

`view=cards|list|table|minimal|button`, `limit`, `events=all|upcoming`, `series`, `event`, `target`, `register`, `include_past`, `theme`.

## Gespeicherte Config

```html
<script src="https://events.geller.men/customerxyz/include.js?config=footer-upcoming" defer></script>
```

## Rendering

`include.js` lädt Config, lädt optional CSS, findet/erzeugt Container, ruft Public API auf, rendert HTML und bindet Eventhandler.

## CSS

Präfix `eep-`, optional Shadow DOM, kleine CSS-Datei, keine globalen Theme-Eingriffe.

## MVP-Empfehlung

Snippet zeigt Events und leitet zur autarken Eventseite. Modale Anmeldung kann später ergänzt werden.

## Akzeptanzkriterien

```text
[x] Admin kann Code kopieren
[x] Snippet zeigt kommende Events
[x] list und cards funktionieren
[x] Limit ist parametrisierbar
[x] Zielcontainer funktioniert
[x] keine sensiblen Daten
```
