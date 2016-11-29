# Kompose (Kubernetes + Compose)

[![Build Status](https://travis-ci.org/kubernetes-incubator/kompose.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/kompose)

[![Join us in #kompose on k8s Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](http://slack.kubernetes.io) in #kompose channel

`kompose` is a tool to help users who are familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). `kompose` takes a Docker Compose file and translates it into Kubernetes resources.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. Transformation of the Docker Compose format to Kubernetes resources manifest may not be exact, but it helps tremendously when first deploying an application on Kubernetes.

## Use Case

If you have a Docker Compose [`docker-compose.yml`](./examples/docker-compose.yml) or a Docker Distributed Application Bundle [`docker-compose-bundle.dab`](./examples/docker-compose-bundle.dab) file, you can convert it into Kubernetes deployments and services like this:

```sh
$ kompose --bundle docker-compose-bundle.dab convert
WARN[0000]: Unsupported key networks - ignoring
file "redis-svc.json" created
file "web-svc.json" created
file "web-deployment.json" created
file "redis-deployment.json" created

$ kompose -f docker-compose.yml convert
WARN[0000]: Unsupported key networks - ignoring
file "redis-svc.json" created
file "web-svc.json" created
file "web-deployment.json" created
file "redis-deployment.json" created
```

Other examples are provided in the _examples_ [directory](./examples).

## Installation

To install `kompose`, you can either `go get` or install the binary from a new release.

#### Go

```sh
go get github.com/kubernetes-incubator/kompose
```

#### GitHub release

Grab the latest [release](https://github.com/kubernetes-incubator/kompose/releases) for your OS, untar and extract the binary.

Linux example:

```sh
=======
wget https://github.com/kubernetes-incubator/kompose/releases/download/v0.1.2/kompose_linux-amd64.tar.gz
tar -xvf kompose_linux-amd64.tar.gz --strip 1
sudo mv kompose /usr/local/bin
```

## Bash Completion

Running this below command in order to benefit from bash completion

```sh
$ PROG=kompose source script/bash_autocomplete
```

## Building

### Building with `go`

- `make`
- `go` v1.6 or later.
- `$GOPATH` set

Change to the Go Kompose path:

```sh
cd $GOPATH/src/github.com/kubernetes-incubator/kompose
```

Build via `make`:

```sh
make binary
```

Or `go build`:

```sh
go build -o kompose main.go
```

If you have `go` v1.5, you can still build `kompose`, but only with the following flags:

```sh
CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 go build -o kompose main.go
```

To create a multi-platform binary, use the `binary-cross` command via `make`:

```sh
make binary-cross
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

* September 15th 2016: Propose to Kubernetes incubator.
* September 30th 2016: Make the first official release of `kompose`, 0.1.0
* October 1st 2016: Add _build_ support connected to a private registry run by Kubernetes
* October 15th 2016: Add multi-container Pods, PVC and service types support.
* November 1st 2016: Add preference file to specify preferred resources for conversion and preferred provider.
* November 15th 2016: Improve support for Docker bundles to target specific image layers.
* December 24th 2016: Second release of `kompose`, 0.2.0

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
