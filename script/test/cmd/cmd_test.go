package cmd

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func fileNotExist(path string) bool {
	return findExecutable(path) != nil
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() {
		return nil
	}
	return os.ErrPermission
}

func binary(cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil || fileNotExist(cmd) {
		home := os.Getenv("HOME")
		if home == "" {
			fmt.Println("No $HOME environment variable found")
		}
	}
	return path
}

//Git get the branch
func Git() (branch string, error error) {
	cmdName := "git"
	cmdArgs := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("")
		return "", errors.Wrap(err, "Error getting branch name")
	}
	branch = string(cmdOut)
	return branch, nil
}

//UpdateUri Replaces variables with current branch and uri
func UpdateURI(uri, branch, inputfile, output string) error {
	input, _ := ioutil.ReadFile(inputfile)
	file := string(input)
	r, _ := regexp.Compile("%URI%")
	f := r.ReplaceAllString(file, uri)
	k, _ := regexp.Compile("%REF%")
	g := k.ReplaceAllString(f, branch)
	err := ioutil.WriteFile(output, []byte(g), 0644)
	if err != nil {
		return errors.Wrapf(err, "Error updating the file with %s and %s", uri, branch)
	}
	return nil
}

//Get origin
func GitOrigin() (origin, branch string, error error) {
	cmdName := "git"
	cmdArgs := []string{"remote", "get-url", "origin"}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("")
		return "", "", errors.Wrap(err, "Error getting origin.")
	}
	origin = string(cmdOut)
	origin = strings.TrimSpace(origin)
	if !strings.Contains(origin, ".git") {
		origin = origin + ".git"
	}
	brch, err := Git()
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	branch = strings.TrimSpace(brch)
	fmt.Println("Test regarding build context running kompose from various directories")
	fmt.Println("Buildconfig using " + origin + "::" + branch + " as source.")
	return origin, branch, nil
}

// Kompose command for buildContext
func KomposeCommand(command, provider, outFile, dockerComposeFile string) (error error) {
	file := strings.Split(dockerComposeFile, "/")
	cmdName := "kompose"
	cmdArgs := []string{command, "--provider", provider, "-f", dockerComposeFile, "--stdout", "-j"}
	out, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return errors.Wrapf(err, "Error in %s application", file[1])
	}
	err = ioutil.WriteFile(outFile, out, 0644)
	if err != nil {
		return errors.Wrapf(err, "Error in writing file.")
	}
	return error
}

// Kompose command
func Kompose(command, provider string, dockerComposeFiles []string, result string) (cmd, output string, error error) {
	file, error := ioutil.ReadFile(result)
	output = string(file)
	if error != nil {
		return "", "", errors.Wrap(error, "Error reading file")
	}
	cmd, err := Run(dockerComposeFiles, command, provider, result)
	return cmd, output, err
}

// Run executes different kompose commands
func Run(dockerComposeFiles []string, command, provider, result string) (cmd string, err error) {
	var dcfiles []string
	for _, file := range dockerComposeFiles {
		dcfiles = append(dcfiles, "-f", file)
	}
	cmdName := "kompose"
	if command == "convert" && provider != "" {
		cmdArgs := []string{command, "--provider", provider, "--stdout", "-j"}
		cmdArgs = append(cmdArgs, dcfiles...)
		cmd, err := Exec(cmdName, cmdArgs, dockerComposeFiles)
		if err != nil {
			log.Printf("%+v", err)
		}
		return cmd, err
	} else if command == "convert" && provider == "" {
		cmdArgs := []string{"--bundle", dockerComposeFiles[0], command, "--stdout", "-j"}
		cmd, err := Exec(cmdName, cmdArgs, dockerComposeFiles)
		if err != nil {
			log.Printf("%+v", err)
		}
		return cmd, err
	}
	return cmd, err
}

