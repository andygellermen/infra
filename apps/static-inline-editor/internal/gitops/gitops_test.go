package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoRoot := t.TempDir()
	if out, err := exec.Command("git", "-C", repoRoot, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, out)
	}

	target := filepath.Join(repoRoot, "index.html")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hash, err := CommitFile(repoRoot, target, "Static Inline Editor", "andy@example.org", "test commit")
	if err != nil {
		t.Fatalf("CommitFile returned error: %v", err)
	}
	if strings.TrimSpace(hash) == "" {
		t.Fatalf("expected commit hash")
	}
}
