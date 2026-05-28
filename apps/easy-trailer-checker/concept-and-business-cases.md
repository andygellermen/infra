# easy-trailer-checker

## Business-Logik, Randbedingungen und Implementierungsregeln

**Version:** V1
**Ziel:** Führerschein-, Anhänger-, Zuladungs- und 100-km/h-Check für MietHänger und perspektivisch weitere Webseiten

---

## 1. Projektziel

Der **easy-trailer-checker** soll Nutzerinnen und Nutzern helfen, auf Basis ihrer Pkw- und Führerscheindaten geeignete Anhänger aus dem MietHänger-Angebot zu finden.

Der Rechner soll nicht nur prüfen, ob ein Anhänger grundsätzlich gefahren werden darf, sondern auch:

* welche Anhänger zur Führerscheinklasse passen,
* welche Anhänger technisch zum Pkw passen,
* welche maximale Zuladung möglich ist,
* ob 100 km/h voraussichtlich möglich sind,
* welcher Anhänger für den gewählten Einsatzzweck sinnvoll empfohlen werden kann,
* welcher Anhänger direkt zur Anmietung verlinkt werden soll.

Der bestehende alte PHP-Rechner arbeitet bereits mit Custom-Fields aus WordPress und berechnet tabellarisch 80/100/Nein-Ausgaben auf Basis von Pkw-Leergewicht, Pkw-zGM, Anhänger-zGM und Richtgewicht. Diese Logik wird fachlich erhalten, aber sauber erweitert und entkoppelt.

---

## 2. Zielarchitektur

Empfohlene Zielarchitektur:

```text
Website / WordPress / Ghost / Partnerseite
        ↓
statischer Embed-Code
        ↓
leichtes Frontend für Formular und Ergebnisanzeige
        ↓
Go-API / easy-trailer-checker
        ↓
Business-Logik / Regel-Engine
        ↓
JSON-Ergebnis
```

### Grundentscheidung

Die Business-Logik soll **nicht** dauerhaft in WordPress/PHP oder JavaScript liegen.

Empfohlen:

* **Go** für API und Regel-Engine
* **JSON/YAML** oder Datenbank für Anhänger-Stammdaten
* **JavaScript nur für Embed, Formular und Ergebnisdarstellung**
* WordPress/Ghost/Partnerseiten nur als Einbettungsumgebung

---

## 3. Usecases

Es gibt zwei Haupt-Usecases:

## 3.1 Schnellcheck

Ziel:

* sehr niedrige Einstiegshürde
* schnelle Orientierung
* keine vollständige technische Prüfung
* gut für Erstkontakt und Conversion-Einstieg

### Pflichtfelder Schnellcheck

```text
license_class
car_empty_weight_kg
car_gross_weight_kg
```

### Formularfelder

| Feld               | Beschreibung              | Quelle         |
| ------------------ | ------------------------- | -------------- |
| Führerscheinklasse | B, B96, BE, alte Klasse 3 | Führerschein   |
| Pkw-Leergewicht    | Feld G                    | Fahrzeugschein |
| Pkw-zGM            | Feld F.1/F.2              | Fahrzeugschein |

### Ergebnisqualität

Der Schnellcheck darf nur formulieren:

```text
grundsätzlich passend
möglicherweise passend
nicht passend nach Führerscheinklasse
100 km/h voraussichtlich möglich / nicht möglich / bitte Detailcheck nutzen
```

### Einschränkung

Im Schnellcheck werden nicht vollständig geprüft:

* Anhängelast O.1/O.2
* Stützlast
* tatsächliche technische Zugfähigkeit
* tatsächliche maximale Zuladung
* ABS/ABV
* konkrete 100-km/h-Zulässigkeit der Kombination

---

## 3.2 Detailcheck

Ziel:

* möglichst belastbare Anhängerempfehlung
* Prüfung von Führerschein, technischer Anhängelast, Zuladung, Einsatzzweck und 100 km/h
* direkte Empfehlung zur Anmietung

### Pflichtfelder Detailcheck

```text
license_class
car_empty_weight_kg
car_gross_weight_kg
braked_towing_capacity_kg
unbraked_towing_capacity_kg
desired_payload_kg
require_tempo100
```