//Exec runs the kompose command
func Exec(cmdName string, cmdArgs, dockerComposeFiles []string) (cmd string, err error) {
	out, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("\x1b[31;1m===> Test Fail <===\x1b[0m")
		fmt.Println("Test Failed for", dockerComposeFiles)
		return "", errors.Wrapf(err, "Error in %s application", dockerComposeFiles)
	}
	cmd = string(out)
	return cmd, err
}

//GenerateArtifacts generates artifacts in file with -o command
func GenerateArtifactsFile(command, provider, outputFile, dockerComposeFile string) (outfile string, error error) {
	file := strings.Split(dockerComposeFile, "/")
	cmdName := "kompose"
	cmdArgs := []string{command, "--provider", provider, "-f", dockerComposeFile, "-o", outputFile, "-j"}
	_, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("\x1b[31;1m===> Test Failed <===\x1b[0m")
		fmt.Println("Test Failed for", file[1])
		return "", errors.Wrapf(err, "Error in %s appication", file[1])
	}
	return outfile, nil
}

//GenerateArtifacts generates artifacts in directory with -o command
func GenerateArtifactsDir(command, provider, directory, dockerComposeFile string) (output string, error error) {
	output, err := CheckArtifacts(command, provider, dockerComposeFile, directory)
	if err != nil {
		return "", errors.Wrap(err, "Err in application")
	}
	return output, err
}

//GenerateArtifactsDirFile generates artifacts in directory/file with -o command
func GenerateArtifactsDirFile(command, provider, dir, outputFile, dockerComposeFile string) (output string, error error) {
	directory := path.Join(dir, outputFile)
	output, err := CheckArtifacts(command, provider, dockerComposeFile, directory)
	if err != nil {
		return "", errors.Wrap(err, "Err in application")
	}
	return output, err
}

// CheckArtifacts checks the generated artifatcs
func CheckArtifacts(command, provider, dockerComposeFile, directory string) (output string, error error) {
	file := strings.Split(dockerComposeFile, "/")
	cmdName := "kompose"
	cmdArgs := []string{command, "--provider", provider, "-f", dockerComposeFile, "-o", directory, "-j"}
	cmd, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("\x1b[31;1m===> Test Failed <===\x1b[0m")
		fmt.Println("Test Failed for", file[1])
		return "", errors.Wrapf(err, "Err in application %s", file[1])
	}
	output = string(cmd)
	return output, nil
}

// CheckArtifactsInCurrentDirectory checks the generated artifacts in directory
func CheckArtifactsInCurrentDirectory(file string) bool {
	fmt.Println("===> Starting Test ")
	fmt.Println("Testing Generated Artifacts in current directory")
	_, err := os.Stat(file)
	if err != nil {
		logrus.Info("\x1b[32;1mTest Pass\x1b[0m ")
		return true
	}
	logrus.Error("\x1b[31;1mTest Failed\x1b[0m ", "no file found")
	log.Print(err)
	return false
}

// CheckArtifacts checks the generated articfacts
func CheckArtifactsDir(dir string) bool {
	files, err := ioutil.ReadDir("/tmp/sample/")
	if err != nil {
		errors.Wrap(err, "Error reading file")
		return false
	}
	fmt.Println("===> Starting Test ")
	fmt.Println("Testing Generated Artifacts in tmp directory")
	for _, f := range files {
		a := f.Name()
		if _, err := os.Stat(path.Join("/tmp/sample/", a)); err == nil {
			logrus.Info("\x1b[32;1mTest Pass\x1b[0m ", "found file ", a)
		} else {
			logrus.Error("\x1b[31;1mTest Failed\x1b[0m ", "no artifacts found")
			log.Print(err)
			return false
		}
	}
	return true
}

// CheckArtifacts checks the generated artifacts in directory
func CheckArtifactsFile(file string) bool {
	fmt.Println("===> Starting Test ")
	fmt.Println("Testing Generated Artifacts in tmp directory/file")
	_, err := os.Stat(file)
	if err != nil {
		logrus.Error("\x1b[31;1mTest Failed\x1b[0m ")
		log.Print(err)
		return false
	}
	logrus.Info("\x1b[32;1mTest Pass\x1b[0m ")
	return true
}

