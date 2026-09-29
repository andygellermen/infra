# Sync-Protokoll

## Ziel

Geräte synchronisieren überwiegend **Events und Zustände**, nicht immer
vollständige Nachrichten.

## Ablauf

1.  Client meldet `last_event_id`.
2.  Server/AWS Relay liefert Delta.
3.  Client wendet Events idempotent an.
4.  Fehlende Bodies werden nur bei Bedarf angefordert.
5.  Client bestätigt Checkpoint.
6.  Alte Relay-Events dürfen nach Snapshot/TTL verschwinden.

## Konflikte

-   `read/unread`: Last accepted state mit Version.
-   `labels`: mengenbasierte Operationen `add/remove`.
-   `delete`: Tombstone statt sofortiger physischer Löschung.
-   `draft`: explizite Revision; keine stillschweigende Zusammenführung
    zweier parallel bearbeiteter Texte.
-   `send`: idempotency key verhindert Doppelversand.

## Bootstrap eines neuen Geräts

``` text
Auth
 -> encrypted mailbox snapshot
 -> metadata/index
 -> recent HOT bodies
 -> event delta
 -> attachments on demand
```

## P2P Technology Storming

P2P bleibt experimentell, nicht MVP-kritisch.

Kandidaten: - LAN peer discovery + mTLS. - WebRTC DataChannel für
direkte Browser-/Geräteübertragung. - QUIC-basierter Device Sync. -
Content-addressed chunks (Hash) für deduplizierte Bodies/Attachments.

Interessante Hybrid-Idee: AWS/Server vermittelt nur Identität,
Berechtigung und Checkpoint. Wenn zwei autorisierte Geräte erreichbar
sind, können verschlüsselte Chunks direkt übertragen werden. Fällt P2P
aus, übernimmt der Relay-Pfad automatisch.

**Leitregel:** P2P darf Verfügbarkeit erhöhen oder Traffic reduzieren,
niemals die korrekte Funktion voraussetzen.
