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

package loader

import (
	"fmt"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader/bundle"
	"github.com/kubernetes/kompose/pkg/loader/compose"
)

// Loader interface defines loader that loads files and converts it to kobject representation
type Loader interface {
	LoadFile(files []string) (kobject.KomposeObject, error)
	///Name() string
}

// GetLoader returns loader for given format
func GetLoader(format string) (Loader, error) {
	var l Loader

	switch format {
	case "bundle":
		l = new(bundle.Bundle)
	case "compose":
		l = new(compose.Compose)
	default:
		return nil, fmt.Errorf("Input file format %s is not supported", format)
	}

	return l, nil

}
