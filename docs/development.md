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

## `glide`, `glide-vc` and dependency management

Kompose uses `glide` to manage dependencies and `glide-vc` to clean vendor directory.
They are not strictly required for building Kompose but they are required when managing dependencies under the `vendor/` directory.
If you want to make changes to dependencies please make sure that `glide` and `glide-vc` are installed and are in your `$PATH`.

### Installing glide

There are many ways to build and host golang binaries. Here is an easy way to get utilities like `glide` and `glide-vc` installed:

Ensure that Mercurial and Git are installed on your system. (some of the dependencies use the mercurial source control system).
Use `apt-get install mercurial git` or `yum install mercurial git` on Linux, or `brew.sh` on OS X, or download them directly.

```console
go get -u github.com/Masterminds/glide
go get github.com/sgotti/glide-vc
```

Check that `glide` and `glide-vc` commands are working.

```console
glide --version
glide-vc -h
```

### Using glide

#### Adding new dependency
1. Update `glide.yml` file.

  Add new packages or subpackages to `glide.yml` depending if you added whole new package as dependency or
  just new subpackage.

2. Run `glide update --strip-vendor` to get new dependencies.
   Then run `glide-vc --only-code --no-tests` to delete all unnecessary files from vendor.

3. Commit updated `glide.yml`, `glide.lock` and `vendor` to git.


#### Updating dependencies

1. Set new package version in  `glide.yml` file.

2. Run `glide update --strip-vendor` to update dependencies.
   Then run `glide-vc --only-code --no-tests` to delete all unnecessary files from vendor.


##### Updating Kubernetes and OpenShift
Kubernetes version depends on what version is OpenShift using.
OpenShift is using forked Kubernetes to carry some patches.
Currently it is not possible to use different Kubernetes version from version that OpenShift uses.
(for more see comments in `glide.yml`)

### Adding CLI tests

[Kompose CLI tests](https://github.com/kubernetes/kompose/tree/master/script/test/cmd) run `kompose convert` with docker-compose files, and cross-check the k8s and OpenShift artifacts generated with the template files.

To generate CLI tests, please run `make gen-cmd`.

### CI

For Kompose, we use numerous CI's:

   - [TravisCI](https://travis-ci.org/kubernetes/kompose): Unit and CLI tests
   - [SemaphoreCI](https://semaphoreci.com/cdrage/kompose-2): Integration / cluster tests
   - [Fabric8CI](http://jenkins.cd.k8s.fabric8.io/): Secondary integration tests / future cluster tests
