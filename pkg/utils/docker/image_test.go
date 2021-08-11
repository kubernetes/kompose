/*
Copyright 2016 The Kubernetes Authors All rights reserved

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

package docker

import (
	"reflect"
	"testing"
)

func TestParseImage(t *testing.T) {
	type args struct {
		fullImageName  string
		targetRegistry string
	}
	tests := []struct {
		name    string
		args    args
		want    Image
		wantErr bool
	}{
		{
			"Given empty registry Then default registry expected",
			args{
				"foo/bar",
				"",
			},
			Image{
				"foo/bar:latest",
				"foo/bar",
				"latest",
				"docker.io",
				"docker.io/foo/bar",
				"docker.io/foo/bar:latest",
			},
			false,
		},
		{
			"Given registry from image Then parsed registry expected",
			args{
				"docker.io/foo/bar",
				"",
			},
			Image{
				"foo/bar:latest",
				"foo/bar",
				"latest",
				"docker.io",
				"docker.io/foo/bar",
				"docker.io/foo/bar:latest",
			},
			false,
		},
		{
			"Given target registry Then target registry expected",
			args{
				"foo/bar",
				"localhost:5000",
			},
			Image{
				"foo/bar:latest",
				"foo/bar",
				"latest",
				"localhost:5000",
				"localhost:5000/foo/bar",
				"localhost:5000/foo/bar:latest",
			},
			false,
		},
		{
			"Given registry from image and target registry Then target registry expected",
			args{
				"docker.io/foo/bar",
				"localhost:5000",
			},
			Image{
				"foo/bar:latest",
				"foo/bar",
				"latest",
				"localhost:5000",
				"localhost:5000/foo/bar",
				"localhost:5000/foo/bar:latest",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImage(tt.args.fullImageName, tt.args.targetRegistry)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseImage() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
