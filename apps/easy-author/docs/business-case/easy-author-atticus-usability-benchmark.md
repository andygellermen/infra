# easy-author – Atticus-inspirierter Usability-Benchmark

## Zweck

Dieses Dokument übersetzt beobachtete Usability-Muster aus Atticus in ein eigenständiges easy-author-Anforderungskonzept. Ziel ist nicht, Atticus zu kopieren, sondern die hilfreichen Einstieg-, Schreib-, Strukturierungs-, Formatierungs- und Exportmuster für easy-author nutzbar zu machen.

easy-author bleibt dabei eigenständig:

- Markdown-nah und exportfreundlich
- Tiptap-basiert
- self-hosted- und infra-tauglich
- mit Workflow-Boxen, Ankern, Clipboard-Slots und Autoren-Wissensbank
- später erweiterbar um Leseraum, Review, Ghost-Publishing und Monetarisierung

## Leitidee

Atticus führt Autoren stark entlang eines klaren Buchprozesses. easy-author sollte diesen Gedanken aufnehmen und in eine eigene Phasenlogik übersetzen:

1. Projekt anlegen
2. Manuskript importieren oder neu beginnen
3. Werkstruktur aufbauen
4. Schreiben und überarbeiten
5. Autorenwissen, Assets und Quellen verwalten
6. Layout-/Theme-Entscheidungen treffen
7. Proofing durchführen
8. Export oder Veröffentlichung vorbereiten

## Bewertungsprinzip

Jede Funktion wird für easy-author nicht als Feature-Wunsch, sondern als Workflow-Nutzen bewertet:

- Hilft sie dem Autor beim Schreiben?
- Vermeidet sie spätere Fehler?
- Unterstützt sie Export, Veröffentlichung oder Review?
- Passt sie zum Markdown-/Tiptap-Konzept?
- Kann sie im MVP klein begonnen und später ausgebaut werden?

## Zentrale UX-Muster

### 1. Einstieg über klare Startoptionen

easy-author sollte im Dashboard nicht nur „Datei öffnen“ anbieten, sondern konkrete Startwege:

- Neues Buchprojekt
- Manuskript importieren
- Markdown-Ordner importieren
- Buchreihe/Sammlung vorbereiten
- Leseprobe/Freebie erstellen
- Ghost-Serie vorbereiten

### 2. Linke Werkstruktur als dauerhafte Orientierung

Autoren benötigen dauerhaft Orientierung über:

- Front Matter
- Kapitel
- Szenen
- Back Matter
- Fragmente
- Notizen
- Export-Sets

Die linke Sidebar wird damit zur Werk-Navigation.

### 3. Rechte Kontextseite als Autorenintelligenz

Die rechte Sidebar bleibt ein easy-author-Unterscheidungsmerkmal. Sie zeigt nicht nur Formatoptionen, sondern Kontext:

- Workflow-Boxen
- Anker
- Clipboard-Slots
- Personen
- Orte
- Ereignisse
- Assets
- Kommentare
- offene Entscheidungen

### 4. Export nicht als letzter Klick, sondern als begleitender Qualitätsprozess

Schon während des Schreibens sollten Hinweise entstehen:

- fehlende Bildlizenzen
- fehlende Alt-Texte
- doppelte Kapiteltitel
- sehr lange Kapitel
- offene TODOs
- Kommentare ohne Freigabe
- fehlendes Impressum
- fehlendes Cover

### 5. Wiederverwendbare Buchseiten

Autoren brauchen Seiten, die projektübergreifend wiederverwendet werden:

- Über den Autor
- Impressum
- Copyright
- Weitere Bücher
- Spendenhinweis
- Newsletter-Hinweis
- Rezensionsbitte
- Verlagskontakt

Änderungen sollten nicht unbemerkt global synchronisiert werden, sondern bewusst bestätigt werden.

## Abgrenzung zu Atticus

Atticus ist stark bei Buchformatierung und Export. easy-author soll stärker beim Denken, Verknüpfen und Workflow werden.

| Bereich | Atticus-Stärke | easy-author-Ziel |
|---|---|---|
| Einstieg | schnelle Buchanlage | Werktypen, Import-Assistent, Projektlogik |
| Schreiben | fokussierter Editor | Tiptap, Markdown-Nähe, Workflow-Anker |
| Struktur | Kapitelverwaltung | Front/Body/Back Matter, Szenen, Fragmente, Export-Sets |
| Formatierung | Themes | Theme Studio mit PDF/EPUB/DOCX/Ghost-Ausrichtung |
| Export | PDF/EPUB | PDF, EPUB, DOCX, Markdown, HTML, Ghost, Verlags-Paket |
| Kontext | eher buchorientiert | Autoren-Wissensbank, Anker, Clipboard, Assets, Quellen |
| Review | Proofing | Leseraum, Gruppen, Kommentare, Rezensionen |

## Umsetzungsempfehlung

Für den nächsten Entwicklungsschritt sollte easy-author nicht vollständig umgebaut werden. Sinnvoll ist eine kontrollierte UX-Erweiterung:

1. Dashboard erweitern
2. Werktypen und Projektstatus ergänzen
3. Front/Body/Back-Matter in der Kapitelstruktur sichtbar machen
4. Kapitelstatus einführen
5. erste Export-/Proofing-Checkliste ergänzen
6. wiederverwendbare Buchseiten als Datenmodell vorbereiten
7. Theme-Grundmodell vorbereiten

## Nicht-Ziele für diesen Schritt

Noch nicht bauen:

- vollständiger DOCX-Import
- echtes PDF/EPUB-Rendering
- Font-Management
- Zahlungsfunktionen
- Ghost-Publishing
- vollständige Leseräume
- komplexe Kollaboration

## Ergebnisbild

Nach Umsetzung dieses UX-Schrittes soll easy-author für Autoren nicht mehr wie ein Editor-Spike wirken, sondern wie der Anfang eines echten Authoring Studios:

- Startklar über Dashboard
- klare Werkstruktur
- sichtbarer Schreibfortschritt
- erste Buchlogik
- vorbereitete Exportqualität
- harmonische Verbindung mit Workflow-Boxen und Tiptap-Editor
