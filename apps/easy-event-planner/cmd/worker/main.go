package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/notification"
)

var buildVersion = "dev"

func main() {
	cfg, err := config.Load(buildVersion)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	sqlDB, err := db.Open(cfg.DBDriver, cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

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
		log.Fatalf("build mailer: %v", err)
	}

	worker := notification.NewWorker(sqlDB, mailer, notification.WorkerConfig{
		PollInterval: cfg.EmailWorkerPollInterval,
		BatchSize:    cfg.EmailWorkerBatchSize,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf(
		"email worker started provider=%s interval=%s batch_size=%d",
		cfg.MailProvider,
		cfg.EmailWorkerPollInterval,
		cfg.EmailWorkerBatchSize,
	)
	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("run email worker: %v", err)
	}
	log.Printf("email worker stopped")
}
