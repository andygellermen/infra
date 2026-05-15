package certificate

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupCertificateService(t *testing.T) (*Service, *sql.DB, tenant.Tenant) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "certificate-test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	if _, err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tenantRepo := tenant.NewRepository(sqlDB)
	tenantItem, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "cert-demo",
		Name:          "Certificate Demo",
		PublicBaseURL: "https://events.example.com/cert-demo",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	service := NewService(sqlDB, Config{
		BaseURL:     "https://events.example.com",
		TokenPepper: "test-pepper",
		StorageDir:  filepath.Join(t.TempDir(), "certificates"),
	})
	return service, sqlDB, tenantItem
}

func createAttendedRegistrationForCertificateTest(t *testing.T, sqlDB *sql.DB, tenantItem tenant.Tenant, email, name string) (registrationID string, participantID string) {
	t.Helper()

	eventRepo := event.NewRepository(sqlDB)
	createdEvent, err := eventRepo.CreateEvent(context.Background(), tenantItem.ID, event.CreateEventParams{
		Slug:     "cert-event-" + strings.ReplaceAll(strings.ToLower(name), " ", "-"),
		Title:    "Certificate Event " + name,
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	publishedEvent, err := eventRepo.PublishEvent(context.Background(), tenantItem.ID, createdEvent.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}

	regService := registration.NewService(sqlDB, registration.Config{
		BaseURL:         "https://events.example.com",
		TokenPepper:     "test-pepper",
		RegistrationTTL: 30 * time.Minute,
	})

	start, err := regService.Start(context.Background(), registration.StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         publishedEvent.ID,
		Name:            name,
		Email:           email,
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("start registration: %v", err)
	}

	token := extractVerificationTokenForCertificateTest(t, sqlDB, tenantItem.ID, start.RegistrationID)
	if _, err := regService.Verify(context.Background(), registration.VerifyInput{
		TenantID: tenantItem.ID,
		RawToken: token,
	}); err != nil {
		t.Fatalf("verify registration: %v", err)
	}

	attended, err := regService.MarkRegistrationAttended(context.Background(), tenantItem.ID, start.RegistrationID)
	if err != nil {
		t.Fatalf("mark attended: %v", err)
	}
	if attended.Status != registration.StatusAttended {
		t.Fatalf("expected attended status, got %q", attended.Status)
	}

	return start.RegistrationID, start.ParticipantID
}

func extractVerificationTokenForCertificateTest(t *testing.T, sqlDB *sql.DB, tenantID, registrationID string) string {
	t.Helper()

	rows, err := sqlDB.QueryContext(
		context.Background(),
		`SELECT body_text, COALESCE(metadata_json, '')
     FROM email_jobs
     WHERE tenant_id = ? AND template_key = ?
     ORDER BY created_at DESC`,
		tenantID,
		registration.DefaultVerificationTemplate,
	)
	if err != nil {
		t.Fatalf("query verification emails: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bodyText string
		var metadataJSON string
		if err := rows.Scan(&bodyText, &metadataJSON); err != nil {
			t.Fatalf("scan verification email: %v", err)
		}
		if !strings.Contains(metadataJSON, registrationID) {
			continue
		}
		for _, field := range strings.Fields(bodyText) {
			if !strings.Contains(field, "/registrations/verify?token=") {
				continue
			}
			parsed, err := url.Parse(strings.TrimSpace(field))
			if err != nil {
				t.Fatalf("parse verify url: %v", err)
			}
			token := parsed.Query().Get("token")
			if token != "" {
				return token
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate verification emails: %v", err)
	}
	t.Fatalf("no verification token found for registration %q", registrationID)
	return ""
}

func TestIssueForRegistrationAndVerify(t *testing.T) {
	service, _, tenantItem := setupCertificateService(t)
	registrationID, _ := createAttendedRegistrationForCertificateTest(
		t,
		service.db,
		tenantItem,
		"participant1@example.com",
		"Participant One",
	)

	issued, err := service.IssueForRegistration(context.Background(), tenantItem.ID, registrationID)
	if err != nil {
		t.Fatalf("issue certificate: %v", err)
	}
	if issued.AlreadyIssued {
		t.Fatalf("expected new certificate issue, got already issued")
	}
	if issued.Certificate.ID == "" {
		t.Fatalf("expected certificate id")
	}
	if issued.VerificationCode == "" {
		t.Fatalf("expected verification code on first issue")
	}
	if !strings.HasPrefix(issued.Certificate.CertificateNumber, "CERT-") {
		t.Fatalf("expected certificate number with CERT- prefix, got %q", issued.Certificate.CertificateNumber)
	}

	info, statErr := os.Stat(issued.Certificate.FilePath)
	if statErr != nil {
		t.Fatalf("stat certificate file: %v", statErr)
	}
	if info.Size() == 0 {
		t.Fatalf("expected certificate pdf file to be non-empty")
	}

	pdfBytes, readErr := os.ReadFile(issued.Certificate.FilePath)
	if readErr != nil {
		t.Fatalf("read certificate file: %v", readErr)
	}
	if !strings.HasPrefix(string(pdfBytes), "%PDF-1.4") {
		t.Fatalf("expected PDF header, got %q", string(pdfBytes[:minInt(8, len(pdfBytes))]))
	}

	reissued, err := service.IssueForRegistration(context.Background(), tenantItem.ID, registrationID)
	if err != nil {
		t.Fatalf("reissue certificate: %v", err)
	}
	if !reissued.AlreadyIssued {
		t.Fatalf("expected already issued on second issue")
	}
	if reissued.Certificate.ID != issued.Certificate.ID {
		t.Fatalf("expected same certificate id on reissue, got %q vs %q", reissued.Certificate.ID, issued.Certificate.ID)
	}

	verified, err := service.VerifyByNumberAndCode(
		context.Background(),
		tenantItem.ID,
		issued.Certificate.CertificateNumber,
		issued.VerificationCode,
	)
	if err != nil {
		t.Fatalf("verify certificate: %v", err)
	}
	if verified.Certificate.ID != issued.Certificate.ID {
		t.Fatalf("expected verified certificate id %q, got %q", issued.Certificate.ID, verified.Certificate.ID)
	}

	_, err = service.VerifyByNumberAndCode(
		context.Background(),
		tenantItem.ID,
		issued.Certificate.CertificateNumber,
		"wrong-code",
	)
	if !errors.Is(err, ErrInvalidVerificationCode) {
		t.Fatalf("expected ErrInvalidVerificationCode, got %v", err)
	}
}

func TestParticipantCertificateAccessControl(t *testing.T) {
	service, _, tenantItem := setupCertificateService(t)
	registrationA, participantA := createAttendedRegistrationForCertificateTest(
		t,
		service.db,
		tenantItem,
		"participant-a@example.com",
		"Participant A",
	)
	registrationB, participantB := createAttendedRegistrationForCertificateTest(
		t,
		service.db,
		tenantItem,
		"participant-b@example.com",
		"Participant B",
	)

	issuedA, err := service.IssueForRegistration(context.Background(), tenantItem.ID, registrationA)
	if err != nil {
		t.Fatalf("issue certificate A: %v", err)
	}
	_, err = service.IssueForRegistration(context.Background(), tenantItem.ID, registrationB)
	if err != nil {
		t.Fatalf("issue certificate B: %v", err)
	}

	itemsA, err := service.ListParticipantCertificates(context.Background(), tenantItem.ID, participantA)
	if err != nil {
		t.Fatalf("list participant A certificates: %v", err)
	}
	if len(itemsA) != 1 {
		t.Fatalf("expected one certificate for participant A, got %d", len(itemsA))
	}
	if itemsA[0].ID != issuedA.Certificate.ID {
		t.Fatalf("expected participant A certificate id %q, got %q", issuedA.Certificate.ID, itemsA[0].ID)
	}

	_, bytes, err := service.LoadParticipantCertificatePDF(context.Background(), tenantItem.ID, participantA, issuedA.Certificate.ID)
	if err != nil {
		t.Fatalf("load participant A certificate pdf: %v", err)
	}
	if len(bytes) == 0 {
		t.Fatalf("expected participant A certificate bytes")
	}

	_, _, err = service.LoadParticipantCertificatePDF(context.Background(), tenantItem.ID, participantB, issuedA.Certificate.ID)
	if !errors.Is(err, ErrCertificateAccessDenied) {
		t.Fatalf("expected ErrCertificateAccessDenied for participant B, got %v", err)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
