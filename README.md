# Kompose (Kubernetes + Compose)

[![Build Status Widget]][Build Status] [![Coverage Status Widget]][Coverage Status] [![GoDoc Widget]][GoDoc] [![Slack Widget]][Slack]

`kompose` is a tool to help users who are familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). `kompose` takes a Docker Compose file and translates it into Kubernetes resources.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. Transformation of the Docker Compose format to Kubernetes resources manifest may not be exact, but it helps tremendously when first deploying an application on Kubernetes.

## Use Case

Convert [`docker-compose.yaml`](https://raw.githubusercontent.com/kubernetes-incubator/kompose/master/examples/docker-compose.yaml) into Kubernetes deployments and services with one simple command:

```sh
kompose convert -f docker-compose.yaml
INFO Kubernetes file "frontend-service.yaml" created         
INFO Kubernetes file "redis-master-service.yaml" created     
INFO Kubernetes file "redis-slave-service.yaml" created      
INFO Kubernetes file "frontend-deployment.yaml" created      
INFO Kubernetes file "redis-master-deployment.yaml" created  
INFO Kubernetes file "redis-slave-deployment.yaml" created 
```

Other examples are provided in the _examples_ [directory](./examples).

## Installation

We have multiple ways to install Kompose. Our prefered method is downloading the binary from the latest GitHub release.

Our entire list of installation methods are located in our [setup.md](/docs/setup.md) document.

Installation methods:
  - [Binary (Prefered method)](README.md)
  - [Go](/docs/setup.md#go)
  - [CentOS](/docs/setup.md#centos)
  - [Fedora](/docs/setup.md#fedora)
  - [macOS (Homebrew)](/docs/setup.md#macos)

#### Binary installation

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes-incubator/kompose/releases).

```sh
# Linux
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.7.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.7.0/kompose-darwin-amd64 -o kompose

# Windows
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.7.0/kompose-windows-amd64.exe -o kompose.exe

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

Alternatively, you can download the less-bandwidth intense [tarball](https://github.com/kubernetes-incubator/kompose/releases).

## Shell autocompletion

We support both Bash and Zsh autocompletion.

```sh
# Bash (add to .bashrc for persistence)
source <(kompose completion bash)

# Zsh (add to .zshrc for persistence)
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

## Documentation

Documentation can be found at our [kompose.io](http://kompose.io) website or our [docs](https://github.com/kubernetes-incubator/kompose/tree/master/docs) folder.

Here is a list of all available docs:

- [Quick start](docs/quickstart.md)
- [Setup](docs/setup.md)
- [User guide](docs/user-guide.md)
- [Conversion](docs/conversion.md)
- [Architecture](docs/architecture.md)
- [Development](docs/development.md)

## Community, Discussion, Contribution, and Support

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/kubernetes-incubator/kompose/issues).

As part of the Kubernetes ecosystem, we follow the Kubernetes community principles. More information can be found on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project on [Slack](http://slack.kubernetes.io) in channel #kompose

`kompose` is being incubated into the Kubernetes community via [SIG-APPS](https://github.com/kubernetes/community/tree/master/sig-apps) on [kubernetes/community](https://github.com/kubernetes/community).

[@ericchiang](https://github.com/ericchiang) is acting champion for [incubation](https://github.com/kubernetes/community/blob/master/incubator.md).

We do a biweekly community meeting which is [open to the public](https://bluejeans.com/404059616). Each week we outline what we have talked about in an [agenda doc](https://docs.google.com/document/d/1I5I21Cp_JZ9Az5MgMcu6Hl7m8WQ1Eqk_WeQLHenNom0/edit?usp=sharing). This meeting occurs every two weeks on Wednesday 18:00-19:00 GMT.

## Road Map

An up-to-date roadmap of upcoming releases is located at [ROADMAP.md](/ROADMAP.md).

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[Build Status]: https://travis-ci.org/kubernetes-incubator/kompose
[Build Status Widget]: https://travis-ci.org/kubernetes-incubator/kompose.svg?branch=master
[GoDoc]: https://godoc.org/github.com/kubernetes-incubator/kompose
[GoDoc Widget]: https://godoc.org/github.com/kubernetes-incubator/kompose?status.svg
[Slack]: http://slack.kubernetes.io#kompose
[Slack Widget]: https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png
[Coverage Status Widget]: https://coveralls.io/repos/github/kubernetes-incubator/kompose/badge.svg?branch=master
[Coverage Status]: https://coveralls.io/github/kubernetes-incubator/kompose?branch=master