### Optionale Felder Detailcheck

```text
use_case
min_load_length_cm
min_load_width_cm
min_load_height_cm
car_support_load_kg
has_abs
```

### Formularfelder

| Feld                   | Beschreibung              | Quelle         |
| ---------------------- | ------------------------- | -------------- |
| Führerscheinklasse     | B, B96, BE, alte Klasse 3 | Führerschein   |
| Pkw-Leergewicht        | Feld G                    | Fahrzeugschein |
| Pkw-zGM                | Feld F.1/F.2              | Fahrzeugschein |
| Anhängelast gebremst   | Feld O.1                  | Fahrzeugschein |
| Anhängelast ungebremst | Feld O.2                  | Fahrzeugschein |
| gewünschte Zuladung    | kg                        | Nutzerangabe   |
| 100 km/h gewünscht     | ja/nein/egal              | Nutzerangabe   |
| Einsatzzweck           | Kategorieauswahl          | Nutzerangabe   |
| Mindest-Lademaße       | Länge/Breite/Höhe         | Nutzerangabe   |

---

## 4. Anhänger-Stammdaten

### Aktueller Stammdatensatz

| Anhänger             |      zGM | Leergewicht | Nutzlast | Bremse     | 100 km/h | Lademaße           | URL                                                               |
| -------------------- | -------: | ----------: | -------: | ---------- | -------- | ------------------ | ----------------------------------------------------------------- |
| Planenanhänger XS    |   750 kg |      320 kg |   430 kg | ungebremst | ja       | 256 × 150 × 110 cm | https://miethaenger.com/anhaenger-groesse-xs-750-kg/              |
| Planenanhänger S     | 1.350 kg |      432 kg |   918 kg | gebremst   | ja       | 256 × 150 × 140 cm | https://miethaenger.com/anhaenger-groesse-s-1350-kg/              |
| Planenanhänger M     | 1.500 kg |      429 kg | 1.071 kg | gebremst   | ja       | 311 × 160 × 160 cm | https://miethaenger.com/anhaenger-groesse-m-1500-kg/              |
| Planenanhänger L     | 2.000 kg |      539 kg | 1.461 kg | gebremst   | ja       | 356 × 180 × 180 cm | https://miethaenger.com/anhaenger-groesse-l-2000-kg/              |
| Planenanhänger XL    | 2.500 kg |      548 kg | 1.952 kg | gebremst   | ja       | 406 × 180 × 180 cm | https://miethaenger.com/anhaenger-groesse-xl-2500-kg/             |
| Kofferanhänger K2    | 2.000 kg |      639 kg | 1.361 kg | gebremst   | ja       | 306 × 180 × 154 cm | https://miethaenger.com/anhaenger-kofferanhaenger-2000kg/         |
| Autotransporter AT   | 2.700 kg |      630 kg | 2.070 kg | gebremst   | ja       | 406 × 200 × 63 cm  | https://miethaenger.com/anhaenger-autotransporter-typ-at-2700-kg/ |
| Motorradanhänger MA2 |   750 kg |      280 kg |   470 kg | ungebremst | ja       | 250 × 150 × 63 cm  | https://miethaenger.com/motorradanhaenger-typ-ma2-750-kg/         |
| Motorradanhänger MA3 | 1.500 kg |      400 kg | 1.100 kg | gebremst   | ja       | 250 × 180 × 63 cm  | https://miethaenger.com/motorradanhaenger-typ-ma3-1500-kg/        |
| Kippanhänger         | 2.700 kg |      642 kg | 2.058 kg | gebremst   | ja       | 256 × 150 × 110 cm | https://miethaenger.com/anhaenger-heckkipper-2700kg/              |

---

## 5. Anhänger-Kategorien

```text
planenanhaenger
kofferanhaenger
kipper
autotransporter
motorradanhaenger
```

### Kategorie-Zuordnung

| Kategorie        | Enthaltene Anhänger |
| ---------------- | ------------------- |
| Planenanhänger   | XS, S, M, L, XL     |
| Kofferanhänger   | Kofferanhänger K2   |
| Kippanhänger     | Kippanhänger        |
| Autotransporter  | Autotransporter AT  |
| Motorradanhänger | MA2, MA3            |

