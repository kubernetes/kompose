/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package compose

import (
	"github.com/compose-spec/compose-go/cli"
	"github.com/kubernetes/kompose/pkg/kobject"
	log "github.com/sirupsen/logrus"
)

//StdinData is data bytes read from stdin
var StdinData []byte

// Compose is docker compose file loader, implements Loader interface
type Compose struct {
}

// LoadFile loads a compose file into KomposeObject
func (c *Compose) LoadFile(files []string) (kobject.KomposeObject, error) {
	options, err := cli.NewProjectOptions(files,
		cli.WithDotEnv,
		cli.WithOsEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	project, err := cli.ProjectFromOptions(options)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	noSupKeys := checkUnsupportedKeyForV3(project)
	for _, keyName := range noSupKeys {
		log.Warningf("Unsupported %s key - ignoring", keyName)
	}

	// Finally, we convert the object from docker/cli's ServiceConfig to our appropriate one
	return dockerComposeToKomposeMapping(project)
}
