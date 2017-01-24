package testutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

// NewCommand TODO: comment
func NewCommand(cmd string) *exec.Cmd {
	return exec.Command("sh", "-c", cmd)
}

// CreateLocalDirectory TODO: comment
func CreateLocalDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir(os.TempDir(), "kompose-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// CreateLocalGitDirectory TODO: comment
func CreateLocalGitDirectory(t *testing.T) string {
	dir := CreateLocalDirectory(t)
	cmd := NewCommand(
		`git init &&
		 git config  user.email "you@example.com" &&
		 git config  user.name "Your Name" &&
		 touch README &&
		 git add README &&
		 git commit --no-gpg-sign -m 'testcommit'`)
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		t.Logf("create local git dir: %v", err)
		t.Fatal(err)
	}
	return dir
}

// SetGitRemote TODO: comment
func SetGitRemote(t *testing.T, dir string, remote string, remoteURL string) {
	cmd := NewCommand(fmt.Sprintf("git remote add %s %s", remote, remoteURL))
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		t.Logf("set git remote: %v", err)
		t.Fatal(err)
	}
}

// CreateGitRemoteBranch TODO: comment
func CreateGitRemoteBranch(t *testing.T, dir string, branch string, remote string) {
	cmd := NewCommand(
		fmt.Sprintf(`git checkout -b %s &&
		    git config branch.%s.remote %s &&
		 	git config branch.%s.merge refs/heads/%s`,
			branch, branch, remote, branch, branch))
	cmd.Dir = dir

	_, err := cmd.Output()
	if err != nil {
		t.Logf("create git branch: %v", err)
		t.Fatal(err)
	}
}

// CreateSubdir TODO: comment
func CreateSubdir(t *testing.T, dir string, subdir string) {
	cmd := NewCommand(fmt.Sprintf("mkdir -p %s", subdir))
	cmd.Dir = dir

	_, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
}
