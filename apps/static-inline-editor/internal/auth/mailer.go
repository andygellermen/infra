package auth

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type Mailer interface {
	SendMagicLink(email, verifyURL string) error
}

type mailer struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string
}

func NewMailer(host string, port int, username, password, fromEmail, fromName string) Mailer {
	return &mailer{
		host:      strings.TrimSpace(host),
		port:      port,
		username:  strings.TrimSpace(username),
		password:  password,
		fromEmail: strings.TrimSpace(fromEmail),
		fromName:  strings.TrimSpace(fromName),
	}
}

func (m *mailer) SendMagicLink(email, verifyURL string) error {
	if m.host == "" || m.fromEmail == "" {
		log.Printf("static-inline-editor: magic link for %s: %s", email, verifyURL)
		return nil
	}

	var msg bytes.Buffer
	subject := "Dein Magic Link fuer den Static Inline Editor"
	from := m.fromEmail
	if m.fromName != "" {
		from = fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
	}

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", email))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("\r\n")
	msg.WriteString("Hallo,\r\n\r\n")
	msg.WriteString("hier ist dein Magic Link fuer die Bearbeitung:\r\n")
	msg.WriteString(verifyURL)
	msg.WriteString("\r\n\r\n")
	msg.WriteString("Der Link ist zeitlich begrenzt gueltig.\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var smtpAuth smtp.Auth
	if m.username != "" {
		smtpAuth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	if err := smtp.SendMail(addr, smtpAuth, m.fromEmail, []string{email}, msg.Bytes()); err != nil {
		log.Printf("static-inline-editor: smtp send failed host=%s port=%d from=%s to=%s: %v", m.host, m.port, m.fromEmail, email, err)
		return fmt.Errorf("send smtp mail: %w", err)
	}
	log.Printf("static-inline-editor: magic link sent via smtp host=%s from=%s to=%s", m.host, m.fromEmail, email)
	return nil
}
