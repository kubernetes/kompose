# Kubernetes compose (Kompose)

[![Build Status](https://travis-ci.org/skippbox/kompose.svg?branch=master)](https://travis-ci.org/skippbox/kompose) [![Join us on Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](https://skippbox.herokuapp.com)

`kompose` is a tool to help users familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). It takes a Docker Compose file and translates it into Kubernetes resources.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. We don't assume that the transformation from docker compose format to Kubernetes API objects will be perfect, but it helps tremendously to start _Kubernetizing_ your application.

## Use Case

For example, if you have a Docker bundle like [`docker-compose-bundle.dsb`](./examples/docker-compose-bundle.dsb), you can convert it into Kubernetes deployments and services like this:

```console
$ kompose convert --bundle docker-compose-bundle.dsb
WARN[0000]: Unsupported key networks - ignoring
file "redis-svc.json" created
file "web-svc.json" created
file "web-deployment.json" created
file "redis-deployment.json" created
```

Other examples are provided in the _examples_ [directory](./examples)

## Download

Grab the latest [release](https://github.com/skippbox/kompose/releases)

## Bash completion
Running this below command in order to benefit from bash completion

```console
$ PROG=kompose source script/bash_autocomplete
```

## Building

### Building with `go`

- You need `go` v1.5
- You need to set export `GO15VENDOREXPERIMENT=1` environment variable
- If your working copy is not in your `GOPATH`, you need to set it
accordingly.

```console
$ go build -tags experimental -o kompose ./cli/main
```

You need `-tags experimental` because the current `bundlefile` package of docker/libcompose is still experimental.

### Building multi-platform binaries with make

- You need `make`

```console
$ make binary-cross
```

## Contributing and Issues

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/skippbox/kompose/issues).

## Community, discussion, contribution, and support

We follow the Kubernetes community principles.

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://skippbox.herokuapp.com): #kompose

Incubation of `kompose` into the Kubernetes project will be discussed in the [API Machinery SIG](https://github.com/kubernetes/community)

## RoadMap

* September 15th 2016: Enter Kubernetes incubator.
* September 30th 2016: Make the first official release of `kompose`, 0.1.0
* October 1st 2016: Add _build_ support connected to a private registry run by Kubernetes
* October 15th 2016: Add support for Rancher compose/provider
* November 1st 2016: Add preference file to specify preferred resources for conversion and preferred provider.
* November 15th 2016: Improve support for Docker bundles to target specific image layers.
* December 24th 2016: Second release of `kompose`, 0.2.0

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
