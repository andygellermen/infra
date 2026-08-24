package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/engine"
)

func main() {
	text := flag.String("text", "", "German text to analyze")
	context := flag.String("context", "UNSPECIFIED", "analysis context, for example SELF_TALK or SAFETY")
	flag.Parse()

	result, err := engine.New().Analyze(domain.AnalysisRequest{Text: *text, Context: *context, InputMode: "TEXT"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
