# Keyboard & Composer

## Vorgeschlagene Shortcuts

  Aktion              Shortcut
  ------------------- ---------------------------
  Senden              Ctrl/Cmd + Enter
  Entwurf speichern   Ctrl/Cmd + S
  Spam                Ctrl/Cmd + Shift + Delete
  Später senden       Ctrl/Cmd + Shift + Enter
  Antworten           R
  Allen antworten     A
  Weiterleiten        F
  Archivieren         E
  Suche               / oder Cmd/Ctrl + K
  Nächste/vorherige   J / K
  Zurück              Esc

Shortcuts müssen konfigurierbar sein und im Browser Konflikte mit
System-/Browserfunktionen berücksichtigen.

## Später senden

Quick Picker: - +15 min - +1 h - heute Abend - morgen früh -
Datum/Uhrzeit in 15-Minuten-Schritten

Server hält den Versandauftrag persistent; ein ausgeschalteter Client
verhindert den Versand nicht.

## Markdown Composer

Neue Mail startet in sauberem Markdown. Mini-HowTo kann dezent beim
ersten Gebrauch erscheinen:

``` text
**fett**   *kursiv*
- Liste
[Link](https://...)
> Zitat
```

Beim Versand: Markdown -\> sanitisiertes HTML + `text/plain`
Alternative.

## Quick React / Inline Comment

Experiment: Textpassage markieren -\> Emoji + Kurzkommentar -\> optional
als zitierte Passage weiterleiten.

Wichtig: Kommentar ist zunächst **lokale Annotation**. Er wird niemals
automatisch Bestandteil der ursprünglichen Mail. Beim Weiterleiten
erfolgt eine explizite Vorschau.
