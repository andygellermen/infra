# START HERE – Sprachkompass für Cody

Du bekommst hier **kein fertiges ML-Modell**, sondern eine fachlich validierte regelbasierte Referenzarchitektur.

## Bitte zuerst lesen

1. [`DEVELOPER-HANDOFF-v0.1.md`](DEVELOPER-HANDOFF-v0.1.md)
2. danach die [Golden/Reference-Unterlagen](../40-golden/)

## Wichtigste Regeln

1. Kein universeller Wortscore.
2. Sense vor Score.
3. Phrase/Kontext schlagen isoliertes Wort.
4. Fehlende Evidenz ist `null`, nicht 50.
5. Homophonie vererbt keine Semantik.
6. Resonanz ist separate Ebene.
7. Corporate und Private benutzen dieselbe Engine.
8. Corporate bekommt ein eigenes ausgeliefertes Presentation Bundle inkl. Fallbacks.
9. WingScore bewertet Text, nie Mensch.
10. Golden Tests gehören vom ersten Sprint an in CI.

## Der erste technische Vertical Slice

Baue zunächst exakt diese sechs Fälle:

- `Ich muss das heute unbedingt noch schaffen.`
- `Du musst sofort das Gebäude verlassen!`
- `Der Eintritt ist frei.`
- `Er soll sehr erfolgreich sein.`
- `Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.`
- `Hast du Geld?`

Wenn diese sechs Fälle mit Trace reproduzierbar laufen, steht das Fundament.

## Nicht zuerst bauen

- KI-Chat
- Community
- Audio
- Graph-UI
- Browser Extension
- komplexes Admin Panel

Erst Engine + Golden Harness + Explainability.

## Ziel des ersten Demos

Input:

`Ich sollte eigentlich längst weiter sein.`

Output:

- erkannter Sense
- erkannte Patterns
- 6 Dimensionen mit Assessability
- ggf. WingScore
- Contribution Trace
- 1 Reflexionsfrage
- 2 Alternativen

Das ist der erste echte Sprachkompass.
