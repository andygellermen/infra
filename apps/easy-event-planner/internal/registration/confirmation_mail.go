package registration

import (
	"fmt"
	"strings"
	"time"
)

type ConfirmationEmailContentInput struct {
	RecipientName                  string
	EventTitle                     string
	EventStartsAt                  time.Time
	EventTimezone                  string
	EventLocationName              string
	EventOnlineURL                 string
	EventURL                       string
	CalendarURL                    string
	ParticipantCancelDeadlineHours int
}

func BuildConfirmedEmailContent(input ConfirmationEmailContentInput) (subject, bodyText string) {
	subject = "Deine Anmeldung wurde bestaetigt"
	bodyText = buildConfirmationEmailBody(
		input,
		fmt.Sprintf("deine Anmeldung fuer \"%s\" wurde bestaetigt.", strings.TrimSpace(input.EventTitle)),
	)
	return subject, bodyText
}

func BuildWaitlistPromotedEmailContent(input ConfirmationEmailContentInput) (subject, bodyText string) {
	subject = "Nachruecken bestaetigt"
	bodyText = buildConfirmationEmailBody(
		input,
		fmt.Sprintf("dein Wartelistenplatz fuer \"%s\" wurde bestaetigt.", strings.TrimSpace(input.EventTitle)),
	)
	return subject, bodyText
}

func buildConfirmationEmailBody(input ConfirmationEmailContentInput, intro string) string {
	var builder strings.Builder

	name := strings.TrimSpace(input.RecipientName)
	if name == "" {
		name = "Teilnehmer"
	}

	builder.WriteString(fmt.Sprintf("Hallo %s,\n\n%s\n", name, strings.TrimSpace(intro)))

	if scheduleLine := formatConfirmationSchedule(input.EventStartsAt, input.EventTimezone); scheduleLine != "" {
		builder.WriteString(fmt.Sprintf("\nTermin: %s\n", scheduleLine))
	}
	if location := strings.TrimSpace(input.EventLocationName); location != "" {
		builder.WriteString(fmt.Sprintf("Ort: %s\n", location))
	}
	if onlineURL := strings.TrimSpace(input.EventOnlineURL); onlineURL != "" {
		builder.WriteString(fmt.Sprintf("Online-Link: %s\n", onlineURL))
	}

	builder.WriteString("\n")
	builder.WriteString(buildParticipantCancelHint(input.EventStartsAt, input.EventTimezone, input.ParticipantCancelDeadlineHours))

	if calendarURL := strings.TrimSpace(input.CalendarURL); calendarURL != "" {
		builder.WriteString(fmt.Sprintf("\nKalendereintrag direkt uebernehmen: %s\n", calendarURL))
	}
	if eventURL := strings.TrimSpace(input.EventURL); eventURL != "" {
		builder.WriteString(fmt.Sprintf("\nVeranstaltungsdetails: %s\n", eventURL))
	}

	return builder.String()
}

func buildParticipantCancelHint(startsAt time.Time, timezone string, hours int) string {
	deadlineHours := hours
	if deadlineHours < 0 {
		deadlineHours = 0
	}

	deadline := participantCancelDeadlineAt(startsAt, deadlineHours)
	if deadlineHours == 0 {
		return fmt.Sprintf(
			"Falls du versehentlich angemeldet bist, krank wirst oder verhindert bist, melde dich bitte spaetestens bis zum Terminbeginn am %s wieder ab.\n",
			formatConfirmationSchedule(deadline, timezone),
		)
	}

	return fmt.Sprintf(
		"Falls du versehentlich angemeldet bist, krank wirst oder verhindert bist, melde dich bitte bis spaetestens %s wieder ab. Das entspricht %d Stunden vor Terminbeginn.\n",
		formatConfirmationSchedule(deadline, timezone),
		deadlineHours,
	)
}

func participantCancelDeadlineAt(startsAt time.Time, hours int) time.Time {
	if hours <= 0 {
		return startsAt.UTC()
	}
	return startsAt.UTC().Add(-time.Duration(hours) * time.Hour)
}

func ParticipantCancelDeadlineAtForMetadata(startsAt time.Time, hours int) string {
	return participantCancelDeadlineAt(startsAt, hours).UTC().Format(time.RFC3339)
}

func formatConfirmationSchedule(startsAt time.Time, timezone string) string {
	if startsAt.IsZero() {
		return ""
	}

	location := time.UTC
	label := "UTC"
	if value := strings.TrimSpace(timezone); value != "" {
		if loaded, err := time.LoadLocation(value); err == nil {
			location = loaded
			label = value
		}
	}
	return fmt.Sprintf("%s (%s)", startsAt.In(location).Format("02.01.2006 um 15:04"), label)
}
