package auth

import (
	"context"
	"log"
	"time"
)

type MagicLinkMessage struct {
	ToEmail    string
	TenantSlug string
	Purpose    string
	VerifyURL  string
	ExpiresAt  time.Time
}

type Sender interface {
	SendMagicLink(ctx context.Context, message MagicLinkMessage) error
}

type LogSender struct{}

func (s *LogSender) SendMagicLink(_ context.Context, message MagicLinkMessage) error {
	log.Printf(
		"magic-link prepared tenant=%s email=%s purpose=%s expires_at=%s",
		message.TenantSlug,
		message.ToEmail,
		message.Purpose,
		message.ExpiresAt.Format(time.RFC3339),
	)
	return nil
}
