# SRE / SLO Model

## SLI-Kandidaten

-   inbound commit latency
-   sync propagation latency
-   outbound queue latency
-   scheduled-send punctuality
-   API availability
-   attachment availability
-   restore success
-   queue age

## Beispiel-SLOs für Pilotphase

Noch keine Marketingversprechen, sondern interne Ziele: - 99.9 %
API-Verfügbarkeit pro Monat - 99 % Inbound-Commits innerhalb 60 s nach
verfügbarer Ingress-Nachricht - 99 % Sync-Events innerhalb 30 s bei
verbundenen Clients - 99 % geplante Sends starten innerhalb 60 s des
Zielzeitpunkts - 0 tolerierter bestätigter Mailverlust

## Error Budget

Featureentwicklung darf Reliability nicht unbegrenzt überstimmen. Wird
das Error Budget überschritten, priorisiert das Team
Stabilität/Recovery.

## Wichtig

SLOs erst nach Messdaten verschärfen. Keine falsche Präzision im MVP.
