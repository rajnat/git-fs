package gitutils

import (
	"os/exec"
	"strings"
)

// AddAndCommit runs `git add .` and `git commit -m "message"` in the given repo path.
func AddAndCommit(repoPath, message string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// If no changes to commit, might fail - handle gracefully if needed
		return err
	}
	return nil
}

// Push pushes the current branch to the given remote and branch.
func Push(repoPath, remote, branch string) error {
	cmd := exec.Command("git", "push", remote, branch)
	cmd.Dir = repoPath
	return cmd.Run()
}

// GetLastCommitHash returns the latest commit hash in the given repository.
func GetLastCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
