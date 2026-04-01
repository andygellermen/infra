package gitops

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func CommitFile(repoRoot, fullPath, authorName, authorEmail, message string) (string, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return "", nil
	}

	relativePath, err := filepath.Rel(repoRoot, fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve git relative path: %w", err)
	}
	if strings.HasPrefix(relativePath, "..") {
		return "", fmt.Errorf("target file is outside repo root")
	}

	if err := runGit(repoRoot, "add", "--", relativePath); err != nil {
		return "", err
	}

	env := []string{
		"GIT_AUTHOR_NAME=" + authorName,
		"GIT_AUTHOR_EMAIL=" + authorEmail,
		"GIT_COMMITTER_NAME=" + authorName,
		"GIT_COMMITTER_EMAIL=" + authorEmail,
	}
	if err := runGitWithEnv(repoRoot, env, "commit", "--message", message); err != nil {
		return "", err
	}

	hash, err := outputGit(repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(hash), nil
}

func Push(repoRoot, remoteName, branch string) (string, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return "", nil
	}
	if strings.TrimSpace(remoteName) == "" {
		remoteName = "origin"
	}
	if strings.TrimSpace(branch) == "" {
		currentBranch, err := outputGit(repoRoot, "branch", "--show-current")
		if err != nil {
			return "", err
		}
		branch = strings.TrimSpace(currentBranch)
	}
	if branch == "" {
		return "", fmt.Errorf("could not detect git branch for push")
	}
	if err := runGit(repoRoot, "push", remoteName, branch); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", remoteName, branch), nil
}

func runGit(repoRoot string, args ...string) error {
	return runGitWithEnv(repoRoot, nil, args...)
}

func runGitWithEnv(repoRoot string, env []string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	if len(env) > 0 {
		cmd.Env = append(cmd.Environ(), env...)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func outputGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
