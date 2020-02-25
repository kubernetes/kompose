# Kompose (Kubernetes + Compose)

[![Build Status Widget]][Build Status] [![Coverage Status Widget]][Coverage Status] [![GoDoc Widget]][GoDoc]  [![GoReportCard Widget]][GoReportCardResult]

`kompose` is a tool to help users who are familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). `kompose` takes a Docker Compose file and translates it into Kubernetes resources.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. Transformation of the Docker Compose format to Kubernetes resources manifest may not be exact, but it helps tremendously when first deploying an application on Kubernetes.

## Use Case

Convert [`docker-compose.yaml`](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml) into Kubernetes deployments and services with one simple command:

```sh
$ kompose convert -f docker-compose.yaml
INFO Kubernetes file "frontend-service.yaml" created         
INFO Kubernetes file "redis-master-service.yaml" created     
INFO Kubernetes file "redis-slave-service.yaml" created      
INFO Kubernetes file "frontend-deployment.yaml" created      
INFO Kubernetes file "redis-master-deployment.yaml" created  
INFO Kubernetes file "redis-slave-deployment.yaml" created 
```

Other examples are provided in the _examples_ [directory](./examples).

## Installation

We have multiple ways to install Kompose. Our preferred method is downloading the binary from the latest GitHub release.

Our entire list of installation methods are located in our [installation.md](/docs/installation.md) document.

Installation methods:
  - [Binary (Preferred method)](/docs/installation.md#github-release)
  - [Go](/docs/installation.md#go)
  - [CentOS](/docs/installation.md#centos)
  - [Fedora](/docs/installation.md#fedora)
  - [openSUSE/SLE](/docs/installation.md#opensusesle)
  - [macOS (Homebrew and MacPorts)](/docs/installation.md#macos)
  - [Windows](/docs/installation.md#windows)

#### Binary installation

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes/kompose/releases).

__Linux and macOS:__

```sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

__Windows:__

Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-windows-amd64.exe) and add the binary to your PATH.

Note: `kompose up` is not supported on windows.

## Shell autocompletion

We support both Bash and Zsh autocompletion.

```sh
# Bash (add to .bashrc for persistence)
source <(kompose completion bash)

# Zsh (add to .zshrc for persistence)
source <(kompose completion zsh)
```

## Development and building of Kompose

### Building with `go`
__Requisites:__
1. [make](https://www.gnu.org/software/make/)
2. [Golang](https://golang.org/) v1.6 or later
3. Set `GOPATH` correctly or click [SettingGOPATH](https://github.com/golang/go/wiki/SettingGOPATH) for details

__Steps:__
1. Clone repository
```console
$ git clone https://github.com/kubernetes/kompose.git $GOPATH/src/github.com/kubernetes/kompose
```
2. Build with `make`
```console
$ make bin
```
3. Or build with `go`
```console
$ go build -o kompose main.go
```
4. Test your changes
```console
$ make test
```

## Documentation

Documentation can be found at our [kompose.io](http://kompose.io) website or our [docs](https://github.com/kubernetes/kompose/tree/master/docs) folder.

Here is a list of all available docs:

- [Quick start](docs/getting-started.md)
- [Installation](docs/installation.md)
- [User guide](docs/user-guide.md)
- [Conversion](docs/conversion.md)
- [Architecture](docs/architecture.md)
- [Development](docs/development.md)

## Community, Discussion, Contribution, and Support

__Issues:__ If you find any issues, please [file it](https://github.com/kubernetes/kompose/issues).

__Kubernetes Community:__ As part of the Kubernetes ecosystem, we follow the Kubernetes community principles. More information can be found on the [community page](http://kubernetes.io/community/).

__Chat (Slack):__ We're fairly active on [Slack](http://slack.kubernetes.io#kompose) and you can find us in the #kompose channel.


### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[Build Status]: https://travis-ci.org/kubernetes/kompose
[Build Status Widget]: https://travis-ci.org/kubernetes/kompose.svg?branch=master
[GoDoc]: https://godoc.org/github.com/kubernetes/kompose
[GoDoc Widget]: https://godoc.org/github.com/kubernetes/kompose?status.svg
[Coverage Status Widget]: https://coveralls.io/repos/github/kubernetes/kompose/badge.svg?branch=master
[Coverage Status]: https://coveralls.io/github/kubernetes/kompose?branch=master
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/kubernetes/kompose
[GoReportCardResult]: https://goreportcard.com/report/github.com/kubernetes/kompose
