# P2P / Adaptive Transport Storming

## Neue Leitidee

Nicht „P2P als Feature", sondern **Adaptive Transport**.

Der Client fragt nicht: \> Benutze ich P2P?

Sondern: \> Was ist der günstigste autorisierte Pfad zu dem Content
Chunk?

## Transport-Kandidaten

1.  lokales Gerät/LAN
2.  direkter QUIC-Pfad
3.  WebRTC DataChannel
4.  EasyPeasy Core
5.  AWS Relay/Object Store

## Discovery

Autorisierte Geräte veröffentlichen nur minimale
Erreichbarkeitsinformationen. Im LAN kann mDNS helfen; über das Internet
vermittelt der Core eine Session, ohne Mailcontent selbst übertragen zu
müssen.

## Security

Jeder Chunk wird über Hash/ID geprüft. Direkter Transport ersetzt keine
Geräteauthentisierung.

## Warum spannend

Ein Smartphone könnte einen 40-MB-Anhang direkt vom eingeschalteten Mac
beziehen, statt ihn erneut aus der Cloud herunterzuladen.

## Warum nicht MVP

NAT, mobile Netze, Browser-Limits und Device Revocation erhöhen
Komplexität. Deshalb bleibt Adaptive Transport eine Optimierung über
einem vollständig funktionierenden Serverpfad.
