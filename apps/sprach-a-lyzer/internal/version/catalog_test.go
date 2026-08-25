package version

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCoreMatchesReleaseManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.1.0.json")
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	var manifest struct {
		ReleaseVersion string `json:"release_version"`
		GitTag         string `json:"git_tag"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode release manifest: %v", err)
	}
	if manifest.ReleaseVersion != Core || manifest.GitTag != "v"+Core {
		t.Fatalf("release manifest = %s/%s; core = %s", manifest.ReleaseVersion, manifest.GitTag, Core)
	}
}
