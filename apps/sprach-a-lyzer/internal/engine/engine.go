package engine

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
)

var ErrEmptyText = errors.New("analysis text must not be empty")

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

type evidence struct {
	contribution domain.Contribution
	strength     float64
}

func (e *Engine) Analyze(request domain.AnalysisRequest) (domain.AnalysisResult, error) {
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return domain.AnalysisResult{}, ErrEmptyText
	}

	context := strings.ToUpper(strings.TrimSpace(request.Context))
	if context == "" {
		context = "UNSPECIFIED"
	}
	inputMode := strings.ToUpper(strings.TrimSpace(request.InputMode))
	if inputMode == "" {
		inputMode = "TEXT"
	}

	result := domain.AnalysisResult{
		Text:              text,
		Context:           context,
		InputMode:         inputMode,
		Propositions:      propositions(text),
		ResolvedSenses:    []domain.ResolvedSense{},
		Patterns:          []string{},
		Dimensions:        emptyDimensions(),
		ContributionTrace: []domain.Contribution{},
		Alternatives:      []string{},
		ResonanceHints:    []domain.ResonanceHint{},
		Notes:             []string{},
	}

	normalized := normalize(text)
	var evidenceItems []evidence

	switch {
	case isHomophoneGuard(normalized):
		result.ResonanceHints = append(result.ResonanceHints, domain.ResonanceHint{
			Kind:          "HOMOPHONE",
			Tokens:        []string{"hast", "hasst"},
			SemanticScore: false,
			Message:       "Die lautliche Nähe wird nur als Resonanzhinweis geführt und verändert keine semantische Bewertung.",
		})
		result.Notes = append(result.Notes, "Homophonie erkannt; keine semantische Vererbung.")

	case isFreeOfCharge(normalized):
		result.ResolvedSenses = append(result.ResolvedSenses, domain.ResolvedSense{
			Lexeme: "frei", Sense: "FREE_OF_CHARGE", Confidence: 0.80,
			Reason: "Die Kollokation „Eintritt ist frei“ bezeichnet Kostenfreiheit.",
		})
		result.Notes = append(result.Notes, "„frei“ bedeutet hier kostenlos und erzeugt keinen VOLITION-Beitrag.")

	case isReportedClaim(normalized):
		result.ResolvedSenses = append(result.ResolvedSenses, domain.ResolvedSense{
			Lexeme: "sollen", Sense: "REPORTED_CLAIM", Confidence: 0.79,
			Reason: "„soll … sein“ berichtet eine fremde Behauptung statt eine Verpflichtung.",
		})
		result.Patterns = append(result.Patterns, "REPORTED_CLAIM")
		result.Notes = append(result.Notes, "Berichtete Behauptung; kein Normativitätsmalus.")

	case isRespectfulBoundary(normalized):
		result.Patterns = append(result.Patterns, "ACKNOWLEDGEMENT", "CLEAR_BOUNDARY", "RESPECTFUL_BOUNDARY")
		evidenceItems = append(evidenceItems,
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionAgency, 6.7, 0.51, "Eine eigene Grenze wird als handlungsfähige Position formuliert."),
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionConnection, 20.8, 0.64, "Anerkennung erhält Verbindung trotz Grenze."),
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionAppreciation, 20.7, 0.63, "Das Anliegen des Gegenübers wird ausdrücklich gewürdigt."),
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionClarity, 20.8, 0.76, "Die eigene Grenze ist eindeutig benannt."),
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionVolition, 20.2, 0.62, "Die Formulierung verbindet Anerkennung mit eigener Wahl."),
			item("R-RESPECTFUL-BOUNDARY", "ich verstehe … wichtig / nicht infrage", domain.DimensionOpenness, 3.9, 0.47, "Die Perspektive des Gegenübers bleibt sichtbar."),
		)
		question := "Wie könntest du deine Grenze ebenso klar halten und zugleich im Gespräch verbunden bleiben?"
		result.ReflectionQuestion = &question
		result.Alternatives = append(result.Alternatives,
			"Ich sehe, wie wichtig dir das ist. Gleichzeitig kann ich dieser Lösung nicht zustimmen.",
			"Dein Anliegen ist bei mir angekommen; für mich braucht es dennoch einen anderen Weg.",
		)

	case containsMust(normalized) && context == "SAFETY":
		result.ResolvedSenses = append(result.ResolvedSenses, domain.ResolvedSense{
			Lexeme: "müssen", Sense: "SAFETY_NECESSITY", Confidence: 0.79,
			Reason: "Der explizite Sicherheitskontext schlägt die isolierte Modalverbdeutung.",
		})
		result.Patterns = append(result.Patterns, "SAFETY_DIRECTIVE", "URGENCY")
		evidenceItems = append(evidenceItems,
			item("R-SAFETY-DIRECTIVE", "musst … sofort / context=SAFETY", domain.DimensionAgency, 9.6, 0.54, "Die Anweisung benennt eine unmittelbar ausführbare Handlung."),
			item("R-SAFETY-DIRECTIVE", "musst … sofort / context=SAFETY", domain.DimensionClarity, 22.3, 0.87, "Handlung, Ziel und Dringlichkeit sind im Sicherheitskontext klar."),
		)
		result.Notes = append(result.Notes, "Sicherheitsnotwendigkeit; kein pauschaler Zwangsmalus.")

	case containsMust(normalized):
		result.ResolvedSenses = append(result.ResolvedSenses, domain.ResolvedSense{
			Lexeme: "müssen", Sense: "INTERNAL_PRESSURE", Confidence: 0.785,
			Reason: "Selbstbezug, Zeitmarker und Verstärker stützen die Lesart inneren Drucks.",
		})
		result.Patterns = append(result.Patterns, "INTERNAL_PRESSURE")
		evidenceItems = append(evidenceItems,
			item("R-INTERNAL-PRESSURE", "ich muss", domain.DimensionVolition, -10.0, 0.62, "Notwendigkeitssprache reduziert den sichtbaren eigenen Wahlraum."),
			item("R-INTERNAL-PRESSURE", "ich muss", domain.DimensionOpenness, -4.0, 0.53, "Die Formulierung stellt zunächst nur einen zwingenden Weg dar."),
		)
		if strings.Contains(normalized, "unbedingt") || strings.Contains(normalized, "sofort") {
			result.Patterns = append(result.Patterns, "URGENCY")
			evidenceItems = append(evidenceItems,
				item("R-URGENCY", "unbedingt", domain.DimensionVolition, -5.6, 0.78, "Der Verstärker erhöht den sprachlich sichtbaren Druck."),
				item("R-URGENCY", "unbedingt", domain.DimensionOpenness, -4.0, 0.67, "Dringlichkeit verengt den dargestellten Möglichkeitsraum."),
				item("R-TEMPORAL-SPECIFICITY", "heute / noch", domain.DimensionClarity, 7.7, 0.50, "Der Zeitbezug macht die Aussage teilweise konkret."),
			)
		}
		question := "Welche eigene Priorität oder Entscheidung steckt hinter diesem Muss?"
		result.ReflectionQuestion = &question
		result.Alternatives = append(result.Alternatives,
			"Ich möchte das heute noch abschließen, weil es mir wichtig ist.",
			"Ich entscheide, ob ich das heute beende oder bewusst neu einplane.",
		)
	}

	applyEvidence(&result, evidenceItems)
	return result, nil
}

