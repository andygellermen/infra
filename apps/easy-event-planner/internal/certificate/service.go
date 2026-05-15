package certificate

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	CertificateStatusIssued = "issued"
)

var (
	ErrRegistrationNotFound      = errors.New("registration not found")
	ErrCertificateNotFound       = errors.New("certificate not found")
	ErrCertificateAccessDenied   = errors.New("certificate access denied")
	ErrCertificateEligibility    = errors.New("registration is not eligible for certificate")
	ErrInvalidVerificationCode   = errors.New("invalid certificate verification code")
	ErrCertificateStorageFailure = errors.New("certificate storage failure")
)

type Config struct {
	BaseURL     string
	TokenPepper string
	StorageDir  string
}

type Service struct {
	db    *sql.DB
	cfg   Config
	nowFn func() time.Time
	idFn  func(prefix string) string
	tokFn func() (string, error)
}

type Certificate struct {
	ID                   string
	TenantID             string
	TenantSlug           string
	RegistrationID       string
	ParticipantID        string
	ParticipantName      string
	ParticipantEmail     string
	EventID              string
	EventSlug            string
	EventTitle           string
	EventStartsAt        time.Time
	EventTimezone        string
	CertificateNumber    string
	Status               string
	VerificationCodeHint string
	IssuedAt             time.Time
	AttendedAt           *time.Time
	RevokedAt            *time.Time
	FilePath             string
	FileSHA256           string
}

type IssueResult struct {
	Certificate      Certificate
	VerificationCode string
	AlreadyIssued    bool
}

type VerifyResult struct {
	Certificate Certificate
	VerifiedAt  time.Time
}

type issueRegistrationRecord struct {
	TenantID         string
	TenantSlug       string
	RegistrationID   string
	ParticipantID    string
	ParticipantName  string
	ParticipantEmail string
	EventID          string
	EventSlug        string
	EventTitle       string
	EventStartsAt    time.Time
	EventTimezone    string
	Status           string
	AttendedAt       *time.Time
}

func NewService(sqlDB *sql.DB, cfg Config) *Service {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	tokenPepper := strings.TrimSpace(cfg.TokenPepper)
	if tokenPepper == "" {
		tokenPepper = "dev-only-change-me"
	}
	storageDir := strings.TrimSpace(cfg.StorageDir)
	if storageDir == "" {
		storageDir = "certificates"
	}
	return &Service{
		db: sqlDB,
		cfg: Config{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			TokenPepper: tokenPepper,
			StorageDir:  storageDir,
		},
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
		tokFn: randomToken,
	}
}

