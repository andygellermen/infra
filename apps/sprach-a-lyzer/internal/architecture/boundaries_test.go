package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/"

func TestFeatureModulesDoNotImportEachOther(t *testing.T) {
	t.Parallel()

	root := appRoot(t)
	modules := []string{"analysis", "knowledge", "rules", "presentation"}
	for _, module := range modules {
		module := module
		t.Run(module, func(t *testing.T) {
			imports := importsBelow(t, filepath.Join(root, "internal", module))
			for _, other := range modules {
				if other == module {
					continue
				}
				for _, imported := range imports {
					if imported == modulePath+other || strings.HasPrefix(imported, modulePath+other+"/") {
						t.Errorf("module %s imports sibling module %s", module, imported)
					}
				}
			}
		})
	}
}

func TestAnalysisInternalsAreOnlyUsedThroughFacade(t *testing.T) {
	t.Parallel()

	root := appRoot(t)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relative, filepath.Join("internal", "analysis")) ||
			strings.HasPrefix(relative, filepath.Join("internal", "engine")) ||
			strings.HasPrefix(relative, filepath.Join("internal", "domain")) ||
			strings.HasPrefix(relative, filepath.Join("internal", "ontology")) ||
			strings.HasPrefix(relative, filepath.Join("internal", "resolver")) {
			return nil
		}
		for _, imported := range importsInFile(t, path) {
			if imported == modulePath+"engine" || imported == modulePath+"domain" || imported == modulePath+"resolver" {
				t.Errorf("%s bypasses the analysis facade by importing %s", relative, imported)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk source tree: %v", err)
	}
}

func importsBelow(t *testing.T, root string) []string {
	t.Helper()
	var result []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return walkErr
		}
		result = append(result, importsInFile(t, path)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return result
}

func importsInFile(t *testing.T, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	result := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", path, err)
		}
		result = append(result, value)
	}
	return result
}

func appRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
