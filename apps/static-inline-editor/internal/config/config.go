package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andygellermann/infra/apps/static-inline-editor/internal/model"
)

type Config struct {
	Addr          string
	DataDir       string
	TenantDir     string
	SessionTTL    string
	SecureCookies bool
	Tenants       map[string]model.Tenant
}

func Load() (Config, error) {
	cfg := Config{
		Addr:          getenv("STATIC_EDITOR_ADDR", ":8090"),
		DataDir:       getenv("STATIC_EDITOR_DATA_DIR", "./data"),
		TenantDir:     getenv("STATIC_EDITOR_TENANT_DIR", "./tenants"),
		SessionTTL:    getenv("STATIC_EDITOR_SESSION_TTL", "12h"),
		SecureCookies: getenvBool("STATIC_EDITOR_SECURE_COOKIES", true),
		Tenants:       map[string]model.Tenant{},
	}

	tenants, err := loadTenantsFromDir(cfg.TenantDir)
	if err != nil {
		return Config{}, err
	}
	cfg.Tenants = tenants
	return cfg, nil
}

func loadTenantsFromDir(dir string) (map[string]model.Tenant, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]model.Tenant{}, nil
		}
		return nil, fmt.Errorf("read tenant dir %s: %w", dir, err)
	}

	tenants := make(map[string]model.Tenant)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		values, err := parseEnvFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse tenant env %s: %w", path, err)
		}

		domain := strings.TrimSpace(values["STATIC_EDITOR_TENANT"])
		if domain == "" {
			continue
		}

		tenants[domain] = model.Tenant{
			Domain:            domain,
			LoginDomain:       firstNonEmpty(values["STATIC_EDITOR_LOGIN_DOMAIN"], "bearbeitung."+domain),
			Aliases:           parseCSV(values["STATIC_EDITOR_ALIASES"]),
			StaticRoot:        strings.TrimSpace(values["STATIC_EDITOR_STATIC_ROOT"]),
			BackupRoot:        strings.TrimSpace(values["STATIC_EDITOR_BACKUP_ROOT"]),
			RepoRoot:          strings.TrimSpace(values["STATIC_EDITOR_REPO_ROOT"]),
			Username:          firstNonEmpty(values["STATIC_EDITOR_USERNAME"], "admin"),
			PasswordHash:      strings.TrimSpace(values["STATIC_EDITOR_PASSWORD_HASH"]),
			CookieSecret:      strings.TrimSpace(values["STATIC_EDITOR_COOKIE_SECRET"]),
			MainSelector:      firstNonEmpty(values["STATIC_EDITOR_MAIN_SELECTOR"], "main"),
			AllowedBlockTags:  parseCSV(firstNonEmpty(values["STATIC_EDITOR_ALLOWED_BLOCK_TAGS"], "h1,h2,h3,h4,h5,p,ul,ol,li")),
			AllowedInlineTags: parseCSV(firstNonEmpty(values["STATIC_EDITOR_ALLOWED_INLINE_TAGS"], "strong,em,a,br")),
			StartPath:         firstNonEmpty(values["STATIC_EDITOR_START_PATH"], "/index.html"),
		}
	}

	return tenants, nil
}

func (c Config) SortedTenantDomains() []string {
	keys := make([]string, 0, len(c.Tenants))
	for key := range c.Tenants {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
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
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
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
