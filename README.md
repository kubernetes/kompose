# Kubernetes compose (Kompose)

[![Build Status](https://travis-ci.org/skippbox/kompose.svg?branch=master)](https://travis-ci.org/skippbox/kompose) [![Join us on Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](https://skippbox.herokuapp.com)

`kompose` is a tool to help users familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). It takes a Docker Compose file and translates it into Kubernetes objects, it can then submit those objects to a Kubernetes endpoint with the `kompose up` command.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. We don't assume that the transformation from docker compose format to Kubernetes API objects will be perfect, but it helps tremendously to start _Kubernetizing_ your application.

## Download

Grab the latest [release](https://github.com/skippbox/kompose/releases)

## Usage

You need a Docker Compose file handy. There is a sample gitlab compose file in the `examples/` directory for testing.
You will convert the compose file to K8s objects with `kompose convert`.

```console
$ cd examples/

$ ls
docker-gitlab.yml

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

Next step you will submit these above objects to a Kubernetes endpoint that is automatically taken from [`kubectl`](http://kubernetes.io/docs/user-guide/kubectl/kubectl/).

```console
$ kompose up
```

Check that the replication controllers and services have been created.

```console
$ kubectl get rc
NAME         DESIRED   CURRENT   AGE
gitlab       1         1         1m
postgresql   1         1         1m
redisio      1         1         1m

$ kubectl get svc
NAME         CLUSTER-IP     EXTERNAL-IP   PORT(S)               AGE
gitlab       10.0.247.129   nodes         10080/TCP,10022/TCP   1m
kubernetes   10.0.0.1       <none>        443/TCP               17h
postgresql   10.0.237.13    <none>        5432/TCP              1m
redisio      10.0.242.93    <none>        6379/TCP              1m
```

kompose also allows you to list the replication controllers and services with the `ps` subcommand.
You can delete them with the `delete` subcommand.

```console
$ kompose ps --rc
Name           Containers     Images                        Replicas  Selectors           
redisio        redisio        sameersbn/redis               1         service=redisio
postgresql     postgresql     sameersbn/postgresql:9.4-18   1         service=postgresql
gitlab         gitlab         sameersbn/gitlab:8.6.4        1         service: gitlab

$ kompose ps --svc
Name         Cluster IP          Ports                   Selectors
gitlab       10.0.247.129        TCP(10080),TCP(10022)   service=gitlab
postgresql   10.0.237.13         TCP(5432)               service=postgresql
redisio      10.0.242.93         TCP(6379)               service=redisio


$ kompose delete --rc --name gitlab
$ kompose ps --rc
Name           Containers     Images                        Replicas  Selectors
redisio        redisio        sameersbn/redis               1         service=redisio
postgresql     postgresql     sameersbn/postgresql:9.4-18   1         service=postgresql
```

And finally you can scale a replication controller with `scale`.

```console
$ kompose scale --scale 3 --rc redisio
Scaling redisio to: 3
$ kompose ps --rc
Name           Containers     Images                        Replicas  Selectors           
redisio        redisio        sameersbn/redis               3         service=redisio
```

Note that you can of course manage the services and replication controllers that have been created with `kubectl`.
The command of kompose have been extended to match the `docker-compose` commands.

## Alternate formats

The default `kompose` transformation will generate Kubernetes [Deployments](http://kubernetes.io/docs/user-guide/deployments/) and [Services](http://kubernetes.io/docs/user-guide/services/), in json format. You have alternative option to generate yaml with `-y`. Also, you can alternatively generate [Replication Controllers](http://kubernetes.io/docs/user-guide/replication-controller/) objects, [Deamon Sets](http://kubernetes.io/docs/admin/daemons/), [Replica Sets](http://kubernetes.io/docs/user-guide/replicasets/) or [Helm](https://github.com/helm/helm) charts.

```console
$ kompose convert 
file "redis-svc.json" created
file "web-svc.json" created
file "redis-deployment.json" created
file "web-deployment.json" created
```
The `*-deployment.json` files contain the Deployment objects.

```console
$ kompose convert --rc "1" -y
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-rc.yaml" created
file "web-rc.yaml" created
```

The `*-rc.yaml` files contain the Replication Controller objects.

```console
$ kompose convert --ds -y
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-daemonset.yaml" created
file "web-daemonset.yaml" created
```

The `*-daemonset.yaml` files contain the Daemon Set objects

```console
$ kompose convert --rs -y
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-replicaset.yaml" created
file "web-replicaset.yaml" created
```

The `*-replicaset.yaml` files contain the Replica Set objects

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

```
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
$ go build -o kompose ./cli/main
```

### Building multi-platform binaries with make

- You need `make`

```console
$ make binary
```

## Contributing and Issues

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/skippbox/kompose/issues).

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- Slack: #kubernetes-dev
- Mailing List: https://groups.google.com/forum/#!forum/kubernetes-dev

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
