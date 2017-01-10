package archive

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateTarball creates a tarball for source and dumps it to target path
func CreateTarball(source, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if baseDir == path {
				return nil
			}
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				if strings.HasSuffix(source, "/") {
					header.Name = strings.TrimPrefix(path, source)
				} else {
					header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
				}
				println("Header name", header.Name)
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}
