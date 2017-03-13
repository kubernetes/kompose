# Kompose (Kubernetes + Compose)

[![Build Status](https://travis-ci.org/kubernetes-incubator/kompose.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/kompose)
[![Coverage Status](https://coveralls.io/repos/github/kubernetes-incubator/kompose/badge.svg?branch=master)](https://coveralls.io/github/kubernetes-incubator/kompose?branch=master)

[![Join us in #kompose on k8s Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](http://slack.kubernetes.io) in #kompose channel

`kompose` is a tool to help users who are familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). `kompose` takes a Docker Compose file and translates it into Kubernetes resources.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. Transformation of the Docker Compose format to Kubernetes resources manifest may not be exact, but it helps tremendously when first deploying an application on Kubernetes.

## Use Case

If you have a Docker Compose [`docker-compose.yml`](./examples/docker-compose.yml) or a Docker Distributed Application Bundle [`docker-compose-bundle.dab`](./examples/docker-compose-bundle.dab) file, you can convert it into Kubernetes deployments and services like this:

```console
$ kompose --bundle docker-compose-bundle.dab convert
WARN Unsupported key networks - ignoring
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "web-deployment.yaml" created
file "redis-deployment.yaml" created

$ kompose -f docker-compose.yml convert
WARN Unsupported key networks - ignoring
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "web-deployment.yaml" created
file "redis-deployment.yaml" created
```

Other examples are provided in the _examples_ [directory](./examples).

## Installation

To install `kompose`, you can either `go get` or install the binary from a new release.

#### Go

```sh
go get -u github.com/kubernetes-incubator/kompose
```

#### GitHub release

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes-incubator/kompose/releases).

```sh
# Linux 
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.3.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.3.0/kompose-darwin-amd64 -o kompose

# Windows
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.3.0/kompose-windows-amd64.exe -o kompose.exe
```

```sh
chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

#### CentOS

Kompose is in [EPEL](https://fedoraproject.org/wiki/EPEL) CentOS repository.
If you don't have [EPEL](https://fedoraproject.org/wiki/EPEL) repository already installed and enabled you can do it by running  `sudo yum install epel-release`

If you have [EPEL](https://fedoraproject.org/wiki/EPEL) enabled in your system, you can install Kompose like any other package.
```bash
sudo yum -y install kompose
```

#### Fedora
Kompose is in Fedora 24 and 25 repositories. You can install it just like any other package.

```bash
sudo dnf -y install kompose
```

## Shell autocompletion

We support both `bash` and `zsh` for autocompletion.

Bash:

```console
source <(kompose completion bash)
```

Zsh:

```console
source <(kompose completion zsh)
```

## Building

### Building with `go`

- You need `make`
- You need `go` v1.6 or later.
- If your working copy is not in your `GOPATH`, you need to set it accordingly.

You can either build via the Makefile:

```console
$ make bin
```

Or `go build`:

```console
$ go build -o kompose main.go
```

If you have `go` v1.5, it's still good to build `kompose` with the following settings:

```console
$ CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 go build -o kompose main.go
```

To create a multi-platform binary, use the `cross` command via `make`:

```console
$ make cross
```

## Contributing and Issues

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/kubernetes-incubator/kompose/issues).

## Community, Discussion, Contribution, and Support

As part of the Kubernetes ecosystem, we follow the Kubernetes community principles. More information can be found on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project on [Slack](http://slack.kubernetes.io) in channel #kompose

`kompose` is being incubated into the Kubernetes community via [SIG-APPS](https://github.com/kubernetes/community/tree/master/sig-apps) on [kubernetes/community](https://github.com/kubernetes/community).

[@ericchiang](https://github.com/ericchiang) is acting champion for [incubation](https://github.com/kubernetes/community/blob/master/incubator.md).

We do a biweekly community meeting which is [open to the public](https://bluejeans.com/404059616). Each week we outline what we have talked about in an [agenda doc](https://docs.google.com/document/d/1I5I21Cp_JZ9Az5MgMcu6Hl7m8WQ1Eqk_WeQLHenNom0/edit?usp=sharing). This meeting occurs every two weeks on Wednesday 18:00-19:00 GMT.

## Road Map

An up-to-date roadmap of upcoming releases is located at [ROADMAP.md](/ROADMAP.md).

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
