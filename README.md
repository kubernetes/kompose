# Kubernetes compose (Kompose)

[![Build Status](https://travis-ci.org/skippbox/kompose.svg?branch=master)](https://travis-ci.org/skippbox/kompose) [![Join us on Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](https://skippbox.herokuapp.com)

`kompose` is a tool to help users familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). It takes a Docker Compose file and translates it into Kubernetes objects, it can then submit those objects to a Kubernetes endpoint with the `kompose up` command.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. We don't assume that the transformation from docker compose format to Kubernetes API objects will be perfect, but it helps tremendously to start _Kubernetizing_ your application.

## Download

Grab the latest [release](https://github.com/skippbox/kompose/releases)

## Usage

Currently Kompose supports to transform either Docker Compose file (both of v1 and v2) and [experimental Distributed Application Bundles](https://blog.docker.com/2016/06/docker-app-bundle/) into Kubernetes objects. There is a couple of sample files in the `examples/` directory for testing. You will convert the compose or dab file to K8s objects with `kompose convert`.

```console
$ cd examples/

$ ls
docker-compose.yml  docker-compose-bundle.dsb  docker-gitlab.yml  docker-voting.yml

$ kompose convert -f docker-gitlab.yml -y
file "redisio-svc.yaml" created
file "gitlab-svc.yaml" created
file "postgresql-svc.yaml" created
file "gitlab-deployment.yaml" created
file "postgresql-deployment.yaml" created
file "redisio-deployment.yaml" created

$ ls *.yaml
gitlab-deployment.yaml  postgresql-deployment.yaml  redis-deployment.yaml    redisio-svc.yaml  web-deployment.yaml
gitlab-svc.yaml         postgresql-svc.yaml         redisio-deployment.yaml  redis-svc.yaml    web-svc.yaml
```

You can try with a Docker Compose version 2 like this:

```console
$ kompose convert --file docker-voting.yml
WARNING: Unsupported key Networks - ignoring
WARNING: Unsupported key Build - ignoring
file "worker-svc.json" created
file "db-svc.json" created
file "redis-svc.json" created
file "result-svc.json" created
file "vote-svc.json" created
file "redis-deployment.json" created
file "result-deployment.json" created
file "vote-deployment.json" created
file "worker-deployment.json" created
file "db-deployment.json" created

$ ls
db-deployment.json  docker-compose.yml         docker-gitlab.yml  redis-deployment.json  result-deployment.json  vote-deployment.json  worker-deployment.json
db-svc.json         docker-compose-bundle.dsb  docker-voting.yml  redis-svc.json         result-svc.json         vote-svc.json         worker-svc.json
```

Using `--bundle, --dab` to specify a DAB file as below:

```console
$ kompose convert --bundle docker-compose-bundle.dsb
WARNING: Unsupported key Networks - ignoring
file "redis-svc.json" created
file "web-svc.json" created
file "web-deployment.json" created
file "redis-deployment.json" created
```

## Alternate formats

The default `kompose` transformation will generate Kubernetes [Deployments](http://kubernetes.io/docs/user-guide/deployments/) and [Services](http://kubernetes.io/docs/user-guide/services/), in json format. You have alternative option to generate yaml with `-y`. Also, you can alternatively generate [Replication Controllers](http://kubernetes.io/docs/user-guide/replication-controller/) objects, [Deamon Sets](http://kubernetes.io/docs/admin/daemons/), or [Helm](https://github.com/helm/helm) charts.

```console
$ kompose convert 
file "redis-svc.json" created
file "web-svc.json" created
file "redis-deployment.json" created
file "web-deployment.json" created
```
The `*-deployment.json` files contain the Deployment objects.

```console
$ kompose convert --rc -y
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-rc.yaml" created
file "web-rc.yaml" created
```

The `*-rc.yaml` files contain the Replication Controller objects. If you want to specify replicas (default is 1), use `--replicas` flag: `$ kompose convert --rc --replicas 3 -y`

```console
$ kompose convert --ds -y
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-daemonset.yaml" created
file "web-daemonset.yaml" created
```

The `*-daemonset.yaml` files contain the Daemon Set objects

If you want to generate a Chart to be used with [Helm](https://github.com/kubernetes/helm) simply do:

```console
$ kompose convert -c -y
file "web-svc.yaml" created
file "redis-svc.yaml" created
file "web-deployment.yaml" created
file "redis-deployment.yaml" created
chart created in "./docker-compose/"

$ tree docker-compose/
docker-compose
├── Chart.yaml
├── README.md
└── templates
    ├── redis-deployment.yaml
    ├── redis-svc.yaml
    ├── web-deployment.yaml
    └── web-svc.yaml
```

The chart structure is aimed at providing a skeleton for building your Helm charts.

## Unsupported docker-compose configuration options

Currently `kompose` does not support the following Docker Compose options.

```
"Build", "CapAdd", "CapDrop", "CPUSet", "CPUShares", "ContainerName", "Devices", "DNS", "DNSSearch",
"Dockerfile", "DomainName", "Entrypoint", "EnvFile", "Hostname", "LogDriver", "MemLimit", "MemSwapLimit",
"Net", "Pid", "Uts", "Ipc", "ReadOnly", "StdinOpen", "SecurityOpt", "Tty", "User", "VolumeDriver",
"VolumesFrom", "Expose", "ExternalLinks", "LogOpt", "ExtraHosts",
```

For example:

```console
$ cat nginx.yml
nginx:
  image: nginx
  dockerfile: foobar
  build: ./foobar
  cap_add:
    - ALL
  container_name: foobar

$ kompose convert -f nginx.yml
WARNING: Unsupported key Build - ignoring
WARNING: Unsupported key CapAdd - ignoring
WARNING: Unsupported key ContainerName - ignoring
WARNING: Unsupported key Dockerfile - ignoring
```

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
$ make binary
```

## Contributing and Issues

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/skippbox/kompose/issues).

## Community, discussion, contribution, and support

We follow the Kubernetes community principles.

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://skippbox.kerokuapp.com): #kompose

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
