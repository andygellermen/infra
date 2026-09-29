# Offline & Resilience --- Netzstecker-Test

## Ziel

EasyPeasyMail soll bei Teil-Ausfällen sinnvoll weiterarbeiten und
transparent zeigen, welche Funktionen gerade lokal möglich sind.

  -------------------------------------------------------------------------
  Ausfall                             Verhalten
  ----------------------------------- -------------------------------------
  Internet                            Lesen/Suchen/Entwerfen/Organisieren
                                      lokal

  AWS Relay                           direkte Server-Synchronisation bzw.
                                      lokales Weiterarbeiten

  eigener Core                        lokale Nutzung; Outbox wartet

  SES Outbound                        Outbox wartet; kein Doppelversand

  Attachment Store                    vorhandene lokale Anhänge bleiben
                                      verfügbar

  einzelnes Gerät                     andere Replikate unbeeinträchtigt

  lokale DB defekt                    Bootstrap aus Server/Snapshot

  Server-DB defekt                    Restore + Event/Snapshot-Recovery

  komplette Infra offline             Clients bleiben lesbar;
                                      Versand/Empfang wird nach Wiederkehr
                                      aufgeholt
  -------------------------------------------------------------------------

## Härtere Resilienz-Ideen

-   Zwei geografisch getrennte verschlüsselte Backups.
-   Object Storage mit Versionierung für Recovery, aber klarer
    Retention.
-   Append-only Event Journal für kritische Zustandswechsel.
-   Health/lag marker pro Gerät.
-   Outbox mit idempotenten Send-IDs.
-   Inbound Queue vor Parsing: angenommene Mail geht nicht verloren,
    wenn Worker ausfällt.
-   Graceful degradation: Preview statt Body, Body statt Attachment.
-   Exportformat: standardisierte `.eml`/Maildir-Rekonstruktion als
    Exit-Strategie.

## Vermeidungsstrategie

Kein Feature darf unbemerkt eine neue unverzichtbare zentrale
Abhängigkeit erzeugen. Jede Komponente bekommt vor Implementierung eine
dokumentierte Failure Mode + Recovery Route.
