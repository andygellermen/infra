# Sprachkompass – Question Golden Corpus v0.1

**Umfang:** 54 Q/A-Golden-Cases  
**Zweck:** Testen, dass Fragekontext hilfreiche Interpretation ermöglicht, ohne Score-Bias, Over-Assessment oder unzulässige Kausalitätsbehauptungen zu erzeugen.

## Testdesign

Der Corpus enthält überwiegend **Triplets**:

1. konstruktiv/stützend
2. kontrastierend/einschränkend
3. off-topic Guard Case

Damit prüfen wir ausdrücklich:

> Dieselbe Frage muss gegensätzliche Antworten zulassen.

und:

> Eine Frage darf ein Konstrukt nicht allein dadurch „finden“, dass sie danach gefragt hat.

## Pflichtregeln

- `question_score_bias = 0`
- off-topic → Question Prior aus
- minimale Ja/Nein-Antwort → geringe Independence
- Frage allein → niemals Assessability
- keine Trait-/Diagnosebehauptung
- keine C4-Kausalitätsbehauptung aus normalen Coachingdaten
- spirituelle Variante darf Core Score nicht verändern

## Fälle

| Case | Question | Answer | Relevance | Expected QA Pattern | Direction | Assessability | Max Causal |
|---|---|---|---:|---|---|---|---|
| QG001 | CQ007 | Ich kann die Entscheidung des Kunden nicht beeinflussen, aber ich kann meine Unterlagen vorbereiten und morgen nachfragen. | 0.88 | DIFFERENTIATED_AGENCY | AGENCY:+;FREE_WILL:+ | ASSESSABLE | C1 |
| QG002 | CQ007 | Eigentlich gar nichts. Ich kann nur abwarten. | 0.86 | LOW_PERCEIVED_INFLUENCE | AGENCY:-;FREE_WILL:- | ASSESSABLE | C1 |
| QG003 | CQ007 | Mein Chef ist diese Woche im Urlaub. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG004 | CQ008 | Mir hilft, dass ich ruhig bleibe und früh nach Unterstützung frage. | 0.88 | RESOURCE_RECOGNITION | AGENCY:+;OPENNESS:+ | ASSESSABLE | C1 |
| QG005 | CQ008 | Im Moment fällt mir ehrlich gesagt nichts ein, was mir hilft. | 0.86 | RESOURCE_NOT_YET_ACCESSIBLE |  | ASSESSABLE | C1 |
| QG006 | CQ008 | Wir treffen uns morgen um neun. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG007 | CQ009 | Mir ist Verlässlichkeit wichtig, auch wenn ich dafür eine unbequeme Grenze setzen muss. | 0.88 | VALUE_ARTICULATION | CLARITY:+;FREE_WILL:+ | ASSESSABLE | C1 |
| QG008 | CQ009 | Ich weiß gar nicht, was mir wichtig ist; ich will nur, dass der Druck aufhört. | 0.86 | VALUE_UNCLEAR | CLARITY:- | ASSESSABLE | C1 |
| QG009 | CQ009 | Der Termin wurde verschoben. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG010 | CQ013 | Ich sage mir ständig: Ich muss das allein schaffen. | 0.88 | REPEATED_INTERNAL_RULE | FREE_WILL:-;AGENCY:- | ASSESSABLE | C1 |
| QG011 | CQ013 | Meist sage ich mir: Ich kann Hilfe holen und trotzdem verantwortlich bleiben. | 0.86 | FLEXIBLE_SELF_TALK | AGENCY:+ | ASSESSABLE | C1 |
| QG012 | CQ013 | Ich habe heute Kaffee getrunken. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG013 | CQ021 | Mit 'unmöglich' meine ich: Wir schaffen nicht alle drei Aufgaben bis Freitag mit zwei Leuten. | 0.88 | SPECIFIC_MEANING | CLARITY:+ | ASSESSABLE | C1 |
| QG014 | CQ021 | Unmöglich heißt einfach unmöglich. | 0.86 | UNRESOLVED_ABSOLUTE | CLARITY:- | ASSESSABLE | C1 |
| QG015 | CQ021 | Ich würde lieber später anfangen. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG016 | CQ023 | Ich nehme an, dass sie meine Kritik persönlich nehmen wird, obwohl ich sie noch nicht gefragt habe. | 0.88 | ASSUMPTION_RECOGNITION | CLARITY:+;OPENNESS:+ | ASSESSABLE | C1 |
| QG017 | CQ023 | Da steckt keine Annahme dahinter. Sie wird es definitiv persönlich nehmen. | 0.86 | ASSUMPTION_AS_FACT | OPENNESS:- | ASSESSABLE | C1 |
| QG018 | CQ023 | Ich brauche eine Pause. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG019 | CQ024 | Sie könnte sagen, dass sie unter Zeitdruck stand und meine Rückfrage als zusätzlichen Druck erlebt hat. | 0.88 | PERSPECTIVE_EXPANSION | CONNECTION:+;OPENNESS:+ | ASSESSABLE | C1 |
| QG020 | CQ024 | Es gibt keine andere Sicht. Sie war einfach respektlos. | 0.86 | PERSPECTIVE_CLOSURE | CONNECTION:-;OPENNESS:- | ASSESSABLE | C1 |
| QG021 | CQ024 | Wir haben einen neuen Drucker. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG022 | CQ028 | Ich kann nicht ändern, was er entscheidet. Ich kann aber klar sagen, was ich brauche. | 0.88 | OWN_INFLUENCE_ACTION | AGENCY:+;FREE_WILL:+ | ASSESSABLE | C1 |
| QG023 | CQ028 | Ich muss warten, bis er sich ändert; vorher kann ich nichts tun. | 0.86 | EXTERNAL_DEPENDENCY_NO_CHOICE | AGENCY:-;FREE_WILL:- | ASSESSABLE | C1 |
| QG024 | CQ028 | Die Datei ist 20 MB groß. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG025 | CQ032 | Ich würde zuerst merken, dass ich vor dem Gespräch nicht mehr zehn Szenarien durchspiele, sondern eine konkrete Bitte formuliere. | 0.88 | SMALL_OBSERVABLE_CHANGE | CLARITY:+;AGENCY:+ | ASSESSABLE | C1 |
| QG026 | CQ032 | Keine Ahnung. Besser wäre halt besser. | 0.86 | PREFERRED_FUTURE_VAGUE | CLARITY:- | ASSESSABLE | C1 |
| QG027 | CQ032 | Morgen regnet es wahrscheinlich. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG028 | CQ034 | Ich habe ausgeschlossen, um eine Fristverlängerung zu bitten, weil ich dachte, das wirke schwach. | 0.88 | EXCLUDED_OPTION_WITH_ASSUMPTION | OPENNESS:+;FREE_WILL:+ | ASSESSABLE | C1 |
| QG029 | CQ034 | Es gibt keine andere Option. | 0.86 | NO_OPTION_SPACE | OPENNESS:-;FREE_WILL:- | ASSESSABLE | C1 |
| QG030 | CQ034 | Mein Kalender ist voll. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG031 | CQ038 | Für die Veränderung spricht mehr Ruhe; dagegen spricht, dass ich Sicherheit aufgebe. | 0.88 | ARTICULATED_AMBIVALENCE | OPENNESS:+;CLARITY:+ | ASSESSABLE | C1 |
| QG032 | CQ038 | Es gibt keinen Grund, es so zu lassen. Ich muss das einfach ändern. | 0.86 | ONE_SIDED_CHANGE_PRESSURE | FREE_WILL:- | ASSESSABLE | C1 |
| QG033 | CQ038 | Das Meeting dauert eine Stunde. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG034 | CQ044 | Ich will nicht noch ein Projekt übernehmen. Ein klares Nein wäre ehrlicher als halb zuzusagen. | 0.88 | CLEAR_BOUNDARY | FREE_WILL:+;CLARITY:+ | ASSESSABLE | C1 |
| QG035 | CQ044 | Ich kann nicht Nein sagen; ich muss das eben machen. | 0.86 | BOUNDARY_SUPPRESSION | FREE_WILL:- | ASSESSABLE | C1 |
| QG036 | CQ044 | Das Projekt startet im September. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG037 | CQ046 | Ich sage oft 'immer', 'nie' und 'muss', obwohl es eigentlich nur für diese Woche gilt. | 0.88 | GENERALIZATION_RECOGNITION | OPENNESS:+;CLARITY:+ | ASSESSABLE | C1 |
| QG038 | CQ046 | Diese Wörter passen, weil es wirklich immer so ist. | 0.86 | GENERALIZATION_MAINTAINED | OPENNESS:- | ASSESSABLE | C1 |
| QG039 | CQ046 | Ich spreche normalerweise schnell. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG040 | CQ048 | Ich habe die Präsentation schlecht vorbereitet; das heißt nicht, dass ich unfähig bin. | 0.88 | PERSON_BEHAVIOR_SEPARATION | APPRECIATION:+;AGENCY:+ | ASSESSABLE | C1 |
| QG041 | CQ048 | Ich habe es vermasselt, also bin ich einfach unfähig. | 0.86 | IDENTITY_FUSION_SIGNAL | APPRECIATION:-;AGENCY:- | ASSESSABLE | C1 |
| QG042 | CQ048 | Die Präsentation war gestern. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG043 | CQ049 | Beim letzten Mal habe ich gemerkt, dass ich früher nachfragen muss. Das kann ich diesmal anders machen. | 0.88 | LEARNING_RECOVERY | AGENCY:+;OPENNESS:+ | ASSESSABLE | C1 |
| QG044 | CQ049 | Ich habe gelernt, dass ich solche Dinge einfach nicht kann. | 0.86 | NEGATIVE_IDENTITY_LEARNING | AGENCY:-;APPRECIATION:- | ASSESSABLE | C1 |
| QG045 | CQ049 | Letztes Jahr war das Büro kleiner. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG046 | CQ051 | Ich schreibe heute bis 16 Uhr die drei offenen Punkte auf und bitte dann um Rückmeldung. | 0.88 | OWNED_COMMITMENT | AGENCY:+;CLARITY:+;FREE_WILL:+ | ASSESSABLE | C1 |
| QG047 | CQ051 | Ich sollte irgendwann wirklich etwas tun. | 0.86 | VAGUE_SHOULD_COMMITMENT | CLARITY:-;FREE_WILL:- | ASSESSABLE | C1 |
| QG048 | CQ051 | Mein Laptop braucht ein Update. | 0.18 |  |  | NOT_ASSESSABLE | C0 |
| QG049 | CQ007 | Ja. | 0.52 |  |  | WEAK | C1 |
| QG050 | CQ044 | Nein. | 0.55 |  |  | WEAK | C1 |
| QG051 | CQ013 | Ich muss das allein schaffen. | 0.92 | REPEATED_INTERNAL_RULE | FREE_WILL:- | ASSESSABLE | C1 |
| QG052 | CQ048 | Ich bin heute unkonzentriert, aber ich bin nicht grundsätzlich unfähig. | 0.90 | PERSON_BEHAVIOR_SEPARATION | APPRECIATION:+ | ASSESSABLE | C1 |
| QG053 | CQ038 | Einerseits möchte ich wechseln, andererseits schätze ich die Sicherheit. | 0.95 | ARTICULATED_AMBIVALENCE | OPENNESS:+ | ASSESSABLE | C1 |
| QG054 | CQ051 | Ich werde morgen um zehn anrufen. | 0.93 | OWNED_COMMITMENT | AGENCY:+;CLARITY:+ | ASSESSABLE | C1 |

## CI Acceptance

Für v0.1:

```text
0 Fälle mit question_score_bias != 0
0 off-topic Fälle, die durch die Frage assessable werden
0 C4-Causal Claims
0 Trait-/Diagnoseclaims
100 % Guard Cases müssen bestehen
```

Numerische Score-Corridors werden erst ergänzt, wenn die Q/A-Kompositionsregeln im Referenz-Runner implementiert sind.
