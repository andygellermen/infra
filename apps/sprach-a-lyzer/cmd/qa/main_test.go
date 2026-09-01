package main

import (
	"reflect"
	"testing"
)

func TestSplitListNormalizesConstructIDs(t *testing.T) {
	t.Parallel()
	want := []string{"OPTIONS", "AGENCY"}
	if got := splitList(" options, agency "); !reflect.DeepEqual(got, want) {
		t.Fatalf("splitList() = %v; want %v", got, want)
	}
}
