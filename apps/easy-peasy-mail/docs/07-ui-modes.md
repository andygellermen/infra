# UI Modes

## 1. Conscious Mode --- Produktkern

Default-Ziel: maximale Ruhe.

Sichtbar: - Absender - Betreff - Zeit - Nachricht - bewusst
zugeschaltete Minimalfunktionen

Optional maximal drei Erweiterungen gleichzeitig: - monochrome Labels -
Attachment Indicator - Thread/Context

Regeln: - keine Avatare - keine Firmenlogos - Newsletter standardmäßig
gebündelt - Werbung/Tracking-nahe Inhalte in Lesedarstellung
unterdrücken - HTML zunächst sanitisiert/reader-like; Originalansicht
bewusst aufrufbar - Keyboard first

## 2. Development Mode

Textzentrierter Arbeitsmodus: - Markdown first - monospace optional -
technische Header aufrufbar - Raw/Source View - Command Palette -
reduzierte IDE-Anmutung

Zweck: nicht „Bewusstsein erzwingen", sondern **bewusste Bedienung
begünstigen**. Das ist die präzisere Produktbehauptung.

## 3. Burnout Mode

Absichtlich funktionsreiche Oberfläche: - Labels - Prioritäten -
zusätzliche Panels - Rich Editor - klassische Mailbox-Werkzeuge

Wichtige Produktentscheidung: Der Name ist intern humorvoll und
prägnant. Für öffentliche UX wäre ein neutralerer Name wie `Full Mode`,
`Classic Mode` oder `Power Mode` wahrscheinlich geeigneter.

## Ablenkungsfreies Schreiben

Beim Schreiben verschwinden Navigation und Nebenflächen ähnlich einem
Focus Editor. Im Mittelpunkt stehen Empfänger, Betreff und Text.

## Status ohne visuelles Rauschen

-   Betreff fett = ungelesen, **nicht** „Body geladen", damit etablierte
    Semantik erhalten bleibt.
-   Body-Cache besser mit sehr kleinem Offline-Punkt/Download-Symbol.
-   Attachment grau = remote; normal = lokal verfügbar.
-   Status muss zusätzlich per Tooltip/ARIA zugänglich sein.
