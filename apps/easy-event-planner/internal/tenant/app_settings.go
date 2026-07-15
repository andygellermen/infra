package tenant

import (
	"encoding/json"
	"strings"
)

const DefaultParticipantCancelDeadlineHours = 24

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