func (s *Service) IssueForRegistration(ctx context.Context, tenantID, registrationID string) (IssueResult, error) {
	if s.db == nil {
		return IssueResult{}, fmt.Errorf("certificate service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	registrationID = strings.TrimSpace(registrationID)
	if tenantID == "" || registrationID == "" {
		return IssueResult{}, ErrRegistrationNotFound
	}

	existing, err := s.GetByRegistration(ctx, tenantID, registrationID)
	if err == nil {
		return IssueResult{
			Certificate:   existing,
			AlreadyIssued: true,
		}, nil
	}
	if !errors.Is(err, ErrCertificateNotFound) {
		return IssueResult{}, err
	}

	record, err := s.lookupIssueRegistration(ctx, tenantID, registrationID)
	if err != nil {
		return IssueResult{}, err
	}
	if !isEligibleForCertificate(record) {
		return IssueResult{}, ErrCertificateEligibility
	}

	now := s.nowFn().UTC()

	var lastInsertErr error
	for attempt := 0; attempt < 5; attempt++ {
		certificateID := s.idFn("crt")
		certificateNumber := buildCertificateNumber(now, s.idFn("seq"))
		verificationCode, tokenErr := s.tokFn()
		if tokenErr != nil {
			return IssueResult{}, fmt.Errorf("generate certificate verification token: %w", tokenErr)
		}
		verificationHash := s.hashVerificationCode(verificationCode)
		verificationHint := verificationCodeHint(verificationCode)

		pdfBytes := s.buildCertificatePDF(record, certificateNumber, verificationCode)
		fileSHA := sha256Hex(pdfBytes)
		filePath := s.certificateFilePath(record.TenantSlug, certificateNumber)
		if err := writeFileAtomic(filePath, pdfBytes); err != nil {
			return IssueResult{}, fmt.Errorf("%w: %v", ErrCertificateStorageFailure, err)
		}

		_, err := s.db.ExecContext(
			ctx,
			`INSERT INTO certificates (
        id, tenant_id, registration_id, participant_id, event_id, certificate_number,
        status, issued_at, attended_at, file_path, file_sha256, verification_code_hash,
        verification_code_hint, revoked_at, created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
			certificateID,
			record.TenantID,
			record.RegistrationID,
			record.ParticipantID,
			record.EventID,
			certificateNumber,
			CertificateStatusIssued,
			now.Format(time.RFC3339),
			formatOptionalTime(record.AttendedAt),
			filePath,
			fileSHA,
			verificationHash,
			verificationHint,
			now.Format(time.RFC3339),
			now.Format(time.RFC3339),
		)
		if err != nil {
			_ = os.Remove(filePath)
			if isRegistrationCertificateConflict(err) {
				existing, existingErr := s.GetByRegistration(ctx, tenantID, registrationID)
				if existingErr == nil {
					return IssueResult{
						Certificate:   existing,
						AlreadyIssued: true,
					}, nil
				}
			}
			if isUniqueConstraint(err) {
				lastInsertErr = err
				continue
			}
			return IssueResult{}, fmt.Errorf("insert certificate: %w", err)
		}

		created, createdErr := s.GetByRegistration(ctx, tenantID, registrationID)
		if createdErr != nil {
			return IssueResult{}, createdErr
		}
		return IssueResult{
			Certificate:      created,
			VerificationCode: verificationCode,
			AlreadyIssued:    false,
		}, nil
	}

	if lastInsertErr != nil {
		return IssueResult{}, fmt.Errorf("insert certificate after retries: %w", lastInsertErr)
	}
	return IssueResult{}, fmt.Errorf("insert certificate failed")
}

func (s *Service) GetByRegistration(ctx context.Context, tenantID, registrationID string) (Certificate, error) {
	tenantID = strings.TrimSpace(tenantID)
	registrationID = strings.TrimSpace(registrationID)
	if tenantID == "" || registrationID == "" {
		return Certificate{}, ErrCertificateNotFound
	}

	row := s.db.QueryRowContext(
		ctx,
		selectCertificateBase+`
 WHERE c.tenant_id = ? AND c.registration_id = ?
 LIMIT 1`,
		tenantID,
		registrationID,
	)
	return scanCertificate(row)
}

func (s *Service) GetByID(ctx context.Context, tenantID, certificateID string) (Certificate, error) {
	tenantID = strings.TrimSpace(tenantID)
	certificateID = strings.TrimSpace(certificateID)
	if tenantID == "" || certificateID == "" {
		return Certificate{}, ErrCertificateNotFound
	}
	row := s.db.QueryRowContext(
		ctx,
		selectCertificateBase+`
 WHERE c.tenant_id = ? AND c.id = ?
 LIMIT 1`,
		tenantID,
		certificateID,
	)
	return scanCertificate(row)
}

func (s *Service) ListParticipantCertificates(ctx context.Context, tenantID, participantID string) ([]Certificate, error) {
	tenantID = strings.TrimSpace(tenantID)
	participantID = strings.TrimSpace(participantID)
	if tenantID == "" || participantID == "" {
		return nil, ErrCertificateAccessDenied
	}

	rows, err := s.db.QueryContext(
		ctx,
		selectCertificateBase+`
 WHERE c.tenant_id = ? AND c.participant_id = ?
 ORDER BY c.issued_at DESC, c.created_at DESC`,
		tenantID,
		participantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list participant certificates: %w", err)
	}
	defer rows.Close()

	items := make([]Certificate, 0)
	for rows.Next() {
		item, scanErr := scanCertificate(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate participant certificates: %w", err)
	}
	return items, nil
}

func (s *Service) GetParticipantCertificate(ctx context.Context, tenantID, participantID, certificateID string) (Certificate, error) {
	item, err := s.GetByID(ctx, tenantID, certificateID)
	if err != nil {
		return Certificate{}, err
	}
	if strings.TrimSpace(item.ParticipantID) != strings.TrimSpace(participantID) {
		return Certificate{}, ErrCertificateAccessDenied
	}
	return item, nil
}

func (s *Service) LoadParticipantCertificatePDF(ctx context.Context, tenantID, participantID, certificateID string) (Certificate, []byte, error) {
	item, err := s.GetParticipantCertificate(ctx, tenantID, participantID, certificateID)
	if err != nil {
		return Certificate{}, nil, err
	}
	bytes, err := os.ReadFile(item.FilePath)
	if err != nil {
		return Certificate{}, nil, fmt.Errorf("%w: %v", ErrCertificateStorageFailure, err)
	}
	return item, bytes, nil
}

func (s *Service) VerifyByNumberAndCode(ctx context.Context, tenantID, certificateNumber, verificationCode string) (VerifyResult, error) {
	tenantID = strings.TrimSpace(tenantID)
	certificateNumber = strings.TrimSpace(certificateNumber)
	verificationCode = strings.TrimSpace(verificationCode)
	if tenantID == "" || certificateNumber == "" || verificationCode == "" {
		return VerifyResult{}, ErrInvalidVerificationCode
	}

	row := s.db.QueryRowContext(
		ctx,
		selectCertificateWithVerification+`
 WHERE c.tenant_id = ? AND c.certificate_number = ?
 LIMIT 1`,
		tenantID,
		certificateNumber,
	)

	item, storedHash, err := scanCertificateWithVerificationHash(row)
	if err != nil {
		if errors.Is(err, ErrCertificateNotFound) {
			return VerifyResult{}, ErrInvalidVerificationCode
		}
		return VerifyResult{}, err
	}

	expectedHash := s.hashVerificationCode(verificationCode)
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(expectedHash)) != 1 {
		return VerifyResult{}, ErrInvalidVerificationCode
	}

	return VerifyResult{
		Certificate: item,
		VerifiedAt:  s.nowFn().UTC(),
	}, nil
}

func (s *Service) VerificationURL(tenantSlug, certificateNumber, verificationCode string) string {
	tenantSlug = strings.TrimSpace(tenantSlug)
	certificateNumber = strings.TrimSpace(certificateNumber)
	verificationCode = strings.TrimSpace(verificationCode)
	if tenantSlug == "" || certificateNumber == "" || verificationCode == "" {
		return ""
	}
	return fmt.Sprintf(
		"%s/api/v1/public/%s/certificates/verify?certificate_no=%s&code=%s",
		s.cfg.BaseURL,
		url.PathEscape(tenantSlug),
		url.QueryEscape(certificateNumber),
		url.QueryEscape(verificationCode),
	)
}

func (s *Service) hashVerificationCode(raw string) string {
	sum := sha256.Sum256([]byte(s.cfg.TokenPepper + ":" + strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) certificateFilePath(tenantSlug, certificateNumber string) string {
	slugDir := sanitizePathSegment(tenantSlug)
	if slugDir == "" {
		slugDir = "tenant"
	}
	fileName := sanitizePathSegment(strings.ToLower(certificateNumber))
	if fileName == "" {
		fileName = "certificate"
	}
	return filepath.Join(s.cfg.StorageDir, slugDir, fileName+".pdf")
}

func (s *Service) lookupIssueRegistration(ctx context.Context, tenantID, registrationID string) (issueRegistrationRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT r.tenant_id, COALESCE(t.slug, ''), r.id, r.participant_id,
            COALESCE(p.name, ''), COALESCE(p.email, ''),
            r.event_id, COALESCE(e.slug, ''), COALESCE(e.title, ''), COALESCE(e.starts_at, ''), COALESCE(e.timezone, 'UTC'),
            COALESCE(r.status, ''), COALESCE(r.attended_at, '')
     FROM registrations r
     JOIN tenants t ON t.id = r.tenant_id
     LEFT JOIN participants p ON p.id = r.participant_id
     LEFT JOIN events e ON e.id = r.event_id
     WHERE r.tenant_id = ? AND r.id = ?
     LIMIT 1`,
		tenantID,
		registrationID,
	)

	var (
		record        issueRegistrationRecord
		eventStartsAt string
		attendedAtRaw string
	)
	if err := row.Scan(
		&record.TenantID,
		&record.TenantSlug,
		&record.RegistrationID,
		&record.ParticipantID,
		&record.ParticipantName,
		&record.ParticipantEmail,
		&record.EventID,
		&record.EventSlug,
		&record.EventTitle,
		&eventStartsAt,
		&record.EventTimezone,
		&record.Status,
		&attendedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return issueRegistrationRecord{}, ErrRegistrationNotFound
		}
		return issueRegistrationRecord{}, fmt.Errorf("query certificate registration: %w", err)
	}

	startsAt, err := parseTime(eventStartsAt)
	if err != nil {
		return issueRegistrationRecord{}, fmt.Errorf("parse certificate event starts_at: %w", err)
	}
	record.EventStartsAt = startsAt.UTC()
	record.AttendedAt, err = parseOptionalTime(attendedAtRaw)
	if err != nil {
		return issueRegistrationRecord{}, fmt.Errorf("parse certificate attended_at: %w", err)
	}

	return record, nil
}

func (s *Service) buildCertificatePDF(record issueRegistrationRecord, certificateNumber, verificationCode string) []byte {
	participantName := strings.TrimSpace(record.ParticipantName)
	if participantName == "" {
		participantName = strings.TrimSpace(record.ParticipantEmail)
	}
	if participantName == "" {
		participantName = "Teilnehmer"
	}
	eventDate := formatEventDate(record.EventStartsAt, record.EventTimezone)
	issuedAt := s.nowFn().UTC().Format("2006-01-02 15:04 UTC")
	verificationURL := s.VerificationURL(record.TenantSlug, certificateNumber, verificationCode)

	lines := []string{
		"Teilnahmebescheinigung",
		"",
		"Hiermit bestaetigen wir die Teilnahme von",
		participantName,
		"",
		"an der Veranstaltung:",
		record.EventTitle,
		"",
		"Termin: " + eventDate,
		"Zertifikatsnummer: " + certificateNumber,
		"Ausgestellt am: " + issuedAt,
		"",
		"Verifikation:",
		verificationURL,
	}

	content := renderPDFTextBlock(lines)
	return buildSimplePDF(content)
}

func renderPDFTextBlock(lines []string) string {
	var builder strings.Builder
	builder.WriteString("BT\n")
	builder.WriteString("/F1 20 Tf\n")
	builder.WriteString("72 790 Td\n")
	for idx, line := range lines {
		if idx == 1 {
			builder.WriteString("0 -12 Td\n")
			continue
		}
		if idx == 2 {
			builder.WriteString("/F1 12 Tf\n")
			builder.WriteString("0 -26 Td\n")
		}
		text := asciiSafe(strings.TrimSpace(line))
		if idx > 0 && idx != 2 {
			builder.WriteString("0 -18 Td\n")
		}
		builder.WriteString("(" + escapePDFText(text) + ") Tj\n")
	}
	builder.WriteString("ET\n")
	return builder.String()
}

func buildSimplePDF(content string) []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)
	for idx, obj := range objects {
		offsets[idx+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", idx+1, obj)
	}

	xrefStart := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for idx := 1; idx <= len(objects); idx++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[idx])
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefStart)
	return out.Bytes()
}

