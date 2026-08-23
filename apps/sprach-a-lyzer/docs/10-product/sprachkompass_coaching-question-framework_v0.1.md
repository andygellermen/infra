# Sprachkompass – Coaching Question Framework & Pool v0.1

**Umfang:** 100 eigenständig formulierte Coaching-Fragen  
**Zweck:** Kontextgeführte Frage-/Antwort-Analyse für Corporate und Private  
**Status:** Draft zur gemeinsamen fachlichen Kalibrierung

## 1. Zentrale Architekturentscheidung

> **Die Frage bewertet den Nutzer nicht. Sie definiert, welche Konstrukte eine passende Antwort sinnvoll sichtbar machen kann.**

`question_score_bias = 0.0`

Die Antwort wird zunächst auf **Answer Relevance** zur Frage geprüft. Nur wenn sie das Zielkonstrukt tatsächlich adressiert, darf der Question Context Prior die Confidence einer passenden Interpretation erhöhen. Positive oder negative Dimensionsbeiträge entstehen ausschließlich aus der Antwort selbst.

## 2. Phasen

- **P1 ENTRY:** Orientierung, Ziel, erste spontane Sprache.
- **P2 FOLLOWUP 1–10:** Realität, Ressourcen, Ausnahmen, Bedürfnisse, Optionen.
- **P3 ADVANCED 10–20:** Annahmen, Werte, Ambivalenz, systemische Muster.
- **P4 DEEP ANALYTICAL:** identitätsnahe und tiefgreifende Reflexion; explizites Opt-in bei HIGH-Risk-Fragen.

## 3. Schutzregeln

- Keine Diagnose von Personen.
- Keine intime Selbstoffenbarung als Corporate-Pflicht.
- HIGH-Risk-Fragen nur mit Opt-in.
- Spiritual-reflektive Varianten sind eine optionale Präsentations-/Deutungsebene.
- Off-topic/zu kurze Antworten werden nicht künstlich interpretiert.
- Die Frage selbst darf niemals den Score „vorladen“.

## 4. Pool

