package gitutils

import (
	"os/exec"
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