func isEligibleForCertificate(record issueRegistrationRecord) bool {
	status := strings.ToLower(strings.TrimSpace(record.Status))
	if status != "attended" {
		return false
	}
	return record.AttendedAt != nil
}

const selectCertificateBase = `SELECT c.id, c.tenant_id, COALESCE(t.slug, ''), c.registration_id, c.participant_id,
       COALESCE(p.name, ''), COALESCE(p.email, ''),
       c.event_id, COALESCE(e.slug, ''), COALESCE(e.title, ''), COALESCE(e.starts_at, ''), COALESCE(e.timezone, 'UTC'),
       c.certificate_number, c.status, COALESCE(c.verification_code_hint, ''), c.issued_at, COALESCE(c.attended_at, ''), COALESCE(c.revoked_at, ''),
       COALESCE(c.file_path, ''), COALESCE(c.file_sha256, '')
FROM certificates c
LEFT JOIN tenants t ON t.id = c.tenant_id
LEFT JOIN participants p ON p.id = c.participant_id
LEFT JOIN events e ON e.id = c.event_id`

const selectCertificateWithVerification = `SELECT c.id, c.tenant_id, COALESCE(t.slug, ''), c.registration_id, c.participant_id,
       COALESCE(p.name, ''), COALESCE(p.email, ''),
       c.event_id, COALESCE(e.slug, ''), COALESCE(e.title, ''), COALESCE(e.starts_at, ''), COALESCE(e.timezone, 'UTC'),
       c.certificate_number, c.status, COALESCE(c.verification_code_hint, ''), c.issued_at, COALESCE(c.attended_at, ''), COALESCE(c.revoked_at, ''),
       COALESCE(c.file_path, ''), COALESCE(c.file_sha256, ''), c.verification_code_hash
FROM certificates c
LEFT JOIN tenants t ON t.id = c.tenant_id
LEFT JOIN participants p ON p.id = c.participant_id
LEFT JOIN events e ON e.id = c.event_id`

