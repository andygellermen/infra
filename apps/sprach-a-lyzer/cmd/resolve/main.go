package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

func main() {
	text := flag.String("text", "", "German text to resolve")
	contextValue := flag.String("context", "UNSPECIFIED", "analysis context")
	flag.Parse()

	result, err := analysis.NewDefault().Resolve(analysis.Request{
		Text: *text, Context: analysis.Context(*contextValue), InputMode: analysis.InputModeText,
	})
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
