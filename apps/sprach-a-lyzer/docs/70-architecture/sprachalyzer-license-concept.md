# Sprach-A-Lyzer – License Concept v0.1

**Status:** Orientierungs- und Klärungshilfe  
**Stand:** 20. August 2026  
**Zweck:** Strategische Grundlage für Lizenzierung, Governance, Markenführung und Community-Beiträge  
**Hinweis:** Dieses Dokument ist keine Rechtsberatung. Vor finaler Veröffentlichung sollte die gewählte Kombination durch eine auf Open-Source-/IT-Recht spezialisierte juristische Stelle geprüft werden.

---

# 1. Leitidee

Der Sprach-A-Lyzer soll als möglichst frei zugängliche, transparente und gemeinschaftlich weiterentwickelbare Grundlage für sprachliche Reflexion entstehen.

Ziel ist nicht nur „Open Source“ im technischen Sinn, sondern eine Struktur, die verhindert, dass der methodische und technische Kern später exklusiv vereinnahmt oder intransparent abgeschlossen wird.

Gleichzeitig braucht das Projekt klare Herkunft, Verantwortung und Schutz gegen irreführende Nutzung der offiziellen Marke.

Daraus folgt die Leitidee:

> **Der Code soll frei bleiben. Das Wissen soll teilbar bleiben. Die Methodik soll nachvollziehbar bleiben. Die Marke soll Herkunft und Verantwortung schützen.**

---

# 2. Was „frei“ hier bedeuten soll

Der Sprach-A-Lyzer soll grundsätzlich ermöglichen:

- Nutzung
- Untersuchung
- Veränderung
- Weiterentwicklung
- Weitergabe
- kommerzielle Nutzung
- Community-Beiträge
- wissenschaftliche und pädagogische Nutzung
- Forks und alternative Implementierungen

Gleichzeitig soll der freie Kern nach Möglichkeit nicht in proprietären Weiterentwicklungen verschwinden.

Daher ist ein **Copyleft-Modell** gegenüber einer rein permissiven Lizenz strategisch vorzuziehen.

---

# 3. Warum nicht einfach eine einzige Lizenz für alles?

Der Sprach-A-Lyzer besteht aus unterschiedlichen rechtlichen und funktionalen Schichten:

```text
1. Software / Source Code
2. Knowledge Base / Daten
3. Methodik / Dokumentation / Golden Corpus
4. Buchtexte / redaktionelle Werke
5. Marke / Logos / offizielle Domains
6. Community Contributions
```

Diese Bestandteile sollten nicht automatisch unter derselben Lizenz stehen.

---

# 4. Empfohlene Lizenzarchitektur

## 4.1 Software / Source Code

**Zu prüfen:**

```text
AGPL-3.0-or-later
ODER
EUPL-1.2
```

### Option A – AGPL-3.0-or-later

Stärken:

- starke Copyleft-Lizenz
- speziell relevant für serverseitige / webbasierte Software
- modifizierte Versionen, die über ein Netzwerk genutzt werden, müssen Nutzern eine Möglichkeit geben, den entsprechenden Quellcode zu erhalten
- in der Entwicklerwelt sehr bekannt
- klarer Schutz gegen das klassische Schließen einer modifizierten SaaS-Version

Passung für den Sprach-A-Lyzer:

> Sehr hoch, weil der Sprach-A-Lyzer voraussichtlich überwiegend als Webanwendung betrieben wird.

### Option B – EUPL-1.2

Stärken:

- offizielle Lizenz der Europäischen Union
- Copyleft / reziprok
- berücksichtigt Netzwerk-/SaaS-Nutzung
- europäischer Rechtsrahmen
- mehrsprachige offizielle Fassungen
- definierte Kompatibilität mit mehreren anderen Copyleft-Lizenzen

Passung für den Sprach-A-Lyzer:

> Ebenfalls sehr hoch, insbesondere wenn europäische Rechtsnähe und öffentliche/gesellschaftliche Positionierung stärker gewichtet werden.

### Vorläufige Entscheidungshilfe

```text
AGPL-3.0-or-later
→ maximale Bekanntheit im Open-Source-/Developer-Umfeld

EUPL-1.2
→ stärkere europäische Identität und Rechtsnähe
```