---

## 6. Einsatzzwecke

Für V1 werden folgende Einsatzzwecke übernommen:

```text
move_furniture
bulky_items
hardware_store
bulk_material
motorcycle
car_transport
general
max_payload
```

| ID             | Anzeigename                  | Bevorzugte Kategorien                            |
| -------------- | ---------------------------- | ------------------------------------------------ |
| move_furniture | Umzug / Möbeltransport       | Kofferanhänger, Planenanhänger                   |
| bulky_items    | Sperrige Gegenstände         | Planenanhänger, Kofferanhänger                   |
| hardware_store | Baumarkt / Materialtransport | Planenanhänger, Kippanhänger                     |
| bulk_material  | Schüttgut / Garten / Erde    | Kippanhänger                                     |
| motorcycle     | Motorradtransport            | Motorradanhänger                                 |
| car_transport  | Autotransport                | Autotransporter                                  |
| general        | Allgemeiner Transport        | Planenanhänger, Kofferanhänger                   |
| max_payload    | Maximale Zuladung gesucht    | Kippanhänger, Autotransporter, Planenanhänger XL |

### Grundregel

Einsatzzwecke dienen primär zur Sortierung und Empfehlung.

Harte Ausschlüsse nur bei eindeutig unpassenden Kombinationen, z. B.:

```text
Autotransport ohne Autotransporter
Motorradtransport ohne Motorradanhänger
Schüttgut mit Kofferanhänger
```

---

## 7. Ergebnisstatus

Jeder Anhänger erhält einen Status:

```text
recommended
restricted
not_suitable
```

### recommended

Bedeutung:

```text
Der Anhänger passt rechtlich, technisch und vom Einsatzzweck her gut.
```

Voraussetzungen:

```text
Führerschein passt
technische Anhängelast passt
gewünschte Zuladung passt
Lademaße passen
Einsatzzweck passt
100 km/h passt, falls gewünscht
Anhänger ist nicht unnötig groß
```

### restricted

Bedeutung:

```text
Der Anhänger kommt grundsätzlich infrage, aber mit relevanter Einschränkung.
```

Typische Gründe:

```text
nur 80 km/h möglich, obwohl 100 km/h gewünscht
Anhänger darf nicht voll beladen werden
Zuladungsreserve ist knapp
Anhänger ist deutlich größer als nötig
Einsatzzweck passt nur teilweise
Schnellcheck ohne O.1/O.2-Werte
```

### not_suitable

Bedeutung:

```text
Der Anhänger passt nach den eingegebenen Daten nicht.
```

Harte Ausschlussgründe:

```text
Führerscheinklasse reicht nicht
Pkw-Anhängelast reicht nicht einmal für Leergewicht
gewünschte Zuladung überschreitet mögliche Zuladung
Lademaße reichen nicht
Einsatzzweck ist eindeutig unvereinbar
```

---

## 8. Führerscheinlogik

### Berechnungsgrundlage

```text
combination_gross_weight = car_gross_weight_kg + trailer_gross_weight_kg
```

### Klasse B

```text
Erlaubt, wenn:
trailer_gross_weight_kg <= 750

ODER

combination_gross_weight <= 3500
```

### Klasse B96

```text
Erlaubt, wenn:
combination_gross_weight <= 4250
```

### Klasse BE

```text
Erlaubt, wenn:
trailer_gross_weight_kg <= 3500
```

### Alte Klasse 3

```text
Vorläufig bis 7500 kg als erlaubt markieren,
aber mit Hinweis auf Detail- und Altbestandsregelungen.
```

### C/C1/C1E/CE

Für V1 optional, nicht priorisiert.

---

## 9. Technische Anhängelastlogik

### Relevante Anhängelast bestimmen

```text
Wenn trailer.braked = true:
    relevant_towing_capacity = car.braked_towing_capacity_kg

Wenn trailer.braked = false:
    relevant_towing_capacity = car.unbraked_towing_capacity_kg
```

### Maximale technische Anhängermasse

