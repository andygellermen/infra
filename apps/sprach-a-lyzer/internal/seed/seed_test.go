package seed

import (
	"os"
	"strings"
	"testing"
)

func TestDecodeFoundation(t *testing.T) {
	t.Parallel()

	input := `{"version":"0.1","dimensions":[
{"id":"AGENCY"},{"id":"CONNECTION"},{"id":"APPRECIATION"},
{"id":"CLARITY"},{"id":"VOLITION"},{"id":"OPENNESS"}]}`
	foundation, err := DecodeFoundation(strings.NewReader(input))
	if err != nil {
		t.Fatalf("DecodeFoundation() error: %v", err)
	}
	if len(foundation.Dimensions) != 6 || foundation.Dimensions[4].ID != "VOLITION" {
		t.Fatalf("unexpected foundation: %+v", foundation)
	}
}

func TestCanonicalFoundationKeepsPresentationBundlesSeparate(t *testing.T) {
	t.Parallel()

	file, err := os.Open("../../data/seed/sprach-a-lyzer_foundation_v0.1.json")
	if err != nil {
		t.Fatalf("open canonical foundation: %v", err)
	}
	defer file.Close()
	foundation, err := DecodeFoundation(file)
	if err != nil {
		t.Fatalf("DecodeFoundation() error: %v", err)
	}
	if len(foundation.PresentationBundles) != 2 {
		t.Fatalf("presentation bundles = %d; want 2", len(foundation.PresentationBundles))
	}
}

func TestDecodeFoundationRejectsLegacyDimensionSet(t *testing.T) {
	t.Parallel()

	input := `{"version":"0.1","dimensions":[
{"id":"AGENCY"},{"id":"CONNECTION"},{"id":"APPRECIATION"},
{"id":"CLARITY"},{"id":"FREE_WILL"},{"id":"OPENNESS"}]}`
	_, err := DecodeFoundation(strings.NewReader(input))
	if err == nil {
		t.Fatal("invalid foundation unexpectedly accepted")
	}
}
