package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

var ProjectPath = "$GOPATH/src/github.com/kubernetes/kompose/"
var BinaryLocation = os.ExpandEnv(ProjectPath + "kompose")

func Test_stdin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	kjson := `{"version": "2","services": {"redis": {"image": "redis:3.0","ports": ["6379"]}}}`
	cmdStr := fmt.Sprintf("%s convert --stdout -j -f - <<EOF\n%s\nEOF\n", BinaryLocation, kjson)
	subproc := exec.Command("/bin/sh", "-c", cmdStr)
	output, err := subproc.Output()
	if err != nil {
		fmt.Println("error", err)
	}
	g, err := ioutil.ReadFile("/tmp/output-k8s.json")
	if !bytes.Equal(output, g) {
		t.Errorf("Test Failed")
	}
}