```text
max_technical_trailer_mass = min(
    trailer_gross_weight_kg,
    relevant_towing_capacity
)
```

### Harte Ablehnung

```text
Wenn relevant_towing_capacity < trailer_empty_weight_kg:
    status = not_suitable
    reason = Anhängelast reicht nicht einmal für das Leergewicht des Anhängers.
```

### Eingeschränkte Nutzung

```text
Wenn relevant_towing_capacity < trailer_gross_weight_kg:
    Anhänger ist nur mit reduzierter Zuladung nutzbar.
```

---

## 10. Zuladungslogik

### Berechnung

```text
max_payload_by_trailer = trailer_payload_kg

max_payload_by_towing_capacity =
    relevant_towing_capacity - trailer_empty_weight_kg

effective_max_payload =
    min(max_payload_by_trailer, max_payload_by_towing_capacity)
```

Im Schnellcheck:

```text
effective_max_payload = trailer_payload_kg
Hinweis: technische Anhängelast des Pkw wurde noch nicht geprüft.
```

### Bewertung

```text
Wenn desired_payload_kg > effective_max_payload:
    status = not_suitable

Wenn desired_payload_kg <= effective_max_payload
und Reserve < 10 %:
    status = restricted

Wenn Reserve 10–25 %:
    status = recommended

Wenn Reserve sehr groß:
    prüfen, ob kleinerer Anhänger sinnvoller ist.
```

---

## 11. 100-km/h-Logik

### Grundidee

Die 100-km/h-Prüfung wird separat ausgewiesen und darf nicht mit der allgemeinen Fahrbarkeit vermischt werden.

### Schnellcheck

Im Schnellcheck nur:

```text
tempo100_plausible
```

auf Basis von:

```text
trailer_tempo100_approved
car_empty_weight_kg
tempo100_min_car_empty_weight_kg
```

### Detailcheck

Im Detailcheck:

```text
tempo100_likely_allowed =
    trailer_tempo100_approved == true
    AND car_empty_weight_kg >= trailer_tempo100_min_car_empty_weight_kg
    AND relevant_towing_capacity >= trailer_gross_weight_kg
    AND car_has_abs != false
```

### Bewertung

| Fall                                                   | Ergebnis                        |
| ------------------------------------------------------ | ------------------------------- |
| 100 km/h nicht gewünscht                               | keine Abwertung                 |
| 100 km/h gewünscht und möglich                         | recommended                     |
| 100 km/h gewünscht, aber nicht möglich                 | restricted                      |
| 100 km/h ausdrücklich erforderlich, aber nicht möglich | not_suitable möglich            |
| Daten unvollständig                                    | Hinweis / Detailcheck empfohlen |

---

## 12. Lademaßprüfung

Nur wenn Nutzer Mindestmaße eingibt.

```text
min_length_cm <= trailer.load_length_cm
min_width_cm <= trailer.load_width_cm
min_height_cm <= trailer.load_height_cm
```

### Bewertung

```text
Wenn Mindestmaß nicht erfüllt:
    status = not_suitable

Wenn Maßreserve sehr knapp:
    status = restricted

Wenn Maße passen:
    status bleibt empfohlen oder eingeschränkt je nach anderer Logik.
```

---

## 13. Scoring und Sortierung

Zusätzlich zum Status wird ein Score berechnet.

### Beispiel-Scoring

```text
+50 Führerschein passt
+50 technische Anhängelast passt
+30 gewünschte Zuladung passt
+20 Einsatzzweck passt
+20 100 km/h möglich, falls gewünscht
+10 Lademaße passen

-30 nur 80 km/h, obwohl 100 gewünscht
-20 Anhänger deutlich größer als nötig
-30 Zuladung sehr knapp
-50 Einsatzzweck unpassend
```

### Sortierreihenfolge

```text
1. recommended
2. restricted
3. not_suitable
4. innerhalb des Status: höchster Score
5. kleinster ausreichend passender Anhänger bevorzugen
6. 100-km/h-Eignung bevorzugen, wenn gewünscht
7. direkter Mietlink anzeigen
```

---

## 14. API-Request V1