**Entscheidung noch offen.**

---

# 5. Knowledge Base / Sprachdaten

Für:

- Lexeme
- Senses
- Phrasen
- Relationen
- Contributions
- Sources
- Golden Test Cases
- Community-Daten

ist **CC BY-SA 4.0** ein starker Kandidat.

Warum?

- freie Weitergabe
- freie Bearbeitung
- kommerzielle Nutzung erlaubt
- Attribution erforderlich
- ShareAlike für Bearbeitungen
- Version 4.0 berücksichtigt auch bestimmte Datenbankrechte

Ziel:

> Verbesserte oder abgeleitete Wissensbestände sollen möglichst wieder offen geteilt werden.

---

# 6. Methodik / Dokumentation / Golden Corpus

Empfehlung:

```text
CC BY-SA 4.0
```

Mögliche Bestandteile:

- Sprachdimensionen
- Relation Taxonomy
- Scoring-Grundlagen
- Assessability-Modell
- Golden Corpus
- Workshopmethodik
- öffentliche Entwicklerdokumentation
- methodische Whitepaper

Gedanke:

> Die Methode darf verwendet, gelehrt, verändert und weiterentwickelt werden – Weiterentwicklungen dieser Materialien bleiben ebenfalls im Commons.

---

# 7. Bücher und redaktionelle Werke

Freie Methodik bedeutet **nicht**, dass jedes Buch vollständig frei lizenziert werden muss.

Mögliche Trennung:

```text
Methodik / Definitionen / offene Datensätze
→ CC BY-SA 4.0

konkreter Buchtext / Illustrationen / Gestaltung
→ klassisches Copyright oder separate Lizenz
```

Das erlaubt:

- offene Methodik
- kommerzielle Publikationen
- redaktionelle Eigenständigkeit
- klare Abgrenzung zwischen Commons und konkretem Werk

---

# 8. Marke und Logos

Empfehlung:

> **Nicht pauschal freigeben.**

Mögliche geschützte Kennzeichen:

- Sprach-A-Lyzer
- Sprachkompass
- Meine Sprache
- Logos
- visuelle Identität
- offizielle Domain-Nutzung

Begründung:

Open Source und Markenschutz sind keine Gegensätze.

Markenschutz dient hier nicht primär zur Abschottung, sondern zur Sicherung von:

- Herkunft
- Authentizität
- Verantwortlichkeit
- Schutz vor irreführenden Forks
- Schutz vor falscher Behauptung einer „offiziellen“ Version

Leitbild:

> **Der Code darf geforkt werden. Die offizielle Herkunft darf nicht gefälscht werden.**

---

# 9. Offizielle Markenpolitik – später ausarbeiten

Eine spätere `TRADEMARK_POLICY.md` sollte klären:

Erlaubt:

- „basiert auf Sprach-A-Lyzer“
- „kompatibel mit Sprach-A-Lyzer“
- „Fork von Sprach-A-Lyzer“

Nicht ohne Freigabe:

- „offizieller Sprach-A-Lyzer“
- Nutzung des offiziellen Logos für abweichende Forks
- irreführende Darstellung als autorisierte Version

---

# 10. Governance – Lizenz allein reicht nicht

Eine freie Lizenz verhindert keine organisatorische Machtkonzentration.

Deshalb sollte langfristig eine Governance-Charta ergänzen:

```text
Sprach-A-Lyzer Commons Principles
```

Mögliche Prinzipien:

1. Der methodische Kern bleibt transparent.
2. Die Engine analysiert Sprache, nicht Personen.
3. Keine offizielle Nutzung für Mitarbeiterranking.
4. Keine psychologische Diagnostik.
5. Fachliche Änderungen bleiben nachvollziehbar.
6. Evidenzklassen und Quellen bleiben sichtbar.
7. Forks bleiben zulässig.
8. Offizielle Releases durchlaufen Review.
9. Community Contributions folgen derselben offenen Lizenzarchitektur.
10. Niemand besitzt die Wahrheit über Sprache.

---