| ID | Phase | Audience | Kategorie | Frage | Primärkonstrukt |
|---|---|---|---|---|---|
| CQ001 | Einstieg | BOTH | Ziel & Fokus | Was wäre am Ende dieses Gesprächs für dich hilfreich geklärt? | GOAL_ORIENTATION |
| CQ002 | Einstieg | BOTH | Ziel & Fokus | Woran würdest du merken, dass sich dieses Gespräch für dich gelohnt hat? | PREFERRED_FUTURE |
| CQ003 | Einstieg | BOTH | Realität | Was beschäftigt dich daran im Moment am stärksten? | REALITY_OBSERVATION |
| CQ004 | Einstieg | BOTH | Realität | Was ist konkret passiert – möglichst ohne Erklärung, warum es passiert ist? | REALITY_OBSERVATION |
| CQ005 | Einstieg | CORPORATE | System | Welche Rolle hast du in dieser Situation, und welche Erwartungen sind damit verbunden? | SYSTEMIC_CONTEXT |
| CQ006 | Einstieg | PRIVATE | Selbstbezug | Was möchtest du über dich selbst in dieser Situation besser verstehen? | IDENTITY_NARRATIVE |
| CQ007 | Einstieg | BOTH | Einfluss | Was davon liegt heute tatsächlich in deinem Einflussbereich? | LOCUS_OF_CONTROL |
| CQ008 | Einstieg | BOTH | Ressourcen | Was funktioniert trotz der Herausforderung bereits ein wenig? | RESOURCES |
| CQ009 | Einstieg | BOTH | Werte | Was ist dir an dieser Situation besonders wichtig? | VALUES |
| CQ010 | Einstieg | BOTH | Optionen | Welche Möglichkeiten siehst du im Moment – auch unvollständige? | OPTIONS |
| CQ011 | Einstieg | CORPORATE | Ziel & Wirkung | Welche Veränderung würde für dich, dein Team oder eure Zusammenarbeit einen sinnvollen Unterschied machen? | PREFERRED_FUTURE |
| CQ012 | Einstieg | PRIVATE | Bedeutung | Warum ist dieses Thema gerade jetzt für dich bedeutsam? | MEANING |
| CQ013 | Einstieg | BOTH | Sprache | Welchen Satz sagst du dir selbst am häufigsten über diese Situation? | BELIEFS |
| CQ014 | Einstieg | BOTH | Sprache | Wenn du die Situation spontan in einem Satz beschreibst: Wie lautet dieser Satz? | REALITY_OBSERVATION |
| CQ015 | Einstieg | BOTH | Beziehung | Wer oder was ist von deiner Entscheidung noch betroffen? | SYSTEMIC_CONTEXT |
| CQ016 | Einstieg | PRIVATE | Bedürfnisse | Was brauchst du im Moment am meisten, um mit diesem Thema gut weiterzugehen? | NEEDS |
| CQ017 | Einstieg | CORPORATE | Klarheit | Welche Beobachtung würdest du in einem neutralen Protokoll über die Situation festhalten? | REALITY_OBSERVATION |
| CQ018 | Einstieg | BOTH | Skalierung | Auf einer Skala von 0 bis 10: Wo stehst du heute in Bezug auf dein gewünschtes Ergebnis? | GOAL_ORIENTATION |
| CQ019 | Einstieg | BOTH | Skalierung | Was macht deinen heutigen Wert bereits möglich und verhindert, dass er niedriger ist? | RESOURCES |
| CQ020 | Einstieg | BOTH | Ziel & Fokus | Welches Thema sollten wir heute bewusst nicht bearbeiten, damit der Fokus klar bleibt? | BOUNDARIES |
| CQ021 | Folgefrage | BOTH | Vertiefung | Was meinst du genau, wenn du dieses Wort verwendest? | CLARITY |
| CQ022 | Folgefrage | BOTH | Vertiefung | Woran machst du diese Einschätzung konkret fest? | REALITY_OBSERVATION |
| CQ023 | Folgefrage | BOTH | Annahmen | Welche Annahme steckt möglicherweise hinter deiner Schlussfolgerung? | ASSUMPTIONS |
| CQ024 | Folgefrage | BOTH | Perspektive | Wie könnte eine andere beteiligte Person dieselbe Situation beschreiben? | PERSPECTIVE_TAKING |
| CQ025 | Folgefrage | CORPORATE | System | Welche Rahmenbedingung beeinflusst dein Verhalten stärker, als dir lieb ist? | SYSTEMIC_CONTEXT |
| CQ026 | Folgefrage | BOTH | Einfluss | Was könntest du verändern, ohne dass jemand anderes zuerst etwas tun muss? | AGENCY |
| CQ027 | Folgefrage | BOTH | Einfluss | Was versuchst du gerade zu kontrollieren, das nicht vollständig in deiner Hand liegt? | LOCUS_OF_CONTROL |
| CQ028 | Folgefrage | BOTH | Ausnahmen | Wann war die Situation zuletzt ein kleines bisschen leichter? | EXCEPTIONS |
| CQ029 | Folgefrage | BOTH | Ausnahmen | Was war in diesem leichteren Moment anders – bei dir oder im Umfeld? | EXCEPTIONS |
| CQ030 | Folgefrage | BOTH | Ressourcen | Welche Fähigkeit von dir hilft dir hier bereits? | RESOURCES |
| CQ031 | Folgefrage | BOTH | Ressourcen | Wer könnte dich unterstützen, ohne dir die Verantwortung abzunehmen? | RESOURCES |
| CQ032 | Folgefrage | BOTH | Zukunft | Wenn es morgen zehn Prozent besser wäre: Was wäre als Erstes anders? | PREFERRED_FUTURE |
| CQ033 | Folgefrage | BOTH | Zukunft | Wer würde die Veränderung vermutlich zuerst bemerken – und woran? | PREFERRED_FUTURE |
| CQ034 | Folgefrage | BOTH | Optionen | Welche Option hast du bisher ausgeschlossen, ohne sie wirklich zu prüfen? | OPTIONS |
| CQ035 | Folgefrage | BOTH | Optionen | Was würdest du erwägen, wenn du nicht sofort entscheiden müsstest? | OPTIONS |
| CQ036 | Folgefrage | CORPORATE | Entscheidung | Welche Kriterien sollte eine tragfähige Entscheidung für dich erfüllen? | DECISION |
| CQ037 | Folgefrage | PRIVATE | Entscheidung | Welche Entscheidung würde sich für dich zugleich klar und stimmig anfühlen? | DECISION |
| CQ038 | Folgefrage | BOTH | Ambivalenz | Was spricht für Veränderung – und was spricht dafür, alles erst einmal so zu lassen? | AMBIVALENCE |
| CQ039 | Folgefrage | BOTH | Ambivalenz | Welchen Vorteil hat der jetzige Zustand, selbst wenn er dich belastet? | AMBIVALENCE |
| CQ040 | Folgefrage | BOTH | Emotion | Welches Gefühl taucht auf, wenn du an den nächsten Schritt denkst? | EMOTIONAL_AWARENESS |
| CQ041 | Folgefrage | BOTH | Bedürfnisse | Welches Bedürfnis steht hinter deinem Wunsch oder deiner Ablehnung? | NEEDS |
| CQ042 | Folgefrage | BOTH | Grenzen | Wo wäre ein klares Nein hilfreicher als ein halbherziges Ja? | BOUNDARIES |
| CQ043 | Folgefrage | CORPORATE | Beziehung | Welche Formulierung würde dein Anliegen klar ausdrücken, ohne die andere Person abzuwerten? | RELATIONSHIP |
| CQ044 | Folgefrage | BOTH | Sprache | Welche Wörter in deiner bisherigen Beschreibung wirken besonders endgültig – etwa immer, nie, muss oder unmöglich? | BELIEFS |
| CQ045 | Folgefrage | BOTH | Sprache | Wie könntest du denselben Sachverhalt beschreiben, ohne deine Beobachtung zu verändern, aber mit mehr Handlungsspielraum? | AGENCY |
| CQ046 | Folgefrage | BOTH | Selbstbild | Beschreibst du gerade ein Verhalten – oder machst du daraus eine Aussage darüber, wer du bist? | IDENTITY_NARRATIVE |
| CQ047 | Folgefrage | BOTH | Lernen | Was hast du aus einem ähnlichen Moment früher bereits gelernt? | LEARNING |
| CQ048 | Folgefrage | BOTH | Commitment | Was wäre ein kleiner nächster Schritt, den du tatsächlich selbst wählen kannst? | COMMITMENT |
| CQ049 | Folgefrage | BOTH | Commitment | Wann genau könntest du diesen Schritt ausprobieren? | COMMITMENT |
| CQ050 | Folgefrage | BOTH | Skalierung | Was müsste geschehen, damit du auf deiner Skala nur einen halben Punkt weiterkommst? | COMMITMENT |
| CQ051 | Fortgeschritten | BOTH | Überzeugungen | Welche Regel über dich, andere oder die Welt scheint in deiner Antwort mitzuschwingen? | BELIEFS |
| CQ052 | Fortgeschritten | BOTH | Überzeugungen | Was wäre möglich, wenn diese Regel nur eine Hypothese und keine Tatsache wäre? | BELIEFS |
| CQ053 | Fortgeschritten | BOTH | Annahmen | Welche Information könnte deine derzeitige Sichtweise widerlegen? | ASSUMPTIONS |
| CQ054 | Fortgeschritten | CORPORATE | System | Welche unausgesprochene Team- oder Organisationsregel prägt die Situation? | SYSTEMIC_CONTEXT |
| CQ055 | Fortgeschritten | CORPORATE | System | Welche Anreize oder Strukturen fördern genau das Verhalten, das ihr eigentlich verändern möchtet? | SYSTEMIC_CONTEXT |
| CQ056 | Fortgeschritten | BOTH | Perspektive | Welche Sichtweise fällt dir am schwersten ernst zu nehmen – und was könnte daran trotzdem wahr sein? | PERSPECTIVE_TAKING |
| CQ057 | Fortgeschritten | BOTH | Perspektive | Was würdest du einem Menschen raten, den du sehr schätzt, wenn er in deiner Situation wäre? | SELF_APPRECIATION |
| CQ058 | Fortgeschritten | PRIVATE | Selbstwert | Wo sprichst du mit dir härter, als du mit einem anderen Menschen sprechen würdest? | SELF_APPRECIATION |
| CQ059 | Fortgeschritten | BOTH | Einfluss | Welcher Teil deiner Belastung entsteht aus dem Ereignis – und welcher aus deiner Deutung davon? | LOCUS_OF_CONTROL |
| CQ060 | Fortgeschritten | BOTH | Einfluss | Welche Verantwortung gehört tatsächlich zu dir – und welche hast du zusätzlich übernommen? | AGENCY |
| CQ061 | Fortgeschritten | BOTH | Werte | Welcher Wert gerät hier mit einem anderen wichtigen Wert in Konflikt? | VALUES |
| CQ062 | Fortgeschritten | CORPORATE | Werte | Welche Entscheidung wäre mit euren erklärten Werten konsistent – auch wenn sie kurzfristig unbequemer ist? | VALUES |
| CQ063 | Fortgeschritten | PRIVATE | Sinn | Welche Bedeutung möchtest du dieser Erfahrung rückblickend einmal geben können? | MEANING |
| CQ064 | Fortgeschritten | BOTH | Ambivalenz | Welcher Teil von dir möchte vorwärts – und welcher möchte Sicherheit bewahren? | AMBIVALENCE |
| CQ065 | Fortgeschritten | BOTH | Bedürfnisse | Was versuchst du durch dein aktuelles Verhalten zu schützen oder zu erhalten? | NEEDS |
| CQ066 | Fortgeschritten | BOTH | Grenzen | Welche Grenze müsste klarer werden, damit Verbindung wieder leichter möglich ist? | BOUNDARIES |
| CQ067 | Fortgeschritten | CORPORATE | Beziehung | Welche Wirkung hat deine aktuelle Sprache vermutlich auf die Handlungsfähigkeit der anderen? | RELATIONSHIP |
| CQ068 | Fortgeschritten | BOTH | Sprache | Welche deiner Formulierungen beschreibt eine Tatsache – und welche formuliert eine Vorhersage? | CLARITY |
| CQ069 | Fortgeschritten | BOTH | Sprache | Wo benutzt du Notwendigkeitssprache, obwohl eigentlich eine Entscheidung oder Priorität dahintersteht? | FREE_WILL |
| CQ070 | Fortgeschritten | BOTH | Sprache | Welche Formulierung würdest du wählen, wenn du Verantwortung übernehmen möchtest, ohne dir Schuld zuzuschreiben? | AGENCY |
| CQ071 | Fortgeschritten | BOTH | Ausnahmen | Welche Situation beweist, dass deine problematische Generalisierung nicht immer stimmt? | EXCEPTIONS |
| CQ072 | Fortgeschritten | BOTH | Ressourcen | Welche Fähigkeit setzt du bereits selbstverständlich ein und unterschätzt sie deshalb? | RESOURCES |
| CQ073 | Fortgeschritten | BOTH | Optionen | Welche dritte Möglichkeit gibt es zwischen den beiden Polen, zwischen denen du gerade wählst? | OPTIONS |
| CQ074 | Fortgeschritten | BOTH | Optionen | Was wäre ein reversibler Versuch statt einer endgültigen Entscheidung? | OPTIONS |
| CQ075 | Fortgeschritten | CORPORATE | Entscheidung | Welche Entscheidung könntest du treffen, auch wenn noch nicht alle Informationen vorliegen? | DECISION |
| CQ076 | Fortgeschritten | BOTH | Risiko | Was ist der realistischste ungünstige Ausgang – und wie würdest du damit umgehen? | AGENCY |
| CQ077 | Fortgeschritten | BOTH | Commitment | Woran wirst du erkennen, dass dein nächster Schritt wirklich deiner Entscheidung entspricht und nicht nur äußeren Erwartungen? | COMMITMENT |
| CQ078 | Fortgeschritten | BOTH | Lernen | Welche Rückmeldung würdest du nach dem nächsten Versuch brauchen, um sinnvoll nachzusteuern? | LEARNING |
| CQ079 | Fortgeschritten | BOTH | Integration | Welche neue Formulierung möchtest du in den nächsten Tagen bewusst ausprobieren? | INTEGRATION |
| CQ080 | Fortgeschritten | BOTH | Integration | Was könnte dich daran erinnern, in einer angespannten Situation auf diese neue Sprache zurückzugreifen? | INTEGRATION |
| CQ081 | Konkret-tiefgreifend | PRIVATE | Identität | Welche Aussage über dich selbst wiederholst du so oft, dass sie sich wie eine Tatsache anfühlt? | IDENTITY_NARRATIVE |
| CQ082 | Konkret-tiefgreifend | PRIVATE | Identität | Wer wärst du in dieser Situation, wenn du dieses Selbsturteil für einen Moment nicht verwenden würdest? | IDENTITY_NARRATIVE |
| CQ083 | Konkret-tiefgreifend | BOTH | Muster | Welche wiederkehrende sprachliche Form taucht in verschiedenen Situationen immer wieder auf? | BELIEFS |
| CQ084 | Konkret-tiefgreifend | BOTH | Muster | Was löst dieses Sprachmuster typischerweise in deinem Denken oder Handeln aus? | AGENCY |
| CQ085 | Konkret-tiefgreifend | CORPORATE | System | Welche Sprachmuster in eurem Umfeld erzeugen unbeabsichtigt Ohnmacht, Distanz oder Unklarheit? | SYSTEMIC_CONTEXT |
| CQ086 | Konkret-tiefgreifend | CORPORATE | System | Welche Formulierungen in eurem Team erhöhen sichtbar Verantwortung, Wahlmöglichkeiten oder Zusammenarbeit? | SYSTEMIC_CONTEXT |
| CQ087 | Konkret-tiefgreifend | BOTH | Verantwortung | Wo endet Verantwortung und wo beginnt Selbst- oder Fremdbeschuldigung? | AGENCY |
| CQ088 | Konkret-tiefgreifend | PRIVATE | Werte | Welche deiner heutigen Entscheidungen passen noch zu einem früheren Wert, der vielleicht nicht mehr derselbe ist? | VALUES |
| CQ089 | Konkret-tiefgreifend | PRIVATE | Sinn | Welche Erfahrung versuchst du unbedingt zu erklären, statt sie zunächst nur wahrzunehmen? | MEANING |
| CQ090 | Konkret-tiefgreifend | PRIVATE | Bedürfnisse | Welche Bedürfnisse erlaubst du anderen leichter als dir selbst? | NEEDS |
| CQ091 | Konkret-tiefgreifend | BOTH | Beziehung | Welche Zuschreibung an eine andere Person hält eure gegenwärtige Beziehung möglicherweise fest? | RELATIONSHIP |
| CQ092 | Konkret-tiefgreifend | BOTH | Beziehung | Was könntest du klar benennen, ohne über Motive oder Charakter des anderen zu urteilen? | RELATIONSHIP |
| CQ093 | Konkret-tiefgreifend | PRIVATE | Freiheit | Welche Verpflichtung in deinem Leben behandelst du sprachlich als alternativlos, obwohl sie ursprünglich aus einer Wahl entstanden ist? | FREE_WILL |
| CQ094 | Konkret-tiefgreifend | BOTH | Freiheit | Welche Konsequenz wärst du bereit zu tragen, um eine wirklich eigene Entscheidung zu treffen? | FREE_WILL |
| CQ095 | Konkret-tiefgreifend | BOTH | Angst & Handlung | Welche Entscheidung würdest du treffen, wenn Angst mitkommen dürfte, aber nicht entscheiden müsste? | EMOTIONAL_AWARENESS |
| CQ096 | Konkret-tiefgreifend | PRIVATE | Innerer Dialog | Welche Stimme in dir benutzt besonders häufig muss, sollte, immer oder nie – und was möchte sie erreichen? | BELIEFS |
| CQ097 | Konkret-tiefgreifend | BOTH | Integration | Welche Erkenntnis aus diesem Gespräch verändert tatsächlich deinen nächsten Satz oder deine nächste Handlung? | INTEGRATION |
| CQ098 | Konkret-tiefgreifend | BOTH | Integration | Welche alte Formulierung möchtest du künftig schneller bemerken, bevor du ihr automatisch folgst? | INTEGRATION |
| CQ099 | Konkret-tiefgreifend | PRIVATE | Selbstbezug | Was möchtest du dir selbst in dieser Situation nicht länger beweisen müssen? | SELF_APPRECIATION |
| CQ100 | Konkret-tiefgreifend | BOTH | Abschluss | Wenn du deinem heutigen Denken einen neuen, ehrlicheren Satz schenken könntest: Wie würde er lauten? | INTEGRATION |

## 5. Nächster fachlicher Schritt

Für v0.2 sollten wir gemeinsam pro Frage entscheiden:

1. Welche Antwortmuster machen das Primärkonstrukt wirklich assessable?
2. Welche Konstrukte dürfen **nicht** aus dieser Frage abgeleitet werden?
3. Welche Question→Answer-Kompositionen brauchen eigene Golden Cases?
4. Welche Fragen sind zu führend, zu ähnlich oder für Corporate zu intim?
5. Welche 20–30 Fragen bilden ein MVP-Core-Set?
