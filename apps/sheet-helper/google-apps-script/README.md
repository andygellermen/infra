# Google Apps Script Trigger

Diese Vorlage bereitet einen einfachen Sync-Trigger fuer die Sheet-Helper-App vor.

Datei:

- [sync-trigger.js](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/google-apps-script/sync-trigger.js)

## Script Properties

Folgende Script-Properties muessen im Google-Apps-Script-Projekt gesetzt werden:

- `SHEET_HELPER_SYNC_URL`
- `SHEET_HELPER_SYNC_TOKEN`

Beispielwerte:

```text
SHEET_HELPER_SYNC_URL=https://geller.men
SHEET_HELPER_SYNC_TOKEN=s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

## Trigger-Empfehlung

Empfohlener Weg:

1. Script speichern.
2. Im Apps-Script-Editor einmal `installRecommendedTriggers()` ausfuehren.
3. Dadurch wird genau ein installierbarer `On change`-Trigger fuer `sheetHelperOnChange` angelegt.
4. Zusaetzlich gern ein taeglicher Vollsync serverseitig.

Hinweis:

- Der Trigger selbst ist nur ein Signalgeber.
- Die eigentliche Logik fuer Authentifizierung, Sync und Normalisierung lebt in der Go-App.
- Der Sync-Token wird als erster URL-Pfad verwendet, zum Beispiel `https://geller.men/<token>`.
- Die reservierten Funktionsnamen `onEdit` und `onChange` werden absichtlich nicht mehr verwendet.
- Ein reserviertes `onEdit` wuerde sonst zusaetzlich als einfacher Trigger laufen und bei `UrlFetchApp` haeufig Autorisierungsfehler erzeugen.

## Alternative Setups

- `installEditOnlyTrigger()` legt nur einen installierbaren Edit-Trigger fuer `sheetHelperOnEdit` an.
- `installEditAndChangeTriggers()` legt beide installierbaren Trigger an.
- `removeSheetHelperTriggers()` entfernt alte Sheet-Helper-Trigger aus diesem Script-Projekt.

## Migration von alten Triggern

Wenn frueher installierbare Trigger fuer `onEdit` oder `onChange` angelegt wurden:

1. Neue Script-Version einfuegen.
2. `removeSheetHelperTriggers()` einmal ausfuehren.
3. Danach `installRecommendedTriggers()` ausfuehren.

Danach sollten die haeufigen `onEdit`-Fehler aus dem Apps-Script-Log verschwinden.
