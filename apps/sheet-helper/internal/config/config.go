package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TenantConfig struct {
	Domain          string
	CookieSecret    string
	SyncToken       string
	SheetID         string
	PublishedURL    string
	RoutesSheet     string
	VCardsSheet     string
	TextsSheet      string
	DefaultListPref string
	StartupSync     bool
}

type Config struct {
	Addr      string
	DBPath    string
	SeedFile  string
	TenantDir string
	Tenants   map[string]TenantConfig
}

func Load() (Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("getwd: %w", err)
	}

	cfg := Config{
		Addr:      getenv("SHEET_HELPER_ADDR", ":8080"),
		DBPath:    getenv("SHEET_HELPER_DB_PATH", filepath.Join(cwd, "sheet-helper.db")),
		SeedFile:  getenv("SHEET_HELPER_SEED_FILE", filepath.Join(cwd, "testdata", "seed.json")),
		TenantDir: strings.TrimSpace(getenv("SHEET_HELPER_TENANT_DIR", "")),
		Tenants:   map[string]TenantConfig{},
	}

	if cfg.TenantDir != "" {
		tenants, err := loadTenantsFromDir(cfg.TenantDir)
		if err != nil {
			return Config{}, err
		}
		cfg.Tenants = tenants
		if len(cfg.Tenants) > 0 {
			// In multi-tenant runtime mode we should never fall back to the
			// baked-in seed file from the container image, otherwise successful
			// sheet sync data gets overwritten with localhost demo content.
			cfg.SeedFile = ""
			return cfg, nil
		}
	}

	tenant := loadTenantFromEnv()
	cfg.Tenants[tenant.Domain] = tenant
	if tenant.SheetID != "" {
		cfg.SeedFile = strings.TrimSpace(getenv("SHEET_HELPER_SEED_FILE", ""))
	}

	return cfg, nil
}

func loadTenantsFromDir(dir string) (map[string]TenantConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read tenant dir %s: %w", dir, err)
	}

	tenants := make(map[string]TenantConfig)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		values, err := parseEnvFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse tenant env %s: %w", path, err)
		}

		domain := strings.TrimSpace(values["SHEET_HELPER_TENANT"])
		if domain == "" {
			domain = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}
		if domain == "" {
			continue
		}

		tenant := TenantConfig{
			Domain:          domain,
			CookieSecret:    firstNonEmpty(values["SHEET_HELPER_COOKIE_SECRET"], "dev-only-change-me"),
			SyncToken:       strings.TrimSpace(values["SHEET_HELPER_SYNC_TOKEN"]),
			SheetID:         strings.TrimSpace(values["SHEET_HELPER_SHEET_ID"]),
			PublishedURL:    strings.TrimSpace(values["SHEET_HELPER_PUBLISHED_URL"]),
			RoutesSheet:     firstNonEmpty(values["SHEET_HELPER_ROUTES_SHEET"], "routes"),
			VCardsSheet:     firstNonEmpty(values["SHEET_HELPER_VCARDS_SHEET"], "vcard_entries"),
			TextsSheet:      firstNonEmpty(values["SHEET_HELPER_TEXTS_SHEET"], "text_entries"),
			DefaultListPref: firstNonEmpty(values["SHEET_HELPER_LIST_PREFIX"], "list_"),
			StartupSync:     parseBool(values["SHEET_HELPER_STARTUP_SYNC"], false),
		}

		tenants[tenant.Domain] = tenant
	}

	return tenants, nil
}

func loadTenantFromEnv() TenantConfig {
	return TenantConfig{
		Domain:          getenv("SHEET_HELPER_TENANT", "localhost"),
		CookieSecret:    getenv("SHEET_HELPER_COOKIE_SECRET", "dev-only-change-me"),
		SyncToken:       getenv("SHEET_HELPER_SYNC_TOKEN", ""),
		SheetID:         getenv("SHEET_HELPER_SHEET_ID", ""),
		PublishedURL:    getenv("SHEET_HELPER_PUBLISHED_URL", ""),
		RoutesSheet:     getenv("SHEET_HELPER_ROUTES_SHEET", "routes"),
		VCardsSheet:     getenv("SHEET_HELPER_VCARDS_SHEET", "vcard_entries"),
		TextsSheet:      getenv("SHEET_HELPER_TEXTS_SHEET", "text_entries"),
		DefaultListPref: getenv("SHEET_HELPER_LIST_PREFIX", "list_"),
		StartupSync:     getenvBool("SHEET_HELPER_STARTUP_SYNC", false),
	}
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open env file: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file: %w", err)
	}
	return values, nil
}

func (c Config) SortedTenants() []TenantConfig {
	keys := make([]string, 0, len(c.Tenants))
	for key := range c.Tenants {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]TenantConfig, 0, len(keys))
	for _, key := range keys {
		out = append(out, c.Tenants[key])
	}
	return out
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	return parseBool(os.Getenv(key), fallback)
}

func parseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}

func firstNonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
