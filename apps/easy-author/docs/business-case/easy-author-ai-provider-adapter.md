# easy-author AI Provider Adapter

## 1. Ziel

`easy-author` soll verschiedene ReviewProvider anbinden können, ohne sich an einen einzelnen Anbieter zu binden.

## 2. Provider-Typen

```text
mock
openai_compatible
local_model
human_reviewer
custom_webhook
```

## 3. Adapter-Verantwortung

Ein ProviderAdapter übernimmt:

- Umwandlung eines ReviewRequest in provider-spezifisches Format
- Ausführung des Reviews
- Validierung der Antwort
- Umwandlung in ReviewResponse
- Fehlerbehandlung
- Metadatenprotokollierung

## 4. Interface-Skizze Go

```go
type ReviewProvider interface {
    Key() string
    Name() string
    Type() string
    Capabilities() []ReviewCapability
    RunReview(ctx context.Context, req ReviewRequest) (ReviewResponse, error)
}
```

## 5. MockProvider

Der erste Provider sollte ein MockProvider sein.

Zweck:

- UI testen
- ReviewSession testen
- AuthorDecision testen
- keine externen Kosten
- keine Datenschutzrisiken

## 6. OpenAI-kompatibler Provider

Später kann ein OpenAI-kompatibler Provider ergänzt werden.

Dieser sollte strukturierte JSON-Antworten liefern, die gegen ein Schema validiert werden.

## 7. Custom Webhook Provider

Für maximale Offenheit kann ein Custom Webhook Provider angeboten werden.

Ablauf:

1. easy-author sendet ReviewRequest an Webhook.
2. Webhook antwortet mit ReviewResponse.
3. easy-author validiert ReviewItems.
4. Autor entscheidet über die Vorschläge.

## 8. Human Reviewer Provider

Ein menschlicher Reviewer kann dieselbe Logik nutzen.

Beispiel:

- Lektor bekommt Leseraum
- kommentiert Textstellen
- Kommentare werden als ReviewItems gespeichert
- Autor entscheidet

## 9. Fehlerstrategie

Provider dürfen die App nicht blockieren.

Bei Fehlern:

- ReviewSession auf `failed`
- Fehlernachricht speichern
- keine Textänderung durchführen
- Wiederholung erlauben
- Provider wechseln ermöglichen

