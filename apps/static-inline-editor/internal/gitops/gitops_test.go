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

func TestPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	remoteRoot := filepath.Join(t.TempDir(), "remote.git")
	if out, err := exec.Command("git", "init", "--bare", remoteRoot).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %v: %s", err, out)
	}

	repoRoot := t.TempDir()
	if out, err := exec.Command("git", "-C", repoRoot, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, out)
	}
	if out, err := exec.Command("git", "-C", repoRoot, "remote", "add", "origin", remoteRoot).CombinedOutput(); err != nil {
		t.Fatalf("git remote add failed: %v: %s", err, out)
	}

	target := filepath.Join(repoRoot, "index.html")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hash, err := CommitFile(repoRoot, target, "Static Inline Editor", "andy@example.org", "test commit")
	if err != nil {
		t.Fatalf("CommitFile returned error: %v", err)
	}

	pushTarget, err := Push(repoRoot, "origin", "", PushAuth{})
	if err != nil {
		t.Fatalf("Push returned error: %v", err)
	}
	if !strings.HasPrefix(pushTarget, "origin/") {
		t.Fatalf("expected push target to contain origin, got %q", pushTarget)
	}

	branchOut, err := exec.Command("git", "-C", repoRoot, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v: %s", err, branchOut)
	}
	branch := strings.TrimSpace(string(branchOut))
	remoteHashOut, err := exec.Command("git", "--git-dir", remoteRoot, "rev-parse", "--verify", "refs/heads/"+branch).CombinedOutput()
	if err != nil {
		t.Fatalf("remote rev-parse failed: %v: %s", err, remoteHashOut)
	}
	if strings.TrimSpace(string(remoteHashOut)) != hash {
		t.Fatalf("expected remote hash %q, got %q", hash, strings.TrimSpace(string(remoteHashOut)))
	}
}

func TestGitAuthConfigArgs(t *testing.T) {
	args := gitAuthConfigArgs(PushAuth{
		HTTPUsername: "x-token-auth",
		HTTPPassword: "secret-token",
	})
	if got := len(args); got != 2 {
		t.Fatalf("expected 2 auth args, got %d", got)
	}
	if !strings.Contains(args[1], "Authorization: Basic ") {
		t.Fatalf("expected http.extraHeader config, got %q", args[1])
	}
	if strings.Contains(args[1], "secret-token") {
		t.Fatalf("expected token to be encoded, got %q", args[1])
	}
}
