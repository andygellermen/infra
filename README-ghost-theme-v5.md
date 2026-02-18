# Ghost 5 Theme-Check: `package.json` für ein Caspar/Casper-Derivat

Beim Upgrade auf Ghost 5 schlagen viele Folgefehler auf, wenn die `package.json` **nicht parsebar** ist.
In deinem Beispiel fehlt in `engines` ein Komma nach `"ghost": ">=4.0.0 <6.0.0"`.
Dadurch kann Ghost die Datei nicht lesen und meldet anschließend viele Pflichtfelder als „fehlend".

## Korrigierte `package.json` (Ghost 5 kompatibel)

```json
{
  "name": "casper-extension",
  "description": "Eine Theme-Erweiterung",
  "version": "1.0.1",
  "engines": {
    "ghost": ">=5.0.0 <6.0.0"
  },
  "license": "MIT",
  "author": {
    "name": "Silke Wolter",
    "email": "hello@example.com",
    "url": "https://example.com"
  },
  "config": {
    "posts_per_page": 25,
    "image_sizes": {
      "xxs": { "width": 30 },
      "xs": { "width": 100 },
      "s": { "width": 300 },
      "m": { "width": 600 },
      "l": { "width": 1000 },
      "xl": { "width": 2000 }
    }
  }
}
```

## Hinweise zu deinen Fehlern

- `name`, `version`, `author.email` sind mit obigem JSON vorhanden und gültig.
- `name` ist lowercase + hyphenated (`casper-extension`).
- `version` ist semver (`1.0.1`).
- `author.email` muss eine gültige Mail sein (ersetze `hello@example.com` durch deine echte Adresse).
- `ghost-api` in `engines` ist veraltet und kann entfernt werden.
- Falls du in `config.custom` viele alte Theme-Settings hast:
  - max. 20 Keys,
  - nur snake_case,
  - nur Typen `select`, `boolean`, `color`, `image`, `text`,
  - `select` braucht min. 2 Optionen + gültiges `default`,
  - `boolean` braucht `true/false` als `default`,
  - `color` braucht Hex wie `#15171a`,
  - `image` darf kein echtes Default-Bild setzen.

## `page.hbs`-Warnung (`@page.show_title_and_feature_image`)

Für Ghost 5 solltest du in `page.hbs` den neuen Schalter berücksichtigen.
Beispiel:

```hbs
{{#if @page.show_title_and_feature_image}}
  <header class="page-header">
    {{#if feature_image}}
      <img src="{{img_url feature_image size="l"}}" alt="{{title}}">
    {{/if}}
    <h1>{{title}}</h1>
  </header>
{{/if}}
```

Dadurch verschwindet die Beta-Editor-Warnung zu fehlenden `@page`-Features.
