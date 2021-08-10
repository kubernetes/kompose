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
