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
	output, ok := selectOutput(result, true, false).(analysis.Trace)
	if !ok {
		t.Fatalf("trace output has type %T", selectOutput(result, true, false))
	}
	if len(output.Contributions) != 5 || len(output.Assessability) != 6 {
		t.Fatalf("trace output = %+v", output)
	}
	traceV02, ok := selectOutput(result, false, true).(analysis.TraceV02)
	if !ok || traceV02.ContractVersion != "0.2" || len(traceV02.Propositions) == 0 {
		t.Fatalf("trace v0.2 output has type/value %T/%+v", selectOutput(result, false, true), traceV02)
	}
	if got, ok := selectOutput(result, false, false).(analysis.Result); !ok || got.Text != result.Text {
		t.Fatalf("analysis output has type/value %T/%+v", selectOutput(result, false, false), got)
	}
}
