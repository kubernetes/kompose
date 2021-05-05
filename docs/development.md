---
layout: default
permalink: /development/
redirect_from: 
  - /docs/development.md/
---

# Development Guide

## Building Kompose

Read about building kompose [here](https://github.com/kubernetes/kompose#building).

## Workflow
### Fork the main repository

1. Go to https://github.com/kubernetes/kompose
2. Click the "Fork" button (at the top right)

### Clone your fork

The commands below require that you have $GOPATH. We highly recommended you put Kompose' code into your $GOPATH.

```console
git clone https://github.com/$YOUR_GITHUB_USERNAME/kompose.git $GOPATH/src/github.com/kubernetes/kompose
cd $GOPATH/src/github.com/kubernetes/kompose
git remote add upstream 'https://github.com/kubernetes/kompose'
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

Note: If you have write access to the main repository at github.com/kubernetes/kompose, you should modify your git configuration so that you can't accidentally push to upstream:

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

## Go Modules and dependency management

Kompose uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies.
If you want to make changes to dependencies please make sure that `go.mod` and `go.sum` are updated properly.

##### Updating Kubernetes and OpenShift
Kubernetes version depends on what version is OpenShift using.
OpenShift is using forked Kubernetes to carry some patches.
Currently it is not possible to use different Kubernetes version from version that OpenShift uses.
(for more see comments in `go.mod`)

### Adding CLI tests

[Kompose CLI tests](https://github.com/kubernetes/kompose/tree/master/script/test/cmd) run `kompose convert` with docker-compose files, and cross-check the k8s and OpenShift artifacts generated with the template files.

To generate CLI tests, please run `make gen-cmd`.

### CI

For Kompose, we use numerous CI's:

   - [TravisCI](https://travis-ci.org/kubernetes/kompose): Unit and CLI tests
   - [SemaphoreCI](https://semaphoreci.com/cdrage/kompose-2): Integration / cluster tests
   - [Fabric8CI](http://jenkins.cd.k8s.fabric8.io/): Secondary integration tests / future cluster tests