func scanCertificate(row interface{ Scan(dest ...any) error }) (Certificate, error) {
	var (
		item          Certificate
		eventStartsAt string
		issuedAtRaw   string
		attendedRaw   string
		revokedRaw    string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.TenantSlug,
		&item.RegistrationID,
		&item.ParticipantID,
		&item.ParticipantName,
		&item.ParticipantEmail,
		&item.EventID,
		&item.EventSlug,
		&item.EventTitle,
		&eventStartsAt,
		&item.EventTimezone,
		&item.CertificateNumber,
		&item.Status,
		&item.VerificationCodeHint,
		&issuedAtRaw,
		&attendedRaw,
		&revokedRaw,
		&item.FilePath,
		&item.FileSHA256,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Certificate{}, ErrCertificateNotFound
		}
		return Certificate{}, fmt.Errorf("scan certificate: %w", err)
	}

	startsAt, err := parseTime(eventStartsAt)
	if err != nil {
		return Certificate{}, fmt.Errorf("parse certificate event starts_at: %w", err)
	}
	item.EventStartsAt = startsAt.UTC()
	item.IssuedAt, err = parseTime(issuedAtRaw)
	if err != nil {
		return Certificate{}, fmt.Errorf("parse certificate issued_at: %w", err)
	}
	item.IssuedAt = item.IssuedAt.UTC()
	item.AttendedAt, err = parseOptionalTime(attendedRaw)
	if err != nil {
		return Certificate{}, fmt.Errorf("parse certificate attended_at: %w", err)
	}
	item.RevokedAt, err = parseOptionalTime(revokedRaw)
	if err != nil {
		return Certificate{}, fmt.Errorf("parse certificate revoked_at: %w", err)
	}
	return item, nil
}

