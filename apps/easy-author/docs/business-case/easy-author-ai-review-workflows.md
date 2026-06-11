# easy-author AI Review Workflows

## 1. Ziel

Dieses Dokument beschreibt, wie Review-Prozesse in `easy-author` menschenzentriert und selbstbestimmt ablaufen.

KI ist dabei ein möglicher Review-Partner, aber nicht die einzige Quelle. Auch menschliche Reviewer, Lektoren oder Verlagspartner können über dieselbe Review-Logik eingebunden werden.

## 2. Workflow 1: Markierten Abschnitt prüfen

### Ablauf

1. Autor markiert eine Textpassage im Tiptap-Editor.
2. Autor wählt `Review anfragen`.
3. Dialog zeigt:
   - Was wird geprüft?
   - Welcher Review-Typ?
   - Welcher Provider?
   - Welche Daten werden gesendet?
4. Autor bestätigt.
5. ReviewSession wird erstellt.
6. Provider erzeugt ReviewItems.
7. ReviewItems erscheinen in der rechten Sidebar.
8. Autor entscheidet pro ReviewItem.

### Geeignet für

- Formulierungen
- Stilfragen
- kurze Absätze
- sensible Passagen
- Alternativformulierungen

## 3. Workflow 2: Kapitelreview

### Ablauf

1. Autor öffnet ein Kapitel.
2. Autor wählt `Kapitel prüfen`.
3. Review-Dialog bietet Profile:
   - Rechtschreibung & Grammatik
   - Stil & Tonalität
   - Struktur & roter Faden
   - Kapitel-Checkliste
   - Veröffentlichungsreife
4. Autor bestätigt Scope und Provider.
5. ReviewSession läuft.
6. Ergebnis wird nach Kategorien gruppiert.
7. Autor arbeitet ReviewItems ab.

### Ergebnisgruppen

- Sprache
- Stil
- Struktur
- Konsistenz
- Workflow
- Veröffentlichung

## 4. Workflow 3: Roter-Faden-Review

### Ablauf

1. Autor definiert Zielrichtung des Buches oder Projektes.
2. Autor wählt mehrere Kapitel oder das gesamte Buch.
3. Review prüft:
   - Hauptaussage
   - Tonalität
   - Wiederholungen
   - fehlende Übergänge
   - Widersprüche
   - offene Motive
4. ReviewItems werden nicht direkt in Textänderungen übersetzt, sondern als Navigations- und Strukturhinweise gespeichert.

### Besonderheit

Dieser Review soll nicht Satz für Satz korrigieren, sondern dem Autor Orientierung geben.

## 5. Workflow 4: Workflow-Box-Review

### Ablauf

1. Autor öffnet eine WorkflowBox, z. B. `Leitmotive` oder `offene Fragen`.
2. Review prüft, ob diese Workflow-Elemente im Kapitel oder Buch berücksichtigt wurden.
3. ReviewItems werden mit WorkflowBoxen verknüpft.
4. Autor kann daraus Aufgaben, Anker oder Notizen erzeugen.

### Beispiel

WorkflowBox: `Mutmach-Übungen`

Review-Ergebnis:

- Kapitel 3 enthält keine erkennbare Mutmach-Übung.
- Kapitel 5 enthält eine Übung, aber keine klare Anleitung.
- Kapitel 7 enthält eine sehr starke Übung, die als Vorlage dienen könnte.

## 6. Workflow 5: Veröffentlichungsreview

### Ablauf

1. Autor wählt `Veröffentlichung prüfen`.
2. System prüft lokal und optional mit ReviewProvider:
   - offene Kommentare
   - fehlende Alt-Texte
   - fehlende Bildlizenzen
   - leere Kapitel
   - doppelte Kapitelüberschriften
   - unklare Links
   - fehlendes Impressum
   - nicht finaler Kapitelstatus
3. Ergebnis erscheint als Proofing-Checkliste.

### Ziel

Nicht kreative Bewertung, sondern Veröffentlichungssicherheit.

## 7. Workflow 6: Alternative Formulierung ohne Übernahmezwang

### Ablauf

1. Autor markiert Satz oder Absatz.
2. Autor wählt `Alternative vorschlagen`.
3. ReviewProvider liefert eine oder mehrere Alternativen.
4. Alternativen erscheinen als ReviewItems.
5. Autor kann:
   - übernehmen
   - in Clipboard speichern
   - als Notiz speichern
   - ablehnen
   - erneut mit anderem Ton anfragen

## 8. Workflow 7: Stilprofil prüfen

### Ablauf

1. Autor definiert Stilprofil.
2. Review prüft Kapitel gegen Profil.
3. Hinweise beschreiben Abweichungen.
4. KI darf Stilprofil nicht ändern.

### Beispiel-Stilprofil

```text
warm
klar
spirituell
mutmachend
nicht therapeutisch
nicht belehrend
professionell
```

## 9. Author Decision Flow

Jedes ReviewItem folgt diesem Lebenszyklus:

```text
open
→ author decision
→ accepted / rejected / deferred / converted / resolved
```

Mögliche Aktionen:

- Textvorschlag übernehmen
- ablehnen
- später prüfen
- in Notiz umwandeln
- in Anchor umwandeln
- in Workflow-Aufgabe umwandeln
- in Clipboard speichern
- Folgefrage stellen

## 10. UI-Empfehlung

Die rechte Sidebar erhält eine WorkflowBox:

```text
KI-/Review-Hinweise
```

Diese Box sollte umbenennbar sein, z. B.:

- Lektorat
- Stilprüfung
- Roter Faden
- Veröffentlichungscheck
- Kritischer Sparringspartner
- Sanfter Review

## 11. Vermeidungsstrategie

Die App darf den Autor nicht in permanente Optimierung treiben.

Deshalb:

- Reviews müssen aktiv gestartet werden.
- ReviewItems sollen gruppiert und filterbar sein.
- Der Autor kann Review-Typen begrenzen.
- Nicht jeder Hinweis muss erledigt werden.
- Ablehnen ist eine vollwertige Entscheidung.

