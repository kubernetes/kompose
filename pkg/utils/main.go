package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubernetes-incubator/kompose/pkg/utils/docker"
)

func main() {
	dir := os.Args[1]
	image := os.Args[2]
	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	fmt.Println("Abs dir", absDir)
	docker.Build(absDir, image)
	docker.Push(image)
}