```json
{
  "mode": "detail",
  "license_class": "B96",
  "car": {
    "empty_weight_kg": 1650,
    "gross_weight_kg": 2150,
    "braked_towing_capacity_kg": 1800,
    "unbraked_towing_capacity_kg": 750,
    "support_load_kg": null,
    "has_abs": true
  },
  "requirements": {
    "desired_payload_kg": 1000,
    "require_tempo100": true,
    "use_case": "move_furniture",
    "min_load_length_cm": null,
    "min_load_width_cm": null,
    "min_load_height_cm": null
  }
}
```

---

## 15. API-Response V1

```json
{
  "mode": "detail",
  "summary": {
    "recommended_count": 2,
    "restricted_count": 3,
    "not_suitable_count": 5
  },
  "results": [
    {
      "trailer_id": "planen-m-1500",
      "name": "Planenanhänger M",
      "status": "recommended",
      "score": 165,
      "effective_max_payload_kg": 1071,
      "max_speed_kmh": 100,
      "messages": [
        "Die Kombination passt zu Ihrer Führerscheinklasse.",
        "Die eingegebene Anhängelast reicht aus.",
        "Die gewünschte Zuladung ist möglich.",
        "100 km/h sind nach den eingegebenen Daten voraussichtlich möglich.",
        "Der Anhänger passt gut zu Ihrem Einsatzzweck."
      ],
      "url": "https://miethaenger.com/anhaenger-groesse-m-1500-kg/"
    }
  ],
  "legal_notice": "Alle Angaben dienen nur als Orientierungshilfe. Bitte prüfen Sie vor der Anmietung die Angaben in Ihrem Fahrzeugschein."
}
```

---

## 16. Kundenausgabe

### Empfohlen

```text
✅ Empfohlen

Dieser Anhänger passt sehr gut zu Ihren Angaben. Die gewünschte Zuladung ist möglich, die Lademaße sind ausreichend und die Kombination ist nach den eingegebenen Daten passend.
```

Button:

```text
Anhänger ansehen / direkt anmieten
```

### Eingeschränkt passend

```text
⚠️ Eingeschränkt passend

Dieser Anhänger kommt grundsätzlich infrage, aber es gibt eine Einschränkung: Die Nutzung mit 100 km/h ist nach Ihren Angaben voraussichtlich nicht möglich.
```

Button:

```text
Trotzdem ansehen
```

### Nicht passend

```text
❌ Nicht passend

Dieser Anhänger passt nach Ihren Angaben nicht. Grund: Die zulässige Gesamtmasse der Kombination überschreitet die Grenze Ihrer Führerscheinklasse.
```

Optional:

```text
Passende Alternativen anzeigen
```

---

## 17. Gewährleistungs- und Hinweistext

Empfohlene HTML-Version:

```html
<h3>Bitte beachten!</h3>
<p>
  Der Rechner dient nur als Orientierungshilfe. Bitte prüfen Sie vor der Anmietung immer
  die Angaben in Ihrem Fahrzeugschein, insbesondere die zulässige Gesamtmasse, das Leergewicht,
  die zulässige Anhängelast und die Stützlast Ihres Fahrzeugs.
</p>

<p>
  Eine 100-km/h-Zulassung des Anhängers bedeutet nicht automatisch, dass jede Kombination
  aus Pkw und Anhänger mit 100 km/h gefahren werden darf. Entscheidend ist, ob Ihr konkretes
  Zugfahrzeug zusammen mit dem gewählten Anhänger alle Voraussetzungen erfüllt.
</p>

<p>
  Alle Angaben erfolgen ohne Gewähr.
</p>
<hr>
```

---

## 18. Testfälle V1

### Pkw-Testprofile

| Profil                   | Leergewicht |      zGM | gebremst | ungebremst |
| ------------------------ | ----------: | -------: | -------: | ---------: |
| Kleinwagen               |    1.150 kg | 1.650 kg | 1.000 kg |     600 kg |
| Kompaktklasse            |    1.450 kg | 1.950 kg | 1.500 kg |     680 kg |
| Mittelklasse-Kombi       |    1.650 kg | 2.150 kg | 1.800 kg |     750 kg |
| SUV/leichter Transporter |    2.050 kg | 2.600 kg | 2.500 kg |     750 kg |