# 11. Wichtige Grenze: Open Source darf Einsatzfelder nicht diskriminieren

Eine OSI-konforme Open-Source-Lizenz darf grundsätzlich nicht bestimmte Personen, Gruppen oder Einsatzfelder ausschließen.

Das bedeutet:

Nicht als Lizenzbedingung formulieren:

```text
„Darf nicht von HR verwendet werden.“
```

oder:

```text
„Darf nicht kommerziell eingesetzt werden.“
```

wenn das Projekt tatsächlich Open Source bleiben soll.

Ethische Grenzen wie:

- kein Employee Ranking
- keine Personendiagnose
- keine Überwachung

sollten daher primär abgesichert werden durch:

- offizielle Produktgestaltung
- Governance
- Markenpolitik
- Hosting-/Nutzungsbedingungen offizieller Instanzen
- Community-Normen
- technische Guardrails in der offiziellen Implementierung

---

# 12. Domain- und Markenarchitektur

Aktuelle Orientierung:

```text
Corporate:
sprachkompass.org

Private:
meinesprache.org

Developer / Support:
sprachalyzer.geller.men
```

`.org` kann dabei kommunikativ für:

- Unabhängigkeit
- Commons-Gedanke
- gesellschaftliche Ausrichtung
- Offenheit

stehen.

Wichtig:

> Die Domain-Endung erzeugt keine rechtliche Freiheit. Die Lizenz- und Governance-Struktur muss diese Bedeutung tatsächlich tragen.

---

# 13. Commons-Seite

Empfehlung für später:

```text
/commons
```

Inhalte:

- Warum ist der Sprach-A-Lyzer offen?
- Welche Teile sind frei?
- Welche Lizenz gilt für welchen Teil?
- Wie kann man beitragen?
- Wie werden Änderungen geprüft?
- Welche Governance gilt?
- Wie werden offizielle Versionen gekennzeichnet?
- Was ist wissenschaftlich belegt, was interpretativ oder spirituell-reflexiv?

---

# 14. Contributor-Modell

Bevorzugtes Prinzip:

> **Beitragen statt Rechte vollständig abtreten.**

Community-Beitragende sollten ihre Urheberrechte grundsätzlich behalten.

Das Projekt erhält die notwendigen Rechte, Beiträge unter der jeweiligen Projektlizenz zu verwenden und weiterzugeben.

Zu prüfen:

```text
DCO (Developer Certificate of Origin)
oder
schlankes Contributor Agreement
```

Ein weitreichendes Copyright Assignment ist nicht automatisch notwendig.

---

# 15. Lizenzzuordnung – vorläufige Matrix

| Bestandteil | Kandidat / Orientierung |
|---|---|
| Engine / Backend / Frontend | AGPL-3.0-or-later **oder** EUPL-1.2 |
| Knowledge Base | CC BY-SA 4.0 |
| Golden Corpus | CC BY-SA 4.0 |
| Methodik-Dokumente | CC BY-SA 4.0 |
| öffentliche technische Doku | CC BY-SA 4.0 oder Softwarelizenz |
| Bücher | separat entscheidbar |
| Illustrationen | separat entscheidbar |
| Logos | geschützt |
| Marken | geschützt |
| offizielle Domains | zentral verwaltet |
| Community Contributions | jeweilige Komponentenlizenz |

---

# 16. Repository-Struktur – Vorschlag

```text
/LICENSE
/LICENSES/
    AGPL-3.0.txt
    EUPL-1.2.txt
    CC-BY-SA-4.0.txt

/NOTICE
/TRADEMARK_POLICY.md
/GOVERNANCE.md
/CONTRIBUTING.md
/COMMONS.md
```

Bei endgültiger Entscheidung sollte nur die tatsächlich relevante Softwarelizenz als primäre Lizenz ausgewiesen werden.

---

# 17. Datei-Header / SPDX

Empfehlung für Code:

```text
SPDX-License-Identifier: AGPL-3.0-or-later
```

oder:

```text
SPDX-License-Identifier: EUPL-1.2
```

Für Daten/Methodik ggf. Metadaten:

```text
license: CC-BY-SA-4.0
```

---

# 18. Lizenz-Metadaten im Importmodell

