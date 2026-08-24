package analysis

import (
	"errors"
	"testing"
)

type coreStub struct {
	result Result
	err    error
}

func (c coreStub) Analyze(Request) (Result, error) { return c.result, c.err }

func TestServiceDelegatesToCore(t *testing.T) {
	t.Parallel()

	want := Result{Text: "Test"}
	service := New(coreStub{result: want})
	got, err := service.Analyze(Request{Text: "Test"})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if got.Text != want.Text {
		t.Fatalf("Analyze() result = %+v; want %+v", got, want)
	}
}

func TestServicePreservesCoreError(t *testing.T) {
	t.Parallel()

	want := errors.New("core failed")
	_, err := New(coreStub{err: want}).Analyze(Request{Text: "Test"})
	if !errors.Is(err, want) {
		t.Fatalf("Analyze() error = %v; want %v", err, want)
	}
}
