package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"
)

func main() {
	inputPath := flag.String("input", "", "legacy JSON or CSV artifact (required)")
	format := flag.String("format", "auto", "input format: auto, json, or csv")
	flag.Parse()
	if strings.TrimSpace(*inputPath) == "" {
		fail(fmt.Errorf("-input is required"))
	}

	input, err := os.Open(*inputPath)
	if err != nil {
		fail(fmt.Errorf("open input: %w", err))
	}
	defer input.Close()

	resolvedFormat, err := resolveFormat(*format, *inputPath)
	if err != nil {
		fail(err)
	}
	var normalized []byte
	var report dimension.CompatibilityReport
	switch resolvedFormat {
	case "json":
		normalized, report, err = dimension.NormalizeReader(input)
	case "csv":
		normalized, report, err = dimension.NormalizeCSV(input)
	}
	if err != nil {
		fail(err)
	}
	if _, err := os.Stdout.Write(append(normalized, '\n')); err != nil {
		fail(fmt.Errorf("write normalized artifact: %w", err))
	}
	reportEnvelope := map[string]any{
		"format": resolvedFormat, "legacy_mappings": report.LegacyCount(), "mappings": report.Mappings,
	}
	if err := json.NewEncoder(os.Stderr).Encode(reportEnvelope); err != nil {
		fail(fmt.Errorf("write compatibility report: %w", err))
	}
}

func resolveFormat(requested, path string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(requested))
	if normalized == "auto" {
		normalized = strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	}
	if normalized != "json" && normalized != "csv" {
		return "", fmt.Errorf("unsupported format %q; use json or csv", normalized)
	}
	return normalized, nil
}

func fail(err error) {
	_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
