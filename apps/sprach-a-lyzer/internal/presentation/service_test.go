package presentation

import (
	"bytes"
	"encoding/json"
	"testing"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
)

func TestCorporateBundleNeverReturnsCanonicalKey(t *testing.T) {
	t.Parallel()

	bundle := Bundle{
		Profile:   "CORPORATE",
		Entries:   map[string]string{"RESONANCE": "Wahrnehmungswirkung"},
		Fallbacks: map[string]string{"UNKNOWN_CONCEPT": "Sprachmuster"},
	}
	if got := bundle.Resolve("RESONANCE"); got != "Wahrnehmungswirkung" {
		t.Fatalf("Resolve(RESONANCE) = %q", got)
	}
	if got := bundle.Resolve("PRIVATE_CANONICAL_KEY"); got != "Sprachmuster" {
		t.Fatalf("unknown key leaked as %q", got)
	}
}

func TestFoundationPresentationBundlesStayProfileIsolated(t *testing.T) {
	t.Parallel()
	var foundation struct {
		Bundles []struct {
			Profile   string            `json:"profile"`
			Locale    string            `json:"locale"`
			Version   string            `json:"version"`
			Entries   map[string]string `json:"entries"`
			Fallbacks map[string]string `json:"fallbacks"`
		} `json:"presentation_bundles"`
	}
	decoder := json.NewDecoder(bytes.NewReader(assets.FoundationV04))
	if err := decoder.Decode(&foundation); err != nil {
		t.Fatal(err)
	}
	bundles := map[string]Bundle{}
	for _, value := range foundation.Bundles {
		bundles[value.Profile] = Bundle{Profile: value.Profile, Locale: value.Locale, Version: value.Version, Entries: value.Entries, Fallbacks: value.Fallbacks}
	}
	private, privateOK := bundles["PRIVATE"]
	corporate, corporateOK := bundles["CORPORATE"]
	if !privateOK || !corporateOK || private.Version != "0.2" || corporate.Version != "0.2" {
		t.Fatalf("presentation bundles = %+v", bundles)
	}
	for _, key := range []string{"METRIC_WING_SCORE", "RESONANCE", "VOLITION"} {
		if private.Resolve(key) == corporate.Resolve(key) {
			t.Errorf("profile key %s is not isolated: %q", key, private.Resolve(key))
		}
	}
	for _, key := range []string{"PRIVATE_CANONICAL_KEY", "RESONANCE_PRIVATE_ONLY"} {
		if corporate.Resolve(key) != "Sprachmuster" {
			t.Errorf("corporate unknown key %s leaked as %q", key, corporate.Resolve(key))
		}
	}
}
