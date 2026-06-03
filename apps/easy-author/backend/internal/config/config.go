package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Addr          string
	DatabasePath  string
	LibraryDir    string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		Addr:          envOrDefault("EASY_AUTHOR_ADDR", "127.0.0.1:8086"),
		DatabasePath:  envOrDefault("EASY_AUTHOR_DB_PATH", filepath.Join("data", "easy-author.sqlite")),
		LibraryDir:    envOrDefault("EASY_AUTHOR_LIBRARY_DIR", filepath.Join("data", "library")),
		AllowedOrigin: envOrDefault("EASY_AUTHOR_ALLOWED_ORIGIN", "http://127.0.0.1:5173"),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
