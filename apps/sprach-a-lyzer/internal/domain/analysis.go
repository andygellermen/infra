package domain

import "github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"

// DimensionID identifies one of the six canonical language dimensions.
type DimensionID = dimension.ID

const (
	DimensionAgency       = dimension.Agency
	DimensionConnection   = dimension.Connection
	DimensionAppreciation = dimension.Appreciation
	DimensionClarity      = dimension.Clarity
	DimensionVolition     = dimension.Volition
	DimensionOpenness     = dimension.Openness
)

func CanonicalDimensions() []DimensionID {
	return dimension.All()
}

// CanonicalDimension maps accepted legacy identifiers to their canonical ID.
func CanonicalDimension(id DimensionID) (DimensionID, bool) {
	canonicalID, err := dimension.Parse(string(id))
	if err != nil {
		return "", false
	}
	return canonicalID, true
}

type AssessabilityState string

const (
	NotAssessable AssessabilityState = "NOT_ASSESSABLE"
	Weak          AssessabilityState = "WEAK"
	Assessable    AssessabilityState = "ASSESSABLE"
	Strong        AssessabilityState = "STRONG"
)
