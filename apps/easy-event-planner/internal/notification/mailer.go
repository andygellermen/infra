package notification

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

const (
	ProviderLog  = "log"
	ProviderSMTP = "smtp"
	ProviderSES  = "ses"
)

type Message struct {
	TenantID     string
	TemplateKey  string
	Recipient    string
	Subject      string
	BodyText     string
	BodyHTML     string
	MetadataJSON string
	FromEmail    string
	FromName     string
}

type Mailer interface {
	Send(ctx context.Context, message Message) error
}

type MailerConfig struct {
	Provider  string
	FromEmail string
	FromName  string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func NewMailer(cfg MailerConfig) (Mailer, error) {
	provider := normalizeProvider(cfg.Provider)
	switch provider {
	case ProviderLog:
		return &LogMailer{
			DefaultFromEmail: strings.TrimSpace(cfg.FromEmail),
			DefaultFromName:  strings.TrimSpace(cfg.FromName),
		}, nil
	case ProviderSMTP, ProviderSES:
		return newSMTPMailer(provider, cfg)
	default:
		return nil, fmt.Errorf("unsupported mail provider %q", provider)
	}
}

type LogMailer struct {
	DefaultFromEmail string
	DefaultFromName  string
}

func (m *LogMailer) Send(_ context.Context, message Message) error {
	recipient, err := parseAddress(strings.TrimSpace(message.Recipient))
	if err != nil {
		return fmt.Errorf("parse recipient address: %w", err)
	}

	fromEmail := strings.TrimSpace(message.FromEmail)
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(m.DefaultFromEmail)
	}
	fromName := strings.TrimSpace(message.FromName)
	if fromName == "" {
		fromName = strings.TrimSpace(m.DefaultFromName)
	}

	log.Printf(
		"email prepared provider=log tenant=%s template=%s to=%s from=%s subject=%q text_len=%d html_len=%d",
		strings.TrimSpace(message.TenantID),
		strings.TrimSpace(message.TemplateKey),
		recipient.Address,
		formatAddress(fromName, fromEmail),
		sanitizeHeaderValue(message.Subject),
		len(message.BodyText),
		len(message.BodyHTML),
	)
	return nil
}

type SMTPMailer struct {
	provider  string
	fromEmail string
	fromName  string
	host      string
	port      int
	user      string
	pass      string
}

func newSMTPMailer(provider string, cfg MailerConfig) (*SMTPMailer, error) {
	host := strings.TrimSpace(cfg.SMTPHost)
	if host == "" {
		return nil, fmt.Errorf("smtp host must not be empty")
	}
	if cfg.SMTPPort <= 0 {
		return nil, fmt.Errorf("smtp port must be > 0")
	}
	fromEmail := strings.TrimSpace(cfg.FromEmail)
	if fromEmail == "" {
		return nil, fmt.Errorf("mail from address must not be empty")
	}
	if _, err := parseAddress(fromEmail); err != nil {
		return nil, fmt.Errorf("parse mail from address: %w", err)
	}

	user := strings.TrimSpace(cfg.SMTPUser)
	pass := strings.TrimSpace(cfg.SMTPPass)
	if (user == "") != (pass == "") {
		return nil, fmt.Errorf("smtp user and pass must be set together")
	}

	return &SMTPMailer{
		provider:  provider,
		fromEmail: fromEmail,
		fromName:  strings.TrimSpace(cfg.FromName),
		host:      host,
		port:      cfg.SMTPPort,
		user:      user,
		pass:      pass,
	}, nil
}

func (m *SMTPMailer) Send(_ context.Context, message Message) error {
	recipient, err := parseAddress(strings.TrimSpace(message.Recipient))
	if err != nil {
		return fmt.Errorf("parse recipient address: %w", err)
	}

	fromEmail := strings.TrimSpace(message.FromEmail)
	if fromEmail == "" {
		fromEmail = m.fromEmail
	}
	if _, err := parseAddress(fromEmail); err != nil {
		return fmt.Errorf("parse sender address: %w", err)
	}

	fromName := strings.TrimSpace(message.FromName)
	if fromName == "" {
		fromName = m.fromName
	}

	msg, err := buildRFC822Message(
		formatAddress(fromName, fromEmail),
		recipient.String(),
		sanitizeHeaderValue(message.Subject),
		message.BodyText,
		message.BodyHTML,
	)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if m.user != "" && m.pass != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}

	addr := m.host + ":" + strconv.Itoa(m.port)
	if err := smtp.SendMail(addr, auth, fromEmail, []string{recipient.Address}, msg); err != nil {
		return fmt.Errorf("send smtp mail via provider=%s: %w", m.provider, err)
	}
	return nil
}

func buildRFC822Message(from, to, subject, bodyText, bodyHTML string) ([]byte, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	subject = sanitizeHeaderValue(subject)
	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to headers must not be empty")
	}

	text := normalizeBody(bodyText)
	html := normalizeBody(bodyHTML)

	var builder strings.Builder
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("From: " + from + "\r\n")
	builder.WriteString("To: " + to + "\r\n")
	builder.WriteString("Subject: " + subject + "\r\n")

	switch {
	case text != "" && html != "":
		boundary := "eep-alt-" + strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
		builder.WriteString(`Content-Type: multipart/alternative; boundary="` + boundary + "\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString("--" + boundary + "\r\n")
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(text + "\r\n")
		builder.WriteString("--" + boundary + "\r\n")
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(html + "\r\n")
		builder.WriteString("--" + boundary + "--\r\n")
	case html != "":
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(html + "\r\n")
	default:
		if text == "" {
			text = "(no content)"
		}
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(text + "\r\n")
	}

	return []byte(builder.String()), nil
}

func normalizeProvider(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ProviderLog
	}
	return value
}

func parseAddress(raw string) (*mail.Address, error) {
	return mail.ParseAddress(strings.TrimSpace(raw))
}

func formatAddress(name, email string) string {
	addr, err := parseAddress(email)
	if err != nil {
		return strings.TrimSpace(email)
	}
	normalized := &mail.Address{
		Name:    strings.TrimSpace(name),
		Address: addr.Address,
	}
	return normalized.String()
}

func normalizeBody(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.ReplaceAll(trimmed, "\r\n", "\n")
	trimmed = strings.ReplaceAll(trimmed, "\r", "\n")
	return strings.ReplaceAll(trimmed, "\n", "\r\n")
}

func sanitizeHeaderValue(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.Join(strings.Fields(value), " ")
}
