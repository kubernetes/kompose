package lookup

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
)

// FileConfigLookup is a "bare" structure that implements the project.ConfigLookup interface
type FileConfigLookup struct {
}

// Lookup returns the content and the actual filename of the file that is "built" using the
// specified file and relativeTo string. file and relativeTo are supposed to be file path.
// If file starts with a slash ('/'), it tries to load it, otherwise it will build a
// filename using the folder part of relativeTo joined with file.
func (f *FileConfigLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	if strings.HasPrefix(file, "/") {
		logrus.Debugf("Reading file %s", file)
		bytes, err := ioutil.ReadFile(file)
		return bytes, file, err
	}

	fileName := path.Join(path.Dir(relativeTo), file)
	logrus.Debugf("Reading file %s relative to %s", fileName, relativeTo)
	bytes, err := ioutil.ReadFile(fileName)
	return bytes, fileName, err
}
