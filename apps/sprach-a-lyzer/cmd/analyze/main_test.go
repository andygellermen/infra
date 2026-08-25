package main

import (
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

func TestSelectOutputCanEmitStandaloneTrace(t *testing.T) {
	t.Parallel()

	result, err := analysis.NewDefault().Analyze(analysis.Request{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: analysis.ContextSelfTalk,
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	output, ok := selectOutput(result, true).(analysis.Trace)
	if !ok {
		t.Fatalf("trace output has type %T", selectOutput(result, true))
	}
	if len(output.Contributions) != 5 || len(output.Assessability) != 6 {
		t.Fatalf("trace output = %+v", output)
	}
	if got, ok := selectOutput(result, false).(analysis.Result); !ok || got.Text != result.Text {
		t.Fatalf("analysis output has type/value %T/%+v", selectOutput(result, false), got)
	}
}
