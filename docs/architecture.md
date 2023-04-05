---
layout: default
permalink: /architecture/
title: Architecture
redirect_from:
  - /docs/architecture.md/
  - /docs/architecture/
---

# Architecture and Internal Design

* TOC
{:toc}

`kompose` has 3 stages: _Loader_, _Transformer_ and _Outputter_. Each stage should have a well-defined interface, so it is easy to write a new Loader, Transformer, or Outputters and plug it in. Currently, only Loader and Transformer interfaces are defined.

![Design Diagram](https://raw.githubusercontent.com/kubernetes/kompose/main/docs/images/design_diagram.png)

## Loader

The Loader reads the input file now `kompose` supports [Docker Compose](https://docs.docker.com/compose) v1, v2 and converts it to KomposeObject.

Loader is represented by a Loader interface:

```go
type Loader interface {
      LoadFile(file string) kobject.KomposeObject
}
```

Every loader “implementation” should be placed into `kompose/pkg/loader` (like compose). More input formats will be supported in the future. You can take a look for more details at:

- [kompose/pkg/loader](https://github.com/kubernetes/kompose/tree/master/pkg/loader)
- [kompose/pkg/loader/compose](https://github.com/kubernetes/kompose/tree/master/pkg/loader/compose)

## KomposeObject

`KomposeObject` is Kompose internal representation of all containers loaded from input file. First version of `KomposeObject` looks like this (source: [kobject.go](https://github.com/kubernetes/kompose/blob/master/pkg/kobject/kobject.go)):

```go
// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
    ServiceConfigs map[string]ServiceConfig
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
    ContainerName string
    Image         string
    Environment   []EnvVar
    Port          []Ports
    Command       []string
    WorkingDir    string
    Args          []string
    Volumes       []string
    Network       []string
    Labels        map[string]string
    Annotations   map[string]string
    CPUSet        string
    CPUShares     int64
    CPUQuota      int64
    CapAdd        []string
    CapDrop       []string
    Entrypoint    []string
    Expose        []string
    Privileged    bool
    Restart       string
    User          string
}
```

## Transformer

The Transformer takes KomposeObject and converts it to target/output format (currently, there are sets of Kubernetes/OpenShift objects). Similar to the `Loader`, Transformer is represented by a Transformer interface:

```go
type Transformer interface {
     Transform(kobject.KomposeObject, kobject.ConvertOptions) []runtime.Object
}
```

If you wish to add more providers containing different kinds of objects, the Transformer would be the place to look into. Currently, Kompose supports Kubernetes (by default) and OpenShift providers. More details at:

- [kompose/pkg/transformer](https://github.com/kubernetes/kompose/tree/master/pkg/transformer)
- [kompose/pkg/transformer/Kubernetes](https://github.com/kubernetes/kompose/tree/master/pkg/transformer/kubernetes)
- [kompose/pkg/transformer/openshift](https://github.com/kubernetes/kompose/tree/master/pkg/transformer/openshift)

## Outputter

The Outputter takes the Transformer result and executes the given action. For example, action can display results to stdout or directly deploy artifacts to Kubernetes/OpenShift.
