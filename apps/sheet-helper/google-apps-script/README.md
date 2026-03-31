# Google Apps Script Trigger

Diese Vorlage bereitet einen einfachen Sync-Trigger fuer die Sheet-Helper-App vor.

Datei:

- [sync-trigger.js](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/google-apps-script/sync-trigger.js)

## Script Properties

Folgende Script-Properties muessen im Google-Apps-Script-Projekt gesetzt werden:

- `SHEET_HELPER_SYNC_URL`
- `SHEET_HELPER_SYNC_TOKEN`
- `SHEET_HELPER_TENANT`

Beispielwerte:

```text
SHEET_HELPER_SYNC_URL=https://sheet-helper.example.org
SHEET_HELPER_SYNC_TOKEN=replace-me
SHEET_HELPER_TENANT=geller.men
```

## Trigger-Empfehlung

Fuer den Start reicht:

- installierbarer `On edit`-Trigger fuer `onEdit`
- zusaetzlich taeglicher Vollsync serverseitig

Hinweis:

- Der Trigger selbst ist nur ein Signalgeber.
- Die eigentliche Logik fuer Authentifizierung, Sync und Normalisierung lebt in der Go-App.
