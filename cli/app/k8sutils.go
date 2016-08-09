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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Sirupsen/logrus"
)

/**
 * Generate Helm Chart configuration
 */
func generateHelm(filename string, outFiles []string) error {
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
	for _, filename := range outFiles {
		if err = cpFileToChart(manifestDir, filename); err != nil {
			logrus.Warningln(err)
		}
		if err = os.Remove(filename); err != nil {
			logrus.Warningln(err)
		}
	}
	logrus.Infof("chart created in %q\n", "."+string(os.PathSeparator)+dirName+string(os.PathSeparator))
	return nil
}

func cpFileToChart(manifestDir, filename string) error {
	infile, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Warningf("Error reading %s: %s\n", filename, err)
		return err
	}

	return ioutil.WriteFile(manifestDir+string(os.PathSeparator)+filename, infile, 0644)
}
