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
		fmt.Println("create local git dir", err)
		t.Fatal(err)
	}
	return dir
}

func SetGitRemote(t *testing.T, dir string, remote string, remoteUrl string) {
	cmd := NewCommand("git remote add newremote https://git.test.com/somerepo")
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("set git remote", err)
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
		fmt.Println("create git branch", err)
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
