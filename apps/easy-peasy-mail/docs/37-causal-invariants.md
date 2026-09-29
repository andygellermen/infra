# Causal Invariants & Folge-Haken

## I1 --- Accepted means recoverable

Eine eingehende Mail darf erst als übernommen gelten, wenn ihre
dauerhafte Recovery-Kette steht.

Folge: - Ingress bleibt bis Commit erhalten. - Worker verwendet Lease. -
Poison Messages gehen in Quarantäne.

## I2 --- Sent means exactly one logical send

UI darf `SENT` erst zeigen, wenn Send Ledger/Provider-Ergebnis dies
rechtfertigt.

Folge: - Idempotency Key. - keine blinden Retries nach unklarem
Provider-Timeout.

## I3 --- Delete must converge

Ein lange offline befindliches Gerät darf gelöschte Daten nicht
reanimieren.

Folge: - Tombstones. - Snapshot kennt Purge-Cursor. - alte Clients
benötigen ggf. Full Rebootstrap.

## I4 --- Attachment availability must be explicit

Mailbody darf lesbar bleiben, wenn Attachment Store ausfällt.

Folge: - Attachment State separat. - UI zeigt
remote/local/unavailable/quarantined.

## I5 --- Newsletter is presentation, not truth

Newsletter-Klassifizierung darf eine Mail nicht aus dem eigentlichen
Mailbestand entfernen.

Folge: - Bundles sind Views. - Nutzer kann Klassifizierung rückgängig
machen.

## I6 --- Spam uncertainty must not destroy mail

Unsichere Klassifizierung führt zu Spam/Quarantine, nicht sofortigem
Purge.

## I7 --- Device cache is disposable

Verlust eines Clients darf keine kanonischen Daten verlieren.

## I8 --- Event relay is disposable

AWS/Relay-Events dürfen verfallen, weil Core Journal/Snapshot Recovery
trägt.

## I9 --- Original interoperability survives optimization

Attachment-Auslagerung, Reader View und Bundling dürfen
`.eml`/Maildir-Export nicht verhindern.

## I10 --- Modes do not mutate truth

Conscious, Development und BurnOut sind unterschiedliche
Darstellungen/Interaktionen derselben kanonischen Daten.
