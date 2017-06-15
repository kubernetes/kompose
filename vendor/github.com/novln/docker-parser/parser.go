//
// Copyright (C) 2015-2017 Thomas LE ROUX <thomas@leroux.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package dockerparser

import (
	"strings"

	"github.com/novln/docker-parser/docker"
)

// Reference is an opaque object that include identifier such as a name, tag, repository, registry, etc...
type Reference struct {
	named docker.Named
	tag   string
}

// Name returns the image's name. (ie: debian[:8.2])
func (r Reference) Name() string {
	return r.named.RemoteName() + r.tag
}

// ShortName returns the image's name (ie: debian)
func (r Reference) ShortName() string {
	return r.named.RemoteName()
}

// Tag returns the image's tag (or digest).
func (r Reference) Tag() string {
	if len(r.tag) > 1 {
		return r.tag[1:]
	}
	return ""
}

// Registry returns the image's registry. (ie: host[:port])
func (r Reference) Registry() string {
	return r.named.Hostname()
}

// Repository returns the image's repository. (ie: registry/name)
func (r Reference) Repository() string {
	return r.named.FullName()
}

// Remote returns the image's remote identifier. (ie: registry/name[:tag])
func (r Reference) Remote() string {
	return r.named.FullName() + r.tag
}

func clean(url string) string {

	s := url

	if strings.HasPrefix(url, "http://") {
		s = strings.Replace(url, "http://", "", 1)
	} else if strings.HasPrefix(url, "https://") {
		s = strings.Replace(url, "https://", "", 1)
	}

	return s
}

// Parse returns a Reference from analyzing the given remote identifier.
func Parse(remote string) (*Reference, error) {

	n, err := docker.ParseNamed(clean(remote))

	if err != nil {
		return nil, err
	}

	n = docker.WithDefaultTag(n)

	var t string
	switch x := n.(type) {
	case docker.Canonical:
		t = "@" + x.Digest().String()
	case docker.NamedTagged:
		t = ":" + x.Tag()
	}

	return &Reference{named: n, tag: t}, nil
}
