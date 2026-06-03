# easy-author – Asset Management

## 1. Ziel

Die Asset-Verwaltung stellt sicher, dass Bilder und andere Medien sauber organisiert, beschrieben, lizenziert und in Exporten korrekt verwendet werden.

## 2. Asset-Typen

- Bilder
- Cover
- Diagramme
- Illustrationen
- Audio optional
- Video optional
- Dokumente optional
- Quellenmaterial optional

## 3. Asset-Metadaten

Jedes Asset sollte folgende Informationen speichern können:

- Titel
- Beschreibung
- Alt-Text
- Dateiname
- Dateityp
- Größe
- Quelle
- Urheber
- Lizenz
- Lizenz-URL
- Lizenzhinweis
- Verwendungszweck
- Ablaufdatum der Lizenz optional
- Freigabe für PDF
- Freigabe für EPUB
- Freigabe für Web

## 4. Asset-Ablage

```text
/assets
  /original
  /optimized
  /export
    /pdf
    /epub
    /web
```

## 5. Importprozess

Beim Import eines Bildes:

1. Datei hochladen
2. Dateiname bereinigen
3. Original speichern
4. optimierte Variante erzeugen
5. Alt-Text abfragen
6. Quelle/Lizenz abfragen
7. Asset in Bibliothek speichern
8. optional direkt in aktuelles Kapitel einfügen

## 6. Verwendung im Markdown

```markdown
![Beschreibung des Bildes](../assets/optimized/bildname.jpg)
```

Optional erweitert:

```markdown
![Beschreibung](../assets/optimized/bildname.jpg){#fig:bildname}
```

## 7. Asset-Verwendungsnachweis

Das System sollte anzeigen:

- Asset verwendet in Kapitel 1
- Asset verwendet im Cover
- Asset nur in Notizen gespeichert
- Asset nicht verwendet
- Asset ohne Lizenzinformation
- Asset ohne Alt-Text

## 8. Export-Prüfung

Vor Exporten prüft das System:

- fehlen Alt-Texte?
- fehlen Lizenzdaten?
- sind Dateien vorhanden?
- sind Bilder zu groß?
- sind Bildformate kompatibel?
- darf das Asset im gewünschten Export verwendet werden?

## 9. Rechte- und Lizenzwarnungen

Ein Asset ohne klare Lizenz sollte nicht unbemerkt in ein öffentliches Buch oder EPUB exportiert werden.

## 10. Vermeidungsstrategie

Bilder nicht nur als dekoratives Material behandeln. Gerade bei späterer Veröffentlichung können fehlende Lizenzangaben hohe Risiken erzeugen. Deshalb sollte die Lizenzpflege von Anfang an Teil des Workflows sein.
