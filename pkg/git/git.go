package git

import (
	"os/exec"
	"strings"
)

// HasGitBinary checks if the 'git' binary is available on the system
func HasGitBinary() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// GetGitCurrentRemoteURL gets current git remote URI for the current git repo
func GetGitCurrentRemoteURL(composeFileDir string) (string, error) {
	cmd := exec.Command("git", "ls-remote", "--get-url")
	cmd.Dir = composeFileDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(string(out), "\n")
	if !strings.HasSuffix(url, ".git") {
		url += ".git"
	}
	return url, nil
}

// GetGitCurrentBranch gets current git branch name for the current git repo
func GetGitCurrentBranch(composeFileDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = composeFileDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}
