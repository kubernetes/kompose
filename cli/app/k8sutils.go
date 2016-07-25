/*
Copyright 2016 Skippbox, Ltd All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Sirupsen/logrus"
)

/* Ancilliary helper functions to interface with the commands interface */

/**
 * Generate Helm Chart configuration
 */
func generateHelm(filename string, svcnames []string, generateYaml, createD, createDS, createRS, createRC bool, outFile string) error {
	type ChartDetails struct {
		Name string
	}

	extension := filepath.Ext(filename)
	dirName := filename[0 : len(filename)-len(extension)]
	details := ChartDetails{dirName}
	manifestDir := dirName + string(os.PathSeparator) + "templates"
	dir, err := os.Open(dirName)

	/* Setup the initial directories/files */
	if err == nil {
		_ = dir.Close()
	}

	if err != nil {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			return err
		}

		err = os.Mkdir(manifestDir, 0755)
		if err != nil {
			return err
		}

		/* Create the readme file */
		readme := "This chart was created by Kompose\n"
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"README.md", []byte(readme), 0644)
		if err != nil {
			return err
		}

		/* Create the Chart.yaml file */
		chart := `name: {{.Name}}
description: A generated Helm Chart for {{.Name}} from Skippbox Kompose
version: 0.0.1
keywords:
  - {{.Name}}
sources:
home:
`

		t, err := template.New("ChartTmpl").Parse(chart)
		if err != nil {
			logrus.Fatalf("Failed to generate Chart.yaml template: %s\n", err)
		}
		var chartData bytes.Buffer
		_ = t.Execute(&chartData, details)

		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"Chart.yaml", chartData.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	/* Copy all related json/yaml files into the newly created manifests directory */
	if len(outFile) > 0 {
		if err = cpFileToChart(manifestDir, outFile); err != nil {
			return err
		}
	} else {
		for _, svcname := range svcnames {
			extension := ".json"
			if generateYaml {
				extension = ".yaml"
			}
			if createD {
				if err = cpToChart(manifestDir, svcname, "deployment", extension); err != nil {
					return err
				}
			}
			if createDS {
				if err = cpToChart(manifestDir, svcname, "daemonset", extension); err != nil {
					return err
				}
			}
			if createRC {
				if err = cpToChart(manifestDir, svcname, "rc", extension); err != nil {
					return err
				}
			}
			if createRS {
				if err = cpToChart(manifestDir, svcname, "replicaset", extension); err != nil {
					return err
				}
			}

			/* The svc file is optional */
			infile, err := ioutil.ReadFile(svcname + "-svc" + extension)
			if err != nil {
				continue
			}
			err = ioutil.WriteFile(manifestDir+string(os.PathSeparator)+svcname+"-svc"+extension, infile, 0644)
			if err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(os.Stdout, "chart created in %q\n", "."+string(os.PathSeparator)+dirName+string(os.PathSeparator))
	return nil
}

func cpToChart(manifestDir, svcname, trailing, extension string) error {
	return cpFileToChart(manifestDir, svcname+"-"+trailing+extension)
}

func cpFileToChart(manifestDir, filename string) error {
	infile, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Infof("Error reading %s: %s\n", filename, err)
		return err
	}

	return ioutil.WriteFile(manifestDir+string(os.PathSeparator)+filename, infile, 0644)
}
