// Package dimension defines the shared, canonical dimension vocabulary.
// It is deliberately small so feature modules can share identifiers without
// depending on one another.
package dimension

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type ID string

const (
	Agency       ID = "AGENCY"
	Connection   ID = "CONNECTION"
	Appreciation ID = "APPRECIATION"
	Clarity      ID = "CLARITY"
	Volition     ID = "VOLITION"
	Openness     ID = "OPENNESS"

	LegacyFreeWill = "FREE_WILL"
)

var canonical = []ID{Agency, Connection, Appreciation, Clarity, Volition, Openness}

type Mapping struct {
	Input     string `json:"input"`
	Canonical ID     `json:"canonical"`
	Legacy    bool   `json:"legacy"`
	Path      string `json:"path,omitempty"`
}

func All() []ID {
	return slices.Clone(canonical)
}

func Parse(value string) (ID, error) {
	canonicalID, _, err := Canonicalize(value)
	return canonicalID, err
}

func Canonicalize(value string) (ID, Mapping, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == LegacyFreeWill {
		return Volition, Mapping{Input: value, Canonical: Volition, Legacy: true}, nil
	}
	for _, id := range canonical {
		if string(id) == normalized {
			return id, Mapping{Input: value, Canonical: id}, nil
		}
	}
	return "", Mapping{}, fmt.Errorf("unknown dimension ID %q", value)
}

func IsCanonical(id ID) bool {
	for _, candidate := range canonical {
		if candidate == id {
			return true
		}
	}
	return false
}

func (id ID) MarshalJSON() ([]byte, error) {
	canonicalID, err := Parse(string(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(canonicalID))
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("dimension ID must be a string: %w", err)
	}
	return id.UnmarshalText([]byte(value))
}

func (id ID) MarshalText() ([]byte, error) {
	canonicalID, err := Parse(string(id))
	if err != nil {
		return nil, err
	}
	return []byte(canonicalID), nil
}

func (id *ID) UnmarshalText(data []byte) error {
	canonicalID, err := Parse(string(data))
	if err != nil {
		return err
	}
	*id = canonicalID
	return nil
}