### Wichtige Testfälle

```text
01 Kleinwagen + Klasse B + Planenanhänger XS + 300 kg Zuladung
Erwartung: nicht passend für 300 kg, eingeschränkt passend bis ca. 280 kg

02 Kleinwagen + Klasse B + Planenanhänger S + 500 kg Zuladung
Erwartung: eingeschränkt passend, Anhänger nicht voll beladbar

03 Kompaktklasse + Klasse B + Planenanhänger M + 900 kg Zuladung
Erwartung: empfohlen

04 Kompaktklasse + Klasse B + Planenanhänger L
Erwartung: nicht passend wegen Führerschein

05 Kompaktklasse + B96 + Planenanhänger L + 850 kg Zuladung
Erwartung: eingeschränkt passend

06 Mittelklasse-Kombi + Klasse B + Planenanhänger M
Erwartung: nicht passend wegen Führerschein

07 Mittelklasse-Kombi + B96 + Planenanhänger M
Erwartung: empfohlen

08 SUV + BE + Planenanhänger XL + 100 km/h gewünscht
Erwartung: eingeschränkt passend, 100 km/h voraussichtlich nicht möglich

09 SUV + BE + Kippanhänger + Schüttgut
Erwartung: eingeschränkt passend, sehr guter Einsatzzweck, aber nicht voll beladbar

10 SUV + BE + Autotransporter
Erwartung: eingeschränkt passend, Autotransport möglich, aber nicht voll beladbar

11 Kompaktklasse + Klasse B + MA2 + Motorradtransport
Erwartung: eingeschränkt passend, Zuladung passt, 100 km/h voraussichtlich nicht möglich

12 Mittelklasse-Kombi + B96 + MA3 + Motorradtransport
Erwartung: empfohlen
```

### Grenzfälle

```text
A: Klasse B, Kombination exakt 3.500 kg
Erwartung: erlaubt

B: Klasse B, Kombination 3.501 kg
Erwartung: nicht passend

C: B96, Kombination exakt 4.250 kg
Erwartung: erlaubt

D: B96, Kombination 4.251 kg
Erwartung: nicht passend

E: Anhängelast reicht nur knapp über Leergewicht
Erwartung: eingeschränkt / praktisch nicht empfohlen
```

---

## 19. Go-Datenmodell V1

### Trailer

```go
type Trailer struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Category                 string `json:"category"`
	GrossWeightKg            int    `json:"gross_weight_kg"`
	EmptyWeightKg            int    `json:"empty_weight_kg"`
	PayloadKg                int    `json:"payload_kg"`
	Braked                   bool   `json:"braked"`
	Tempo100Approved         bool   `json:"tempo100_approved"`
	Tempo100MinCarWeightKg   *int   `json:"tempo100_min_car_weight_kg,omitempty"`
	LoadLengthCm             int    `json:"load_length_cm"`
	LoadWidthCm              int    `json:"load_width_cm"`
	LoadHeightCm             int    `json:"load_height_cm"`
	URL                      string `json:"url"`
}
```

### VehicleInput

```go
type VehicleInput struct {
	LicenseClass             string `json:"license_class"`
	CarEmptyWeightKg         int    `json:"car_empty_weight_kg"`
	CarGrossWeightKg         int    `json:"car_gross_weight_kg"`
	BrakedTowingCapacityKg   int    `json:"braked_towing_capacity_kg"`
	UnbrakedTowingCapacityKg int    `json:"unbraked_towing_capacity_kg"`
	DesiredPayloadKg         int    `json:"desired_payload_kg"`
	RequireTempo100          bool   `json:"require_tempo100"`
	UseCase                  string `json:"use_case"`
	MinLoadLengthCm          *int   `json:"min_load_length_cm,omitempty"`
	MinLoadWidthCm           *int   `json:"min_load_width_cm,omitempty"`
	MinLoadHeightCm          *int   `json:"min_load_height_cm,omitempty"`
	CarSupportLoadKg         *int   `json:"car_support_load_kg,omitempty"`
	HasABS                   *bool  `json:"has_abs,omitempty"`
}
```

