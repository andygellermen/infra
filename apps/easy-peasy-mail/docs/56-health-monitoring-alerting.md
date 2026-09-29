# Health, Monitoring & Alerting

## Health Ebenen

### Liveness

Prozess lebt.

### Readiness

Prozess kann echte Arbeit übernehmen.

### Dependency Health

DB, Object Store, SES/Event Edge erreichbar.

## Kernmetriken

-   ingress queue depth/age
-   outbox queue depth/age
-   failed sends
-   parser quarantine rate
-   attachment scan backlog
-   sync event lag
-   DB connection saturation
-   object store errors
-   bounce/complaint trend

## Alarmprinzip

Nicht jede technische Abweichung weckt einen Menschen.

Alert nur, wenn: - Nutzerwirkung wahrscheinlich, - Datenrisiko
besteht, - automatische Recovery nicht greift, - SLO gefährdet ist.

## Dashboard

BurnOut darf alles zeigen. Operations-Dashboard bleibt davon getrennt
und ist für Betreiber.
