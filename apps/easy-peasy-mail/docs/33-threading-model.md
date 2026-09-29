# Threading Model

## Priorität

1.  `In-Reply-To`
2.  `References`
3.  bekannte `Message-ID`
4.  vorsichtige Subject-Heuristik

## Subject Normalization

Nur Fallback: - `Re:`, `Fwd:` und lokalisierte Varianten normalisieren -
Whitespace vereinheitlichen - niemals allein aufgrund gleichen Betreffs
aggressive Threads bilden

## Teilnehmer

Teilnehmer sind unterstützendes Signal, kein harter Schlüssel.

## Split / Merge

Automatisches Merge muss konservativ sein. Falsch getrennte Threads sind
weniger gefährlich als falsch zusammengeführte vertrauliche
Konversationen.

## Newsletter

Newsletter-Bundling ist **kein Threading**. Es ist eine separate
Attention-/Presentation-Schicht.
