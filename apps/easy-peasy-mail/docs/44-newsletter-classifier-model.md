# Newsletter Classifier Model v1

## Newsletter ist nicht Spam

Klassifikation beeinflusst Darstellung, nicht Zustellwahrheit.

## Signale

Gewichtet: - `List-ID` - `List-Unsubscribe` - `List-Unsubscribe-Post` -
`Precedence` - wiederkehrender Sender - Bulk-/Campaign-Header -
Nutzerentscheidung - optional spätere statistische Signale

## Ergebnis

``` text
PERSONAL
TRANSACTIONAL
NEWSLETTER
UNKNOWN
```

## Tabellenidee

``` text
newsletter_sources
- id
- account_id
- source_key
- list_id
- sender_domain
- display_name
- classification
- confidence
- user_override
- updated_at
```

## User Override

Nutzerentscheidung schlägt Automatik.

`Immer normal anzeigen` `Als Newsletter bündeln`

## Conscious View

Viele einzelne Messages bleiben kanonisch vorhanden, aber View
aggregiert:

`Newsletter · 137 neu`

## BurnOut View

Kann sämtliche Einzelmails, Header-Signale, Klassifikationsgrund und
Bundle-Zugehörigkeit zeigen.
