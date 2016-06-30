package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/logger"
)

var projectRegexp = regexp.MustCompile("[^a-zA-Z0-9_.-]")

// Context holds context meta information about a libcompose project, like
// the project name, the compose file, etc.
type Context struct {
	Timeout             uint
	Log                 bool
	Volume              bool
	ForceRecreate       bool
	NoRecreate          bool
	NoCache             bool
	Signal              int
	ComposeFile         string
	ComposeBytes        []byte
	ProjectName         string
	isOpen              bool
	ServiceFactory      ServiceFactory
	EnvironmentLookup   EnvironmentLookup
	ConfigLookup        ConfigLookup
	LoggerFactory       logger.Factory
	IgnoreMissingConfig bool
	Project             *Project
}

func (c *Context) readComposeFile() error {
	if c.ComposeBytes != nil {
		return nil
	}

	logrus.Debugf("Opening compose file: %s", c.ComposeFile)

	if c.ComposeFile == "-" {
		composeBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			logrus.Errorf("Failed to read compose file from stdin: %v", err)
			return err
		}
		c.ComposeBytes = composeBytes
	} else if c.ComposeFile != "" {
		if composeBytes, err := ioutil.ReadFile(c.ComposeFile); os.IsNotExist(err) {
			if c.IgnoreMissingConfig {
				return nil
			}
			logrus.Errorf("Failed to find %s", c.ComposeFile)
			return err
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.ComposeFile)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	}

	return nil
}

func (c *Context) determineProject() error {
	name, err := c.lookupProjectName()
	if err != nil {
		return err
	}

	c.ProjectName = projectRegexp.ReplaceAllString(strings.ToLower(name), "-")

	if c.ProjectName == "" {
		return fmt.Errorf("Falied to determine project name")
	}

	if strings.ContainsAny(c.ProjectName[0:1], "_.-") {
		c.ProjectName = "x" + c.ProjectName
	}

	return nil
}

func (c *Context) lookupProjectName() (string, error) {
	if c.ProjectName != "" {
		return c.ProjectName, nil
	}

	if envProject := os.Getenv("COMPOSE_PROJECT_NAME"); envProject != "" {
		return envProject, nil
	}

	f, err := filepath.Abs(c.ComposeFile)
	if err != nil {
		logrus.Errorf("Failed to get absolute directory for: %s", c.ComposeFile)
		return "", err
	}

	f = toUnixPath(f)

	parent := path.Base(path.Dir(f))
	if parent != "" && parent != "." {
		return parent, nil
	} else if wd, err := os.Getwd(); err != nil {
		return "", err
	} else {
		return path.Base(toUnixPath(wd)), nil
	}
}

func toUnixPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	if err := c.readComposeFile(); err != nil {
		return err
	}

	if err := c.determineProject(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}
