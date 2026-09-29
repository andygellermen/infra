# EasyPeasyMail --- Vision

## Leitidee

EasyPeasyMail ist keine Gmail-Kopie. Es ist eine bewusst reduzierte,
local-first E-Mail-Plattform aus drei zusammengehörigen Bausteinen:

1.  **EasyPeasyMail Client** --- Web/Desktop/Mobile mit hartem Conscious
    Mode.
2.  **EasyPeasyMail Server** --- eigener Empfang, Zustandsverwaltung,
    Synchronisation und cloud-gestützte Verteilung.
3.  **Amazon SES Outbound** --- authentifizierter Versand über
    verifizierte Domains mit DKIM, SPF/Custom MAIL FROM und DMARC.

## Leitprinzipien

-   Local First: Lesen, Suchen, Entwerfen und Organisieren funktionieren
    ohne Netz.
-   Cloud is transient: AWS ist Verteil- und Wiederanlaufebene, nicht
    zwangsläufig das ewige Archiv.
-   Conscious by default: Kommunikation steht vor Oberfläche.
-   Body on demand: Nicht jeder Body und Anhang muss auf jedem Gerät
    liegen.
-   Provider independence: Das System soll nicht von Gmail/Yahoo als
    primärer Mailbox abhängen.
-   Replaceable components: Kein einzelner selbst betriebener Dienst
    darf unnötig zum irreparablen Single Point of Failure werden.
-   Open protocols: SMTP für Transport; IMAP/JMAP als mögliche
    Client-/Migrationsschnittstellen.
-   Encryption and least privilege by design.

## Zielbild

EasyPeasyMail soll eine gesunde Alternative zu großen
Webmail-Oberflächen bieten: ruhig, ausfallsicher, ressourcenschonend und
selbst kontrollierbar.

## Anti-Ziele

-   Kein Funktionswettrüsten mit Gmail/Outlook.
-   Kein dauerhafter Cloud-Zwang für die Bedienung.
-   Keine visuelle Gamification des Postfachs.
-   Kein P2P im MVP nur um P2P zu besitzen.
