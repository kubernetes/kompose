package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Godep struct {
	Deps []Dep
}

type Dep struct {
	ImportPath string
	Comment    string
	Rev        string
}

func main() {
	comment := false
	args := os.Args[1:]
	if len(args) == 3 {
		if args[2] == "comment" {
			comment = true
		} else {
			fmt.Fprintf(os.Stderr, "The third argument must be 'comment' or not specified.\n")
			os.Exit(1)
		}
		args = args[:2]
	}
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Expects two arguments, a path to the Godep.json file and a package to get the commit for (and optionally, 'comment' as the third option)\n")
		os.Exit(1)
	}

	path := args[0]
	pkg := args[1]

	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read %s: %v\n", path, err)
		os.Exit(1)
	}
	godeps := &Godep{}
	if err := json.Unmarshal(data, godeps); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read %s: %v\n", path, err)
		os.Exit(1)
	}

	for _, dep := range godeps.Deps {
		if dep.ImportPath != pkg {
			continue
		}
		if len(dep.Rev) > 7 {
			dep.Rev = dep.Rev[0:7]
		}
		if comment && len(dep.Comment) > 0 {
			dep.Rev = dep.Comment
		}
		fmt.Fprintf(os.Stdout, dep.Rev)
		return
	}

	fmt.Fprintf(os.Stderr, "Could not find %s in %s\n", pkg, path)
	os.Exit(1)
}