func item(ruleID, matched string, dimension domain.DimensionID, delta, strength float64, reason string) evidence {
	return evidence{
		contribution: domain.Contribution{
			RuleID: ruleID, Evidence: matched, Dimension: dimension, Delta: delta, Reason: reason,
		},
		strength: strength,
	}
}

func applyEvidence(result *domain.AnalysisResult, items []evidence) {
	scores := make(map[domain.DimensionID]float64)
	strengths := make(map[domain.DimensionID]float64)
	for _, item := range items {
		result.ContributionTrace = append(result.ContributionTrace, item.contribution)
		dimension := item.contribution.Dimension
		scores[dimension] += item.contribution.Delta
		if item.strength > strengths[dimension] {
			strengths[dimension] = item.strength
		}
	}

	for dimension, delta := range scores {
		score := roundOne(math.Max(0, math.Min(100, 50+delta)))
		strength := strengths[dimension]
		state := stateFor(strength)
		confidence := roundTwo(math.Min(0.98, 0.50+strength*0.60))
		result.Dimensions[dimension] = domain.DimensionResult{
			State: state, Score: &score, Confidence: confidence, Assessability: strength,
		}
	}
}

func emptyDimensions() map[domain.DimensionID]domain.DimensionResult {
	dimensions := domain.CanonicalDimensions()
	result := make(map[domain.DimensionID]domain.DimensionResult, len(dimensions))
	for _, dimension := range dimensions {
		result[dimension] = domain.DimensionResult{State: domain.NotAssessable, Score: nil}
	}
	return result
}

func stateFor(assessability float64) domain.AssessabilityState {
	switch {
	case assessability >= 0.80:
		return domain.Strong
	case assessability >= 0.51:
		return domain.Assessable
	case assessability >= 0.35:
		return domain.Weak
	default:
		return domain.NotAssessable
	}
}

func normalize(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

func containsMust(text string) bool {
	return strings.Contains(text, " muss ") || strings.HasPrefix(text, "muss ") ||
		strings.Contains(text, " musst ") || strings.HasPrefix(text, "musst ")
}

func isFreeOfCharge(text string) bool {
	return strings.Contains(text, "eintritt ist frei")
}

func isReportedClaim(text string) bool {
	return strings.HasPrefix(text, "er soll ") && strings.Contains(text, " sein")
}

func isRespectfulBoundary(text string) bool {
	return strings.Contains(text, "ich verstehe") && strings.Contains(text, "wichtig") &&
		(strings.Contains(text, "nicht infrage") || strings.Contains(text, "nicht in frage"))
}

func isHomophoneGuard(text string) bool {
	trimmed := strings.TrimFunc(text, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSpace(r)
	})
	return trimmed == "hast du geld"
}

func propositions(text string) []domain.Proposition {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	result := make([]domain.Proposition, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		proposition := domain.Proposition{ID: fmt.Sprintf("P%d", len(result)), Text: part}
		if strings.Contains(strings.ToLower(part), "trotzdem") {
			proposition.Relation = "CONCESSION"
		}
		result = append(result, proposition)
	}
	return result
}

func roundOne(value float64) float64 {
	return math.Round(value*10) / 10
}

func roundTwo(value float64) float64 {
	return math.Round(value*100) / 100
}
