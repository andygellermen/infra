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

type SeedInput struct {
	Slug            string
	Name            string
	PublicBaseURL   string
	DefaultTimezone string
	DefaultLocale   string
	Status          string
	Settings        TenantSettingsInput
}

type SeedResult struct {
	Tenant   Tenant
	Settings TenantSettings
	Created  bool
}