// ArtifactsFile checks artifacts in current directory and calls CheckArtifacts and GenerateArtifacts
func ArtifactsFile(command, provider, file, dockerComposeFile string) bool {
	output, err := GenerateArtifactsFile(command, provider, file, dockerComposeFile)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if !CheckArtifactsInCurrentDirectory(output) {
		return false
	}
	os.Remove(file)
	return true
}

// ArtifactsDir calls CheckArtifactsDir and GenerateArtifactsDir
func ArtifactsDir(command, provider, directory, dockerComposeFile string) bool {
	os.Mkdir("/tmp/sample", os.FileMode(0777))
	output, err := GenerateArtifactsDir(command, provider, directory, dockerComposeFile)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if !CheckArtifactsDir(output) {
		return false
	}
	// Remove the generated files
	os.Remove("/tmp/sample")
	return true
}

// ArtifactsDirFile calls CheckArtifactsDirFile and GenerateArtifactsDirFile
func ArtifactsDirFile(command, provider, dir, outputFile, dockerComposeFile string) bool {
	os.Mkdir("/tmp/kompose", os.FileMode(0777))
	directory := path.Join(dir, outputFile)
	_, err := GenerateArtifactsDirFile(command, provider, dir, outputFile, dockerComposeFile)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if !CheckArtifactsFile(directory) {
		return false
	}
	os.Remove(directory)
	return true
}

// Compare compares two json files
func Compare(output, cmd, provider string) bool {
	if !reflect.DeepEqual(output, cmd) {
		fmt.Println("\x1b[31;1m===> Test Failed <===\x1b[0m")
		logrus.Error(provider, " \x1b[31;1mTest Failed\x1b[0m")
		return false
	}
	fmt.Println("===> Starting Test ")
	fmt.Println("Testing for", provider)
	logrus.Infoln("\x1b[32;1mTest Pass\x1b[0m")
	fmt.Println("")
	return true

}

// ExpectSuccess Test for K8s
func ExpectSuccess(command, provider string, dockerComposeFile []string, result string) bool {
	cmd, output, err := Kompose(command, provider, dockerComposeFile, result)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if !Compare(output, cmd, provider) {
		return false
	}
	return true
}

// ExpectSuccessAndWarning Test
func ExpectSuccessAndWarning(command, provider string, dockerComposeFiles []string, result string) bool {
	warning := "WARN"
	info := "INFO"
	cmd, output, err := Kompose(command, provider, dockerComposeFiles, result)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if strings.Contains(cmd, warning) || strings.Contains(cmd, info) {
		var dcfiles []string
		for _, file := range dockerComposeFiles {
			dcfiles = append(dcfiles, "-f", file)
		}
		cmdName := "kompose"
		if command == "convert" && provider != "" {
			cmdArgs := []string{command, "--provider", provider, "--stdout", "-j"}
			cmdArgs = append(cmdArgs, dcfiles...)
			str, err := exec.Command(cmdName, cmdArgs...).Output()
			if err != nil {
				errors.Wrapf(err, "Err in cmd")
				return false
			}
			cmd := string(str)
			if !Compare(output, cmd, provider) {
				return false
			}
		}
		if command == "convert" && provider == "" {
			cmdArgs := []string{"--bundle", dockerComposeFiles[0], command, "--stdout", "-j"}
			str, err := exec.Command(cmdName, cmdArgs...).Output()
			if err != nil {
				errors.Wrapf(err, "Err in cmd")
				return false
			}
			cmd := string(str)
			if !Compare(output, cmd, dockerComposeFiles[0]) {
				return false
			}
		}
	}
	return true
}

