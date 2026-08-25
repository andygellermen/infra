package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

func main() {
	text := flag.String("text", "", "German text to analyze")
	context := flag.String("context", "UNSPECIFIED", "analysis context, for example SELF_TALK or SAFETY")
	traceOnly := flag.Bool("trace", false, "emit the standalone Contribution Trace instead of the analysis result")
	flag.Parse()

	result, err := analysis.NewDefault().Analyze(analysis.Request{
		Text: *text, Context: analysis.Context(*context), InputMode: analysis.InputModeText,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	output := selectOutput(result, *traceOnly)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func selectOutput(result analysis.Result, traceOnly bool) any {
	if traceOnly {
		return result.Trace()
	}
	return result
}
