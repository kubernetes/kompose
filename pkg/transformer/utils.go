package transformer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/skippbox/kompose/pkg/kobject"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandStringBytes generates randomly n-character string
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Create the file to write to if --out is specified
func CreateOutFile(out string) *os.File {
	var f *os.File
	var err error
	if len(out) != 0 {
		f, err = os.Create(out)
		if err != nil {
			logrus.Fatalf("error opening file: %v", err)
		}
	}
	return f
}

// parseVolume parse a given volume, which might be [name:][host:]container[:access_mode]
func ParseVolume(volume string) (name, host, container, mode string, err error) {
	separator := ":"
	volumeStrings := strings.Split(volume, separator)
	if len(volumeStrings) == 0 {
		return
	}
	// Set name if existed
	if !isPath(volumeStrings[0]) {
		name = volumeStrings[0]
		volumeStrings = volumeStrings[1:]
	}
	if len(volumeStrings) == 0 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}
	if volumeStrings[len(volumeStrings)-1] == "rw" || volumeStrings[len(volumeStrings)-1] == "ro" {
		mode = volumeStrings[len(volumeStrings)-1]
		volumeStrings = volumeStrings[:len(volumeStrings)-1]
	}
	container = volumeStrings[len(volumeStrings)-1]
	volumeStrings = volumeStrings[:len(volumeStrings)-1]
	if len(volumeStrings) == 1 {
		host = volumeStrings[0]
	}
	if !isPath(container) || (len(host) > 0 && !isPath(host)) || len(volumeStrings) > 1 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}
	return
}

func isPath(substring string) bool {
	return strings.Contains(substring, "/")
}

// Configure label
func ConfigLabels(name string) map[string]string {
	return map[string]string{"service": name}
}

// Configure annotations
func ConfigAnnotations(service kobject.ServiceConfig) map[string]string {
	annotations := map[string]string{}
	for key, value := range service.Annotations {
		annotations[key] = value
	}

	return annotations
}

// Transform data to json/yaml
func TransformData(obj runtime.Object, GenerateYaml bool) ([]byte, error) {
	//  Convert to versioned object
	objectVersion := obj.GetObjectKind().GroupVersionKind()
	version := unversioned.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
	versionedObj, err := api.Scheme.ConvertToVersion(obj, version)
	if err != nil {
		return nil, err
	}

	// convert data to json / yaml
	data, err := json.MarshalIndent(versionedObj, "", "  ")
	if GenerateYaml == true {
		data, err = yaml.Marshal(versionedObj)
	}
	if err != nil {
		return nil, err
	}
	logrus.Debugf("%s\n", data)
	return data, nil
}

// Either print to stdout or to file/s
func Print(name, trailing string, data []byte, toStdout, generateYaml bool, f *os.File) string {

	file := ""
	if generateYaml {
		file = fmt.Sprintf("%s-%s.yaml", name, trailing)
	} else {
		file = fmt.Sprintf("%s-%s.json", name, trailing)
	}
	if toStdout {
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
		return ""
	} else if f != nil {
		// Write all content to a single file f
		if _, err := f.WriteString(fmt.Sprintf("%s\n", string(data))); err != nil {
			logrus.Fatalf("Failed to write %s to file: %v", trailing, err)
		}
		f.Sync()
	} else {
		// Write content separately to each file
		if err := ioutil.WriteFile(file, []byte(data), 0644); err != nil {
			logrus.Fatalf("Failed to write %s: %v", trailing, err)
		}
		logrus.Printf("file %q created", file)
	}
	return file
}