Für Knowledge-Base-Importe sollte langfristig ein Lizenzfeld vorgesehen werden:

```yaml
source_key
license_key
attribution
source_url
evidence_class
```

Damit können auch externe freie Datenbestände sauber eingebunden werden.

---

# 19. Keine Lizenzvermischung ohne Prüfung

Vor Übernahme externer Daten / Code-Bestandteile immer prüfen:

- Lizenz
- Kompatibilität
- Attribution
- ShareAlike-/Copyleft-Folgen
- Datenbankrechte
- Markenrechte
- mögliche Sonderbedingungen

---

# 20. Entscheidungspunkte vor Release

Vor dem ersten öffentlichen Release klären:

- [ ] AGPL-3.0-or-later oder EUPL-1.2?
- [ ] CC BY-SA 4.0 für Knowledge Base final bestätigen
- [ ] CC BY-SA 4.0 für Golden Corpus final bestätigen
- [ ] Buch-/Content-Lizenz separat festlegen
- [ ] Markenstrategie festlegen
- [ ] Trademark Policy erstellen
- [ ] Governance-Charta erstellen
- [ ] Contributor-Modell festlegen
- [ ] DCO oder Contributor Agreement wählen
- [ ] SPDX-Konvention festlegen
- [ ] Lizenz-Metadaten im Importformat ergänzen
- [ ] juristische Prüfung durchführen

---

# 21. Vorläufige strategische Empfehlung

Bis zur juristischen Prüfung:

```text
Software:
AGPL-3.0-or-later als Favorit
EUPL-1.2 als ernsthafte europäische Alternative

Knowledge Base:
CC BY-SA 4.0

Methodik / Golden Corpus:
CC BY-SA 4.0

Marke:
geschützt

Governance:
offen, transparent, community-orientiert
```

Begründung:

Die Kombination bietet aktuell die beste Balance zwischen:

- Freiheit
- Nachnutzbarkeit
- Copyleft
- SaaS-Schutz
- Offenheit der Wissensbasis
- Schutz der offiziellen Herkunft
- langfristiger Community-Fähigkeit

---

# 22. Philosophischer Leitgedanke

> **Die Marke hat einen Hüter.  
> Die Infrastruktur hat Verantwortliche.  
> Beiträge haben Urheber.  
> Der methodische und technische Commons soll von niemandem wieder eingefangen werden.**

---

# 23. Zweiter Leitgedanke

> **Niemand besitzt Wahrheit über Sprache.**

Der Sprach-A-Lyzer soll Orientierung, Reflexion und Transparenz fördern – nicht selbst zu einer Instanz werden, die festlegt, wie Menschen „richtig“ sprechen müssen.

---

# 24. Dritter Leitgedanke

> **Freiheit ohne Transparenz kann vereinnahmt werden. Schutz ohne Freiheit kann beherrschen. Der Sprach-A-Lyzer soll beides vermeiden.**

---

# 25. Offizielle Referenzen zur finalen Prüfung

- GNU AGPL v3: https://www.gnu.org/licenses/agpl-3.0.html
- GNU License Recommendations: https://www.gnu.org/licenses/license-recommendations.html
- EUPL 1.2: https://interoperable-europe.ec.europa.eu/licence/european-union-public-licence-version-12-eupl
- CC BY-SA 4.0: https://creativecommons.org/licenses/by-sa/4.0/
- CC 4.0 Database Rights: https://wiki.creativecommons.org/wiki/4.0/Sui_generis_database_rights
- Open Source Definition: https://opensource.org/osd

---

# 26. Status

Dieses Dokument ist eine **Orientierungs- und Klärungshilfe**, keine finale Lizenzentscheidung.

Nächste empfohlene Vertiefung:

```text
License Decision Record v0.1
AGPL-3.0-or-later vs. EUPL-1.2
```

mit konkretem Vergleich:

- SaaS-/Netzwerkpflichten
- Kompatibilität
- Contributor-Auswirkungen
- Unternehmensakzeptanz
- europäische Rechtsnähe
- Bibliotheks-/Dependency-Kompatibilität
- mögliche Dual-Licensing-Optionen
