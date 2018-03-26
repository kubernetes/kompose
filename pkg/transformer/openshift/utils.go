package openshift

import (
	"os/exec"
	"strings"
)

// GetImageTag get tag name from image name
// if no tag is specified return 'latest'
func GetImageTag(image string) string {
	// format:      registry_host:registry_port/repo_name/image_name:image_tag
	// example:
	// 1)     myregistryhost:5000/fedora/httpd:version1.0
	// 2)     myregistryhost:5000/fedora/httpd
	// 3)     myregistryhost/fedora/httpd:version1.0
	// 4)     myregistryhost/fedora/httpd
	// 5)     fedora/httpd
	// 6)     httpd
	imageAndTag := image

	i := strings.Split(image, "/")
	if len(i) >= 2 {
		imageAndTag = i[len(i)-1]
	}

	p := strings.Split(imageAndTag, ":")
	if len(p) == 2 {
		return p[1]
	}
	return "latest"

}

// GetAbsBuildContext returns build context relative to project root dir
func GetAbsBuildContext(context string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Dir = context
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	//convert output of command to string
	contextDir := strings.Trim(string(out), "\n")
	return contextDir, nil
}

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
