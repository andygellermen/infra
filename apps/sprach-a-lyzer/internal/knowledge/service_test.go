package knowledge

import (
	"context"
	"testing"
)

type repositoryStub struct{ snapshot Snapshot }

func (r repositoryStub) Snapshot(context.Context) (Snapshot, error) { return r.snapshot, nil }

func TestServiceUsesRepository(t *testing.T) {
	t.Parallel()

	want := Snapshot{Dimensions: 6, Lexemes: 3}
	got, err := New(repositoryStub{snapshot: want}).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}
	if got != want {
		t.Fatalf("Snapshot() = %+v; want %+v", got, want)
	}
}