func scanCertificateWithVerificationHash(row interface{ Scan(dest ...any) error }) (Certificate, string, error) {
	var (
		item          Certificate
		eventStartsAt string
		issuedAtRaw   string
		attendedRaw   string
		revokedRaw    string
		hash          string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.TenantSlug,
		&item.RegistrationID,
		&item.ParticipantID,
		&item.ParticipantName,
		&item.ParticipantEmail,
		&item.EventID,
		&item.EventSlug,
		&item.EventTitle,
		&eventStartsAt,
		&item.EventTimezone,
		&item.CertificateNumber,
		&item.Status,
		&item.VerificationCodeHint,
		&issuedAtRaw,
		&attendedRaw,
		&revokedRaw,
		&item.FilePath,
		&item.FileSHA256,
		&hash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Certificate{}, "", ErrCertificateNotFound
		}
		return Certificate{}, "", fmt.Errorf("scan certificate verification row: %w", err)
	}

	startsAt, err := parseTime(eventStartsAt)
	if err != nil {
		return Certificate{}, "", fmt.Errorf("parse certificate event starts_at: %w", err)
	}
	item.EventStartsAt = startsAt.UTC()
	item.IssuedAt, err = parseTime(issuedAtRaw)
	if err != nil {
		return Certificate{}, "", fmt.Errorf("parse certificate issued_at: %w", err)
	}
	item.IssuedAt = item.IssuedAt.UTC()
	item.AttendedAt, err = parseOptionalTime(attendedRaw)
	if err != nil {
		return Certificate{}, "", fmt.Errorf("parse certificate attended_at: %w", err)
	}
	item.RevokedAt, err = parseOptionalTime(revokedRaw)
	if err != nil {
		return Certificate{}, "", fmt.Errorf("parse certificate revoked_at: %w", err)
	}
	return item, strings.TrimSpace(hash), nil
}

func parseTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("time value is empty")
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func parseOptionalTime(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseTime(value)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func formatEventDate(startsAt time.Time, timezone string) string {
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	return startsAt.In(loc).Format("2006-01-02 15:04 MST")
}

func buildCertificateNumber(now time.Time, suffixSeed string) string {
	suffix := strings.ToUpper(sanitizePathSegment(suffixSeed))
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	if suffix == "" {
		suffix = "00000000"
	}
	return fmt.Sprintf("CERT-%d-%s", now.Year(), suffix)
}

func verificationCodeHint(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 6 {
		return trimmed
	}
	return "***" + trimmed[len(trimmed)-6:]
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sanitizePathSegment(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteRune('_')
		}
	}
	return strings.Trim(builder.String(), "._")
}

func writeFileAtomic(path string, content []byte) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is empty")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create certificate directory: %w", err)
	}

	tmpName := fmt.Sprintf(".tmp-%d-%d", time.Now().UTC().UnixNano(), len(content))
	tmpPath := filepath.Join(dir, tmpName)
	if err := os.WriteFile(tmpPath, content, 0o644); err != nil {
		return fmt.Errorf("write temporary certificate file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temporary certificate file: %w", err)
	}
	return nil
}

func escapePDFText(raw string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		"(", `\\(`,
		")", `\\)`,
		"\n", " ",
		"\r", " ",
	)
	return replacer.Replace(strings.TrimSpace(raw))
}

func asciiSafe(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range raw {
		if r >= 32 && r <= 126 {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('?')
	}
	return builder.String()
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed")
}

func isRegistrationCertificateConflict(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "certificates.tenant_id") && strings.Contains(message, "certificates.registration_id")
}

func defaultID(prefix string) string {
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, randomBytes)
}

func randomToken() (string, error) {
	var randomBytes [24]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes[:]), nil
}