### TrailerResult

```go
type TrailerResult struct {
	TrailerID             string   `json:"trailer_id"`
	Name                  string   `json:"name"`
	Status                string   `json:"status"`
	Score                 int      `json:"score"`
	EffectiveMaxPayloadKg int      `json:"effective_max_payload_kg"`
	MaxSpeedKmh           int      `json:"max_speed_kmh"`
	Messages              []string `json:"messages"`
	URL                   string   `json:"url"`
}
```

---

## 20. Validierung

### Eingabegrenzen

```text
car_empty_weight_kg:
Minimum 500
Maximum 3500

car_gross_weight_kg:
Minimum 800
Maximum 3500

braked_towing_capacity_kg:
Minimum 0
Maximum 3500

unbraked_towing_capacity_kg:
Minimum 0
Maximum 750

desired_payload_kg:
Minimum 0
Maximum 2100
```

### Plausibilitätsregeln

```text
car_empty_weight_kg darf nicht größer als car_gross_weight_kg sein

unbraked_towing_capacity_kg sollte maximal 750 kg betragen

Wenn Detailcheck gewählt:
O.1/O.2 müssen vorhanden sein

Wenn Schnellcheck gewählt:
fehlende O.1/O.2 müssen sichtbar als Einschränkung ausgegeben werden
```

---

## 21. Vermeidungsstrategien

Um falsche Empfehlungen zu vermeiden:

```text
Schnellcheck immer als Orientierung kennzeichnen
Detailcheck klar von Schnellcheck trennen
Tempo 100 separat ausweisen
Jede Ablehnung begründen
Bei fehlenden Daten keine harte Freigabe formulieren
Rechtliche/technische Hinweise verständlich ausgeben
Fahrzeugschein-Felder direkt am Formular erklären
Anhänger nicht nur nach Größe, sondern nach Zweck und Nutzlast empfehlen
```

---

## 22. Implementierungsreihenfolge

### Phase 1: Fachliche Grundlage

```text
Anhänger-Stammdaten als JSON erfassen
Führerscheinlogik implementieren
technische Anhängelastlogik implementieren
Zuladungslogik implementieren
100-km/h-Logik implementieren
Bewertungsstatus und Scoring implementieren
```

### Phase 2: API

```text
POST /api/check
Request validieren
Anhänger iterieren
Prüfergebnisse berechnen
Status und Score setzen
Ergebnis sortieren
JSON zurückgeben
```

### Phase 3: Frontend / Embed

```text
Schnellcheck-Formular
Detailcheck-Formular
Ergebnisgruppen:
- Empfohlen
- Eingeschränkt passend
- Nicht passend

CTA:
- Anhänger ansehen
- Direkt anmieten
- Detailcheck starten
```

### Phase 4: Tests

```text
Unit-Tests für Führerscheinlogik
Unit-Tests für Anhängelast
Unit-Tests für Zuladung
Unit-Tests für 100 km/h
Unit-Tests für Scoring
Integrationstests mit Testfällen V1
```

---

## 23. Offene spätere Erweiterungen

```text
Admin-Oberfläche für Anhänger-Stammdaten
Mandantenfähigkeit für Partnerseiten
White-Label-Embeds
Tracking von Conversion und Abbruchstellen
Mehrsprachigkeit
präzisere Stützlastprüfung
Reifenalter-/100-km/h-Prüfung
Verfügbarkeitsprüfung je Standort
direkter Buchungsprozess
```

---

## 24. Finale V1-Entscheidungen

```text
Schnellcheck:
- Führerscheinklasse
- Pkw-Leergewicht
- Pkw-zGM

Detailcheck:
- Führerscheinklasse
- Pkw-Leergewicht
- Pkw-zGM
- Anhängelast gebremst
- Anhängelast ungebremst
- gewünschte Zuladung
- 100 km/h gewünscht
- optional Einsatzzweck und Mindestmaße

Status:
- empfohlen
- eingeschränkt passend
- nicht passend

Technik:
- Go-Regelservice
- statischer Embed-Code
- JavaScript nur für UI
- WordPress/Ghost/Partnerseiten als Einbettungsziele
```
