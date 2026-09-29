# Delivery Observability

## Ziel

Auf die Frage „Wo ist meine Mail?" muss das System technisch antworten
können.

## Inbound Timeline

``` text
EDGE_ACCEPTED
INGRESS_DURABLE
PARSED
AUTH_CHECKED
CLASSIFIED
COMMITTED
SYNC_EVENT_CREATED
DEVICE_SYNCED
```

## Outbound Timeline

``` text
OUTBOX_CREATED
SCHEDULED?
SEND_STARTED
PROVIDER_ACCEPTED
PROVIDER_MESSAGE_ID
DELIVERY_EVENT?
BOUNCE?
COMPLAINT?
```

## Correlation

Jede Pipeline nutzt: - request_id - message_id - ingress_id oder
outbox_id - account_id - provider_message_id falls vorhanden

Keine Mailinhalte in normalen Logs.

## User-facing

Conscious: `Gesendet` / `Wartet` / `Zustellung fehlgeschlagen`

Development: technische Timeline auf Wunsch.

BurnOut: vollständige Delivery-/Sync-/Provider-Sicht.
