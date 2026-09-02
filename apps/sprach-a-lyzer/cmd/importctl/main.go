// Command importctl runs a local managed-import validation or one-shot commit.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/managedimport"
)

func main() {
	file := flag.String("file", "", "JSON, CSV or XLSX source file")
	format := flag.String("format", "", "source format; inferred from extension when empty")
	collection := flag.String("collection", "", "JSON collection key")
	entity := flag.String("entity", "", "target entity, for example LEXEME")
	operation := flag.String("operation", "VALIDATE_ONLY", "IMPORT, UPDATE or VALIDATE_ONLY")
	commit := flag.Bool("commit", false, "commit a READY plan in the one-shot local store")
	flag.Parse()
	data, err := os.ReadFile(*file)
	if err != nil {
		fail(err)
	}
	sourceType := strings.ToUpper(*format)
	if sourceType == "" {
		sourceType = inferFormat(*file)
	}
	request := managedimport.PrepareRequest{BatchKey: "CLI-" + filepath.Base(*file), OperationType: strings.ToUpper(*operation), SourceType: sourceType, SourceName: filepath.Base(*file), SourceCollection: *collection, TargetEntity: strings.ToUpper(*entity), ConflictPolicy: "USE_IMPORT", AllowInsert: true, AllowUpdate: true, ActorID: "local-cli", ActorRole: "ADMIN"}
	if sourceType == "XLSX" {
		request.SourceBase64 = base64.StdEncoding.EncodeToString(data)
	} else {
		request.SourceContent = string(data)
	}
	service := managedimport.New(managedimport.NewMemoryRepository())
	ctx := context.Background()
	plan, err := service.Prepare(ctx, request)
	if err != nil {
		fail(err)
	}
	var output any = plan
	if *commit {
		if plan.Status != "READY" {
			fail(fmt.Errorf("plan status %s is not READY", plan.Status))
		}
		output, err = service.Commit(ctx, managedimport.CommitRequest{BatchID: plan.BatchID, ActorID: "local-cli", ActorRole: "ADMIN"})
		if err != nil {
			fail(err)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fail(err)
	}
}

func inferFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "JSON"
	case ".csv":
		return "CSV"
	case ".xlsx":
		return "XLSX"
	default:
		return ""
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(2) }
