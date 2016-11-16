# Development Guide

## Building Kompose
### Build with `go`

- you need `go v1.5` or later
- You need to set export `GO15VENDOREXPERIMENT=1` environment variable
- If your working copy is not in your `GOPATH`, you need to set it accordingly.

```console
$ go build -tags experimental -o kompose ./cli/main
```

You need `-tags experimental` because the current bundlefile package of docker/libcompose is still experimental.

### Building binaries with make

- You need `make`

#### Build multi-platforms:
```console
$ make binary-cross
```

#### Build current platform only:
```console
$ make binary
```

## Workflow
### Fork the main repository

1. Go to https://github.com/kubernetes-incubator/kompose
2. Click the "Fork" button (at the top right)

### Clone your fork

The commands below require that you have $GOPATH. We highly recommended you put Kompose' code into your $GOPATH.

```console
mkdir -p $GOPATH/src/github.com/kubernetes-incubator
cd $GOPATH/src/github.com/kubernetes-incubator
git clone https://github.com/$YOUR_GITHUB_USERNAME/kompose.git
cd kompose
git remote add upstream 'https://github.com/kubernetes-incubator/kompose'
```

### Create a branch and make changes

```console
git checkout -b myfeature
# Make your code changes
```

### Keeping your development fork in sync

```console
git fetch upstream
git rebase upstream/master
```

Note: If you have write access to the main repository at github.com/kubernetes-incubator/kompose, you should modify your git configuration so that you can't accidentally push to upstream:

```console
git remote set-url --push upstream no_push
```

### Committing changes to your fork

```console
git commit
git push -f origin myfeature
```

### Creating a pull request

1. Visit https://github.com/$YOUR_GITHUB_USERNAME/kompose.git
2. Click the "Compare and pull request" button next to your "myfeature" branch.
3. Check out the pull request process for more details

## Godep and dependency management

Kompose uses `godep` to manage dependencies. It is not strictly required for building Kompose but it is required when managing dependencies under the vendor/ tree, and is required by a number of the build and test scripts. Please make sure that godep is installed and in your `$PATH`, and that `godep version` says it is at least v63.

### Installing godep

There are many ways to build and host Go binaries. Here is an easy way to get utilities like `godep` installed:

1) Ensure that mercurial is installed on your system. (some of godep's dependencies use the mercurial source control system). Use `apt-get install mercurial` or `yum install mercurial` on Linux, or brew.sh on OS X, or download directly from mercurial.

2) Create a new GOPATH for your tools and install godep:

```console
export GOPATH=$HOME/go-tools
mkdir -p $GOPATH
go get -u github.com/tools/godep
```

3) Add this $GOPATH/bin to your path. Typically you'd add this to your ~/.profile:

```console
export GOPATH=$HOME/go-tools
export PATH=$PATH:$GOPATH/bin
```

To check your version of godep:

```console
godep version
godep v74 (linux/amd64/go1.6.2)
```

Note: At this time, godep version >= v64 is known to work in the Kompose project.

If it is not a valid version try, make sure you have updated the godep repo with `go get -u github.com/tools/godep`.

### Using godep

1. Populate dependencies for your Kompose.

```console
cd $GOPATH/src/github.com/kubernetes-incubator/kompose/
script/godep-restore.sh
```

Reason for calling `script/godep-restore.sh` instead of just `godep restore` is that Kompose is using `github.com/openshift/kubernetes` instead of `github.com/kubernetes/kubernetes`.


2. Add a new dependency.

To add a new package, do this:

```console
go get foo/bar
# edit your code to import foo/bar
godep save ./...
```
