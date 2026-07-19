package tenant

import (
	"encoding/json"
	"sort"
	"strings"
)

const DefaultParticipantCancelDeadlineHours = 24

var defaultEnabledFeatures = []string{
	FeatureCalendar,
	FeatureCertificates,
	FeatureCustomDomains,
	FeatureDonations,
	FeatureParticipantPortal,
	FeaturePayments,
	FeatureSeries,
	FeatureSnippets,
	FeatureWaitlist,
}

func ParticipantCancelDeadlineHoursFromSettingsJSON(raw string) int {
	content := strings.TrimSpace(raw)
	if content == "" {
		return DefaultParticipantCancelDeadlineHours
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return DefaultParticipantCancelDeadlineHours
	}

	value, ok := payload["participant_cancel_deadline_hours"]
	if !ok {
		return DefaultParticipantCancelDeadlineHours
	}

	number, ok := value.(float64)
	if !ok {
		return DefaultParticipantCancelDeadlineHours
	}
	if number < 0 {
		return DefaultParticipantCancelDeadlineHours
	}
	return int(number)
}

func CustomerStatusFromSettingsJSON(raw string) string {
	payload := parseTenantSettingsJSON(raw)
	if payload == nil {
		return CustomerStatusActive
	}
	value, _ := payload["customer_status"].(string)
	return NormalizeCustomerStatus(value)
}

func EnabledFeaturesFromSettingsJSON(raw string) []string {
	payload := parseTenantSettingsJSON(raw)
	if payload == nil {
		return append([]string(nil), defaultEnabledFeatures...)
	}
	features := NormalizeEnabledFeatures(rawStringArray(payload["enabled_features"]))
	if len(features) == 0 {
		return append([]string(nil), defaultEnabledFeatures...)
	}
	return features
}

func FeatureEnabledInSettings(raw, feature string) bool {
	target := strings.TrimSpace(feature)
	if target == "" {
		return false
	}
	for _, item := range EnabledFeaturesFromSettingsJSON(raw) {
		if item == target {
			return true
		}
	}
	return false
}

func NormalizeCustomerStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", CustomerStatusActive:
		return CustomerStatusActive
	case CustomerStatusTrial, CustomerStatusGrace, CustomerStatusPaused, CustomerStatusClosed:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return CustomerStatusActive
	}
}

func NormalizeEnabledFeatures(values []string) []string {
	allowed := map[string]struct{}{
		FeatureCalendar:          {},
		FeatureCertificates:      {},
		FeatureCustomDomains:     {},
		FeatureDonations:         {},
		FeatureParticipantPortal: {},
		FeaturePayments:          {},
		FeatureSeries:            {},
		FeatureSnippets:          {},
		FeatureWaitlist:          {},
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if _, ok := allowed[normalized]; !ok {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func parseTenantSettingsJSON(raw string) map[string]any {
	content := strings.TrimSpace(raw)
	if content == "" {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil
	}
	return payload
}

func rawStringArray(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		result = append(result, strings.TrimSpace(text))
	}
	return result
}
