package httpapp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/notification"
)

type mailerMagicLinkSender struct {
	mailer    notification.Mailer
	provider  string
	env       string
	fromEmail string
	fromName  string
}

func buildMagicLinkSender(cfg config.Config) auth.Sender {
	mailer, err := notification.NewMailer(notification.MailerConfig{
		Provider:  cfg.MailProvider,
		FromEmail: cfg.MailFromEmail,
		FromName:  cfg.MailFromName,
		SMTPHost:  cfg.SESHost,
		SMTPPort:  cfg.SESPort,
		SMTPUser:  cfg.SESUser,
		SMTPPass:  cfg.SESPass,
	})
	if err != nil {
		log.Printf("magic-link sender fallback to log sender: %v", err)
		return &auth.LogSender{}
	}

	return &mailerMagicLinkSender{
		mailer:    mailer,
		provider:  strings.ToLower(strings.TrimSpace(cfg.MailProvider)),
		env:       strings.TrimSpace(cfg.Env),
		fromEmail: strings.TrimSpace(cfg.MailFromEmail),
		fromName:  strings.TrimSpace(cfg.MailFromName),
	}
}

func (s *mailerMagicLinkSender) SendMagicLink(ctx context.Context, message auth.MagicLinkMessage) error {
	subject, bodyText, bodyHTML := buildMagicLinkMailContent(message)
	metadataJSON := buildMagicLinkMetadataJSON(message)

	if !strings.EqualFold(s.env, "production") {
		log.Printf(
			"magic-link debug tenant=%s email=%s purpose=%s verify_url=%s expires_at=%s",
			strings.TrimSpace(message.TenantSlug),
			strings.TrimSpace(message.ToEmail),
			strings.TrimSpace(message.Purpose),
			strings.TrimSpace(message.VerifyURL),
			message.ExpiresAt.UTC().Format(time.RFC3339),
		)
	}

	return s.mailer.Send(ctx, notification.Message{
		TemplateKey:  "auth_" + strings.TrimSpace(message.Purpose),
		Recipient:    strings.TrimSpace(message.ToEmail),
		Subject:      subject,
		BodyText:     bodyText,
		BodyHTML:     bodyHTML,
		MetadataJSON: metadataJSON,
		FromEmail:    s.fromEmail,
		FromName:     s.fromName,
	})
}

func buildMagicLinkMailContent(message auth.MagicLinkMessage) (subject, bodyText, bodyHTML string) {
	purpose := strings.ToLower(strings.TrimSpace(message.Purpose))
	subject = "Dein Magic Link fuer Easy Event Planner"
	salutation := "dein"
	scope := "Login"

	switch purpose {
	case auth.PurposeOrganizerLogin:
		subject = "Dein Login-Link fuer Easy Event Planner"
		scope = "Organizer-Login"
	case auth.PurposeParticipantLogin:
		subject = "Dein Link fuer das Teilnehmerportal"
		scope = "Teilnehmerportal"
		salutation = "dein"
	case auth.PurposeWaitlistOffer:
		subject = "Dein Wartelisten-Link"
		scope = "Wartelistenangebot"
	case auth.PurposeRegistrationCancel:
		subject = "Dein Storno-Link"
		scope = "Stornierung"
	case auth.PurposeCertificateAccess:
		subject = "Dein Zertifikat-Link"
		scope = "Zertifikatszugriff"
	case auth.PurposeRegistrationVerify:
		subject = "Bestaetige deine Anmeldung"
		scope = "Anmeldebestaetigung"
	}

	expiresAt := message.ExpiresAt.UTC().Format(time.RFC3339)
	verifyURL := strings.TrimSpace(message.VerifyURL)
	tenant := strings.TrimSpace(message.TenantSlug)
	if tenant == "" {
		tenant = "-"
	}

	bodyText = fmt.Sprintf(
		"Hallo,\n\n"+
			"hier ist %s Link fuer den Bereich \"%s\".\n\n"+
			"Tenant: %s\n"+
			"Link: %s\n\n"+
			"Gueltig bis: %s (UTC)\n\n"+
			"Wenn du den Link nicht angefordert hast, kannst du diese E-Mail ignorieren.\n",
		salutation,
		scope,
		tenant,
		verifyURL,
		expiresAt,
	)

	bodyHTML = fmt.Sprintf(
		"<p>Hallo,</p>"+
			"<p>hier ist %s Link fuer den Bereich <strong>%s</strong>.</p>"+
			"<p><strong>Tenant:</strong> %s<br>"+
			"<strong>Link:</strong> <a href=\"%s\">%s</a></p>"+
			"<p><strong>Gueltig bis:</strong> %s (UTC)</p>"+
			"<p>Wenn du den Link nicht angefordert hast, kannst du diese E-Mail ignorieren.</p>",
		escapeHTMLText(salutation),
		escapeHTMLText(scope),
		escapeHTMLText(tenant),
		escapeHTMLText(verifyURL),
		escapeHTMLText(verifyURL),
		escapeHTMLText(expiresAt),
	)

	return subject, bodyText, bodyHTML
}

func buildMagicLinkMetadataJSON(message auth.MagicLinkMessage) string {
	metadata := map[string]any{
		"purpose":     strings.TrimSpace(message.Purpose),
		"tenant_slug": strings.TrimSpace(message.TenantSlug),
		"verify_url":  strings.TrimSpace(message.VerifyURL),
		"expires_at":  message.ExpiresAt.UTC().Format(time.RFC3339),
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(payload)
}

func escapeHTMLText(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(value)
}
