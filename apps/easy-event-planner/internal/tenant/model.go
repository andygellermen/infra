package tenant

import "time"

const (
	DefaultTimezone      = "Europe/Berlin"
	DefaultLocale        = "de-DE"
	DefaultStatus        = "active"
	DefaultPayPalMode    = "disabled"
	DefaultRetentionDays = 30
)

type Tenant struct {
	ID              string
	Slug            string
	Name            string
	PublicBaseURL   string
	DefaultTimezone string
	DefaultLocale   string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TenantSettings struct {
	TenantID             string
	SenderEmail          string
	SenderName           string
	PayPalMode           string
	PayPalClientID       string
	PayPalMerchantID     string
	DefaultRetentionDays int
	SettingsJSON         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type TenantSettingsInput struct {
	SenderEmail          string
	SenderName           string
	PayPalMode           string
	PayPalClientID       string
	PayPalMerchantID     string
	DefaultRetentionDays int
	SettingsJSON         string
}

const (
	DomainBindingStatusPendingDNS  = "pending_dns"
	DomainBindingStatusDNSVerified = "dns_verified"
	DomainBindingStatusSSLPending  = "ssl_pending"
	DomainBindingStatusActive      = "active"
	DomainBindingStatusDisabled    = "disabled"

	DomainBindingSSLStatusPending = "pending"
	DomainBindingSSLStatusValid   = "valid"
	DomainBindingSSLStatusInvalid = "invalid"
	DomainBindingSSLStatusExpired = "expired"
)

type TenantDomainBinding struct {
	ID                      string
	TenantID                string
	Domain                  string
	BasePath                string
	Status                  string
	IsPrimary               bool
	VerificationToken       string
	DNSVerifiedAt           *time.Time
	RoutingVerifiedAt       *time.Time
	LastDNSCheckAt          *time.Time
	LastDNSError            string
	LastRoutingCheckAt      *time.Time
	LastRoutingError        string
	SSLStatus               string
	SSLCertificateIssuer    string
	SSLCertificateExpiresAt *time.Time
	LastSSLCheckAt          *time.Time
	LastSSLError            string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type CreateTenantDomainBindingParams struct {
	TenantID  string
	Domain    string
	BasePath  string
	Status    string
	IsPrimary bool
}

type UpdateTenantDomainBindingParams struct {
	Domain    *string
	BasePath  *string
	Status    *string
	IsPrimary *bool
}

type PublicRouteMatch struct {
	Tenant   Tenant
	BaseURL  string
	BasePath string
	Source   string
}

type DomainCheckResult struct {
	VerificationRecordName  string
	VerificationRecordValue string
	DNSVerified             bool
	RoutingVerified         bool
	LastDNSCheckAt          *time.Time
	LastDNSError            string
	LastRoutingCheckAt      *time.Time
	LastRoutingError        string
	SSLStatus               string
	SSLCertificateIssuer    string
	SSLCertificateExpiresAt *time.Time
	LastSSLCheckAt          *time.Time
	LastSSLError            string
	Status                  string
}

type CreateTenantParams struct {
	ID              string
	Slug            string
	Name            string
	PublicBaseURL   string
	DefaultTimezone string
	DefaultLocale   string
	Status          string
	Settings        TenantSettingsInput
}

type UpsertTenantSettingsParams struct {
	TenantID string
	Settings TenantSettingsInput
}

type UpdateTenantParams struct {
	Name            *string
	PublicBaseURL   *string
	DefaultTimezone *string
	DefaultLocale   *string
}

type SeedAdminUserInput struct {
	Email  string
	Name   string
	Role   string
	Status string
}

type SeedInput struct {
	Slug            string
	Name            string
	PublicBaseURL   string
	DefaultTimezone string
	DefaultLocale   string
	Status          string
	Settings        TenantSettingsInput
	AdminUser       SeedAdminUserInput
}

type SeedResult struct {
	Tenant   Tenant
	Settings TenantSettings
	Created  bool
}
