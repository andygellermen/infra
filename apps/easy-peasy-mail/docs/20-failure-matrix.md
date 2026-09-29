# Failure Matrix

  ------------------------------------------------------------------------------
  Komponente        Ausfall           Nutzerwirkung     Puffer/Recovery
  ----------------- ----------------- ----------------- ------------------------
  Client-Netz       offline           keine neuen Mails lokale DB + Outbox

  Client-Gerät      verloren          Gerät fehlt       revoke + neues Bootstrap

  SES Outbound      gestört           Versand wartet    persistente Outbox

  SES Inbound       gestört           Zustellung        SMTP retries / später
                                      verzögert         zweiter MX

  Event Relay       gestört           Sync verzögert    Core journal + replay

  S3 Temp           gestört           Ingest            Queue/Retry; keine
                                      beeinträchtigt    Bestätigung vor durable
                                                        commit

  EasyPeasy Core    gestört           neue Syncs warten Edge queues + lokale
                                                        Clients

  DB                gestört           Core              restore + journal replay
                                      eingeschränkt     

  Attachment Store  gestört           Remote-Anhänge    lokale Kopien + Backup
                                      fehlen            

  DNS               gestört           neue Zustellung   lange
                                      riskant           TTL-Planung/Monitoring

  Region            gestört           Cloudpfad         spätere
                                      betroffen         Multi-Region-Strategie
  ------------------------------------------------------------------------------

## Goldene Regel

Eine Mail darf aus der temporären Empfangsschicht erst verschwinden,
nachdem der dauerhafte Commit nachweisbar abgeschlossen ist.

## Übungen

Regelmäßig kontrollierte Game-Day-Szenarien durchführen und
Recovery-Zeit sowie Datenverlustfenster dokumentieren.