// ExpectFailure Test
func ExpectFailure(command, provider, dockerComposeFile string) bool {
	cmdName := "kompose"
	cmdArgs := []string{command, provider, "--stdout", "-f", dockerComposeFile, "-j"}
	_, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	file := strings.Split(dockerComposeFile, "/")
	if err != nil {
		fmt.Println("===> Starting Test ")
		fmt.Println("Testing", file[1], "for", provider)
		logrus.Info("\x1b[32;1mTest Pass\x1b[0m")
		return true
	}
	fmt.Println("\x1b[31;1m===> Test Failed <===\x1b[0m")
	logrus.Error(provider, " \x1b[31;1mTest Failed\x1b[0m for", file[1])
	return false
}

// LoadEnv files
func LoadEnv(envFile string) bool {
	err := godotenv.Overload(envFile)
	if err != nil {
		logrus.Error("Unable to load env file ", err)
		return false
	}
	return true
}

func TestCmd(t *testing.T) {
	if binary("kompose") != "" {
		if !ExpectFailure("convert", "kubernetes", "fixtures/etherpad/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectFailure("convert", "openshift", "fixtures/etherpad/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		// Load Env file
		if !LoadEnv("fixtures/etherpad/envs") {
			t.Errorf("Test Failed")
		}
		// Kubernetes Test
		if !ExpectSuccessAndWarning("convert", "kubernetes", []string{"fixtures/etherpad/docker-compose.yml"}, "fixtures/etherpad/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccessAndWarning("convert", "openshift", []string{"fixtures/etherpad/docker-compose.yml"}, "fixtures/etherpad/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/fixtures/gitlab
		// Kubernetes Test
		if !ExpectFailure("convert", "kubernetes", "fixtures/gitlab/docker-compose.yml convert") {
			t.Errorf("Test Failed")
		}
		// Load Env File
		if !LoadEnv("fixtures/gitlab/envs") {
			t.Errorf("Test Failed")
		}
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/gitlab/docker-compose.yml"}, "fixtures/gitlab/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/gitlab/docker-compose.yml"}, "fixtures/gitlab/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/fixtures/nginx-node-redis
		// Kubernetes Test
		if !ExpectSuccessAndWarning("convert", "kubernetes", []string{"fixtures/nginx-node-redis/docker-compose.yml"}, "fixtures/nginx-node-redis/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		//Openshift Test
		origin, branch, err := GitOrigin()
		if err != nil {
			log.Printf("%+v", err)
			return
		}
		error := UpdateURI(origin, branch, "fixtures/nginx-node-redis/output-os-template.json", "/tmp/output.json")
		if error != nil {
			log.Printf("%+v", err)
			return
		}
		er := KomposeCommand("convert", "openshift", "/tmp/output-os.json", "fixtures/nginx-node-redis/docker-compose.yml")
		if er != nil {
			log.Printf("%+v", er)
			return
		}
		if !ExpectSuccessAndWarning("convert", "openshift", []string{"fixtures/nginx-node-redis/docker-compose.yml"}, "/tmp/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/entrypoint-command
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/entrypoint-command/docker-compose.yml"}, "fixtures/entrypoint-command/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/entrypoint-command/docker-compose.yml"}, "fixtures/entrypoint-command/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/mem-limit
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/mem-limit/docker-compose.yml"}, "fixtures/mem-limit/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/mem-limit/docker-compose-mb.yml"}, "fixtures/mem-limit/output-mb-k8s.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/ports-with-proto
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/ports-with-proto/docker-compose.yml"}, "fixtures/ports-with-proto/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/ports-with-proto/docker-compose.yml"}, "fixtures/ports-with-proto/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/volume-mounts/simple-vol-mounts
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml"}, "fixtures/volume-mounts/simple-vol-mounts/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml"}, "fixtures/volume-mounts/simple-vol-mounts/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/volume-mounts/volumes-from
		// Kubernetes Test
		if !ExpectSuccessAndWarning("convert", "kubernetes", []string{"fixtures/volume-mounts/volumes-from/docker-compose.yml"}, "fixtures/volume-mounts/volumes-from/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccessAndWarning("convert", "openshift", []string{"fixtures/volume-mounts/volumes-from/docker-compose.yml"}, "fixtures/volume-mounts/volumes-from/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to docker-compose file in /script/test/cmd/fixtures/envvars-separators
		// Kubernetes Test
		if !ExpectSuccessAndWarning("convert", "kubernetes", []string{"fixtures/envvars-separators/docker-compose.yml"}, "fixtures/envvars-separators/output-k8s.json") {
			t.Errorf("Test Failed")
		}

		// Tests related to unknown arguments with cli commands
		if !ExpectFailure("up", "kubernetes", "fixtures/gitlab/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		if !ExpectFailure("down", "kubernetes", "fixtures/gitlab/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		if !ExpectFailure("convert", "kubernetes", "fixtures/gitlab/docker-compose.yml") {
			t.Errorf("Test Failed")
		}

		// Test related to kompose --bundle convert to ensure that docker bundles are converted properly
		if !ExpectSuccess("convert", "", []string{"fixtures/bundles/dab/docker-compose-bundle.dab"}, "fixtures/bundles/dab/output-k8s.json") {
			t.Errorf("Test Failed")
		}

		// Test related to kompose --bundle convert to ensure that DSB bundles are converted properly
		if !ExpectSuccessAndWarning("convert", "", []string{"fixtures/bundles/dsb/docker-voting-bundle.dsb"}, "fixtures/bundles/dsb/output-k8s.json") {
			t.Errorf("Test Failed")
		}

		//// Test related to multiple-compose files
		//// Kubernetes Test
		//if !ExpectSuccessAndWarning("convert", "kubernetes", "fixtures/multiple-compose-files/output-k8s.json", []string{"fixtures/multiple-compose-files/docker-os.yml", "fixtures/multiple-compose-files/docker-k8s.json"}) {
		//	t.Errorf("Test Failed")
		//}
		//// OpenShift Test
		//if !ExpectSuccessAndWarningMultipleFiles("convert", "openshift", "fixtures/multiple-compose-files/docker-k8s.yml", "fixtures/multiple-compose-files/docker-os.yml", "fixtures/multiple-compose-files/output-openshift.json") {
		//	t.Errorf("Test Failed")
		//}
		// Test related to restart options in docker-compose
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/restart-options/docker-compose-restart-no.yml"}, "fixtures/restart-options/output-k8s-restart-no.json") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/restart-options/docker-compose-restart-onfail.yml"}, "fixtures/restart-options/output-k8s-restart-onfail.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/restart-options/docker-compose-restart-no.yml"}, "fixtures/restart-options/output-os-restart-no.json") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/restart-options/docker-compose-restart-onfail.yml"}, "fixtures/restart-options/output-os-restart-onfail.json") {
			t.Errorf("Test Failed")
		}

		// Test key-only envrionment variable
		if !LoadEnv("fixtures/keyonly-envs/envs") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/keyonly-envs/env.yml"}, "fixtures/keyonly-envs/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// Test related to host:port:container in docker-compose
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/ports-with-ip/docker-compose.yml"}, "fixtures/ports-with-ip/output-k8s.json") {
			t.Errorf("Test Failed")
		}

		// Test related to "stdin_open: true" in docker-compose
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/stdin-true/docker-compose.yml"}, "fixtures/stdin-true/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/stdin-true/docker-compose.yml"}, "fixtures/stdin-true/output-oc.json") {
			t.Errorf("Test Failed")
		}

		// Test related to "tty: true" in docker-compose
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/tty-true/docker-compose.yml"}, "fixtures/tty-true/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/tty-true/docker-compose.yml"}, "fixtures/tty-true/output-oc.json") {
			t.Errorf("Test Failed")
		}

		// Test related to kompose.expose.service label in docker compose file to ensure that services are exposed properly
		// Kubernetes Test
		// when kompose.service.expose="True"
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/expose-service/compose-files/docker-compose-expose-true.yml"}, "fixtures/expose-service/provider-files/kubernetes-expose-true.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.expose.service="<hostname>"
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml"}, "fixtures/expose-service/provider-files/kubernetes-expose-hostname.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.service.expose="True" and multiple ports in docker compose file (first port should be selected)
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml"}, "fixtures/expose-service/provider-files/kubernetes-expose-true-multiple-ports.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.service.expose="<hostname>" and multiple ports in docker compose file (first port should be selected)
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml"}, "fixtures/expose-service/provider-files/kubernetes-expose-hostname-multiple-ports.json") {
			t.Errorf("Test Failed")
		}

		// Test related to kompose.expose.service label in docker compose file to ensure that services are exposed properly
		// OpenShift Test
		// when kompose.service.expose="True"
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/expose-service/compose-files/docker-compose-expose-true.yml"}, "fixtures/expose-service/provider-files/openshift-expose-true.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.expose.service="<hostname>"
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml"}, "fixtures/expose-service/provider-files/openshift-expose-hostname.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.service.expose="True" and multiple ports in docker compose file (first port should be selected)
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml"}, "fixtures/expose-service/provider-files/openshift-expose-true-multiple-ports.json") {
			t.Errorf("Test Failed")
		}
		// when kompose.service.expose="<hostname>" and multiple ports in docker compose file (first port should be selected)
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml"}, "fixtures/expose-service/provider-files/openshift-expose-hostname-multiple-ports.json") {
			t.Errorf("Test Failed")
		}

		// Test the change in the service name
		// Kubernetes Test
		if !ExpectSuccessAndWarning("convert", "kubernetes", []string{"fixtures/service-name-change/docker-compose.yml"}, "fixtures/service-name-change/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// OpenShift Test
		if !ExpectSuccessAndWarning("convert", "openshift", []string{"fixtures/service-name-change/docker-compose.yml"}, "fixtures/service-name-change/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Test the output file behavior of kompose convert
		if !ArtifactsFile("convert", "kubernetes", "output", "fixtures/redis-example/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		if !ArtifactsDir("convert", "kubernetes", "/tmp/sample", "fixtures/redis-example/docker-compose.yml") {
			t.Errorf("Test Failed")
		}
		if !ArtifactsDirFile("convert", "kubernetes", "/tmp/kompose", "outfile", "fixtures/redis-example/docker-compose.yml") {
			t.Errorf("Test Failed")
		}

		// Test related to support docker-compose.yaml beside docker-compose.yml
		// Kubernetes Test
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/yaml-and-yml/docker-compose.yaml"}, "fixtures/yaml-and-yml/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "kubernetes", []string{"fixtures/yaml-and-yml/yml/docker-compose.yml"}, "fixtures/yaml-and-yml/yml/output-k8s.json") {
			t.Errorf("Test Failed")
		}
		// Openshift Test
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/yaml-and-yml/docker-compose.yaml"}, "fixtures/yaml-and-yml/output-os.json") {
			t.Errorf("Test Failed")
		}
		if !ExpectSuccess("convert", "openshift", []string{"fixtures/yaml-and-yml/yml/docker-compose.yml"}, "fixtures/yaml-and-yml/yml/output-os.json") {
			t.Errorf("Test Failed")
		}

		// Test regarding build context (running kompose from various directories)
		fmt.Println("BuildConfig using " + origin + "::" + branch + " as source.")
		if !LoadEnv("fixtures/buildargs/envs") {
			t.Errorf("Test Failed")
		}
		eror := UpdateURI(origin, branch, "fixtures/buildargs/output-os-template.json", "/tmp/output-buildarg-os.json")
		if eror != nil {
			log.Printf("%+v", eror)
			return
		}
		erorr := KomposeCommand("convert", "openshift", "/tmp/out-build.json", "fixtures/buildargs/docker-compose.yml")
		if erorr != nil {
			log.Printf("%+v", erorr)
			return
		}
		if !ExpectSuccessAndWarning("convert", "openshift", []string{"fixtures/buildargs/docker-compose.yml"}, "/tmp/out-build.json") {
			t.Errorf("Test Failed")
		}

	}
}
