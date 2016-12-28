package testutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func NewCommand(cmd string) *exec.Cmd {
	return exec.Command("sh", "-c", cmd)
}

func CreateLocalDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir(os.TempDir(), "kompose-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func CreateLocalGitDirectory(t *testing.T) string {
	dir := CreateLocalDirectory(t)
	cmd := NewCommand(
		`git init && touch README &&
		git add README &&
		git commit -m 'testcommit'`)
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		t.Logf("create local git dir: %v", err)
		t.Fatal(err)
	}
	return dir
}

func SetGitRemote(t *testing.T, dir string, remote string, remoteUrl string) {
	cmd := NewCommand(fmt.Sprintf("git remote add %s %s", remote, remoteUrl))
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		t.Logf("set git remote: %v", err)
		t.Fatal(err)
	}
}

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

func CreateSubdir(t *testing.T, dir string, subdir string) {
	cmd := NewCommand(fmt.Sprintf("mkdir -p %s", subdir))
	cmd.Dir = dir

	_, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
}
