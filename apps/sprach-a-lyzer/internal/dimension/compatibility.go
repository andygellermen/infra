package dimension

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var dimensionFields = map[string]bool{
	"dimension":                    true,
	"dimension_id":                 true,
	"dimensions":                   true,
	"canonical_dimension_ids":      true,
	"expected_dimensions":          true,
	"expected_dimension_direction": true,
	"unassessable":                 true,
	"primary_dimension":            true,
	"secondary_dimensions":         true,
	"primary_construct":            true,
	"secondary_constructs":         true,
	"reason_key":                   true,
}

type CompatibilityReport struct {
	Mappings []Mapping `json:"mappings"`
}

func (r CompatibilityReport) LegacyCount() int {
	count := 0
	for _, mapping := range r.Mappings {
		if mapping.Legacy {
			count++
		}
	}
	return count
}

// NormalizeJSON converts legacy dimension identifiers at known semantic
// dimension fields. Other strings are kept byte-semantically unchanged.
func NormalizeJSON(data []byte) ([]byte, CompatibilityReport, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, CompatibilityReport{}, fmt.Errorf("decode dimension-compatible JSON: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, CompatibilityReport{}, err
	}

	report := CompatibilityReport{Mappings: []Mapping{}}
	normalized, err := normalizeValue(value, "$", "", false, &report)
	if err != nil {
		return nil, CompatibilityReport{}, err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return nil, CompatibilityReport{}, fmt.Errorf("encode dimension-compatible JSON: %w", err)
	}
	return encoded, report, nil
}

func NormalizeReader(reader io.Reader) ([]byte, CompatibilityReport, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, CompatibilityReport{}, fmt.Errorf("read dimension-compatible input: %w", err)
	}
	return NormalizeJSON(data)
}

// NormalizeCSV canonicalizes known dimension columns while preserving all
// unrelated cells. It is suitable for the managed-import parsing boundary.
func NormalizeCSV(reader io.Reader) ([]byte, CompatibilityReport, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, CompatibilityReport{}, fmt.Errorf("decode dimension-compatible CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, CompatibilityReport{}, fmt.Errorf("dimension-compatible CSV is empty")
	}

	report := CompatibilityReport{Mappings: []Mapping{}}
	headers := records[0]
	for rowIndex := 1; rowIndex < len(records); rowIndex++ {
		for columnIndex, header := range headers {
			if columnIndex >= len(records[rowIndex]) || !dimensionFields[strings.TrimSpace(header)] {
				continue
			}
			normalized, count := NormalizeExpression(records[rowIndex][columnIndex])
			records[rowIndex][columnIndex] = normalized
			for range count {
				report.Mappings = append(report.Mappings, Mapping{
					Input: LegacyFreeWill, Canonical: Volition, Legacy: true,
					Path: fmt.Sprintf("$[%d].%s", rowIndex, header),
				})
			}
		}
	}

	var output bytes.Buffer
	csvWriter := csv.NewWriter(&output)
	if err := csvWriter.WriteAll(records); err != nil {
		return nil, CompatibilityReport{}, fmt.Errorf("encode dimension-compatible CSV: %w", err)
	}
	return output.Bytes(), report, nil
}

// NormalizeExpression supports scalar, comma-separated and directional legacy
// forms such as FREE_WILL or AGENCY:+;FREE_WILL:-.
func NormalizeExpression(value string) (string, int) {
	var normalized strings.Builder
	count := 0
	cursor := 0
	searchFrom := 0
	for searchFrom < len(value) {
		relative := strings.Index(value[searchFrom:], LegacyFreeWill)
		if relative < 0 {
			break
		}
		start := searchFrom + relative
		end := start + len(LegacyFreeWill)
		if expressionBoundary(value, start-1) && expressionBoundary(value, end) {
			normalized.WriteString(value[cursor:start])
			normalized.WriteString(string(Volition))
			cursor = end
			searchFrom = end
			count++
			continue
		}
		searchFrom = start + 1
	}
	if count == 0 {
		return value, 0
	}
	normalized.WriteString(value[cursor:])
	return normalized.String(), count
}

func expressionBoundary(value string, index int) bool {
	if index < 0 || index >= len(value) {
		return true
	}
	character := value[index]
	return !(character >= 'a' && character <= 'z') &&
		!(character >= 'A' && character <= 'Z') &&
		!(character >= '0' && character <= '9')
}

func normalizeValue(value any, path, parentKey string, dimensionContext bool, report *CompatibilityReport) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		keyContext := parentKey == "dimensions" || parentKey == "constructs" ||
			parentKey == "scores" || parentKey == "entries"
		for key, child := range typed {
			normalizedKey := key
			keyMappings := 0
			if keyContext {
				normalizedKey, keyMappings = NormalizeExpression(key)
			}
			if keyMappings > 0 {
				if _, collision := typed[normalizedKey]; collision {
					return nil, fmt.Errorf("dimension compatibility collision at %s: %s and %s", path, key, normalizedKey)
				}
				for range keyMappings {
					report.Mappings = append(report.Mappings, Mapping{
						Input: LegacyFreeWill, Canonical: Volition, Legacy: true, Path: path + "." + key,
					})
				}
			}
			childPath := path + "." + normalizedKey
			childDimensionContext := dimensionFields[key] || (dimensionContext && key == "id")
			normalizedChild, err := normalizeValue(child, childPath, key, childDimensionContext, report)
			if err != nil {
				return nil, err
			}
			result[normalizedKey] = normalizedChild
		}
		return result, nil

	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			normalizedChild, err := normalizeValue(child, fmt.Sprintf("%s[%d]", path, index), parentKey, dimensionContext, report)
			if err != nil {
				return nil, err
			}
			result[index] = normalizedChild
		}
		return result, nil

	case string:
		if !dimensionContext {
			return typed, nil
		}
		normalized, count := NormalizeExpression(typed)
		for range count {
			report.Mappings = append(report.Mappings, Mapping{
				Input: LegacyFreeWill, Canonical: Volition, Legacy: true, Path: path,
			})
		}
		return normalized, nil

	default:
		return value, nil
	}
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("dimension-compatible input contains more than one JSON value")
		}
		return fmt.Errorf("decode trailing dimension-compatible JSON: %w", err)
	}
	return nil
}

func IsLegacy(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), LegacyFreeWill)
}
