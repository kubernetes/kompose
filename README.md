# Kubernetes compose (Kompose)

[![Build Status](https://travis-ci.org/skippbox/kompose2.svg?branch=master)](https://travis-ci.org/skippbox/kompose2)

`kompose` is a tool to help users familiar with `docker-compose` move to [Kubernetes](http://kubernetes.io). It takes a Docker Compose file and translates it into Kubernetes objects, it can then submit those objects to a Kubernetes endpoint with the `kompose up` command.

`kompose` is a convenience tool to go from local Docker development to managing your application with Kubernetes. We don't assume that the transformation from docker compose format to Kubernetes API objects will be perfect, but it helps tremendously to start _Kubernetizing_ your application.

## Download

Grab the latest [release](https://github.com/skippbox/kompose2/releases)

## Usage

You need a Docker Compose file handy. There is a sample gitlab compose file in the `samples/` directory for testing.
You will convert the compose file to K8s objects with `kompose k8s convert`.

```bash
$ cd examples/

$ ls
docker-gitlab.yml

$ kompose convert -f docker-gitlab.yml -y

$ ls
docker-gitlab.yml       gitlab-rc.yaml   postgresql-deployment.yaml  postgresql-svc.yaml      redisio-rc.yaml
gitlab-deployment.yaml  gitlab-svc.yaml  postgresql-rc.yaml          redisio-deployment.yaml  redisio-svc.yaml
```

Next step you will submit these above objects to a kubernetes endpoint that is automatically taken from kubectl.

```bash
$ kompose up
```

Check that the replication controllers and services have been created.

```bash
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

```bash
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

```bash
$ kompose scale --scale 3 --rc redisio
Scaling redisio to: 3
$ kompose ps --rc
Name           Containers     Images                        Replicas  Selectors           
redisio        redisio        sameersbn/redis               3         service=redisio
```

Note that you can of course manage the services and replication controllers that have been created with `kubectl`.
The command of kompose have been extended to match the `docker-compose` commands.

## Alternate formats

The default `kompose` transformation will generate replication controllers and services and in format of json. You have alternative option to generate yaml with `-y`. Also, you can alternatively generate [Deployment](http://kubernetes.io/docs/user-guide/deployments/) objects, [DeamonSet](http://kubernetes.io/docs/admin/daemons/), [ReplicaSet](http://kubernetes.io/docs/user-guide/replicasets/) or [Helm](https://github.com/helm/helm) charts.

```bash
$ kompose convert -d -y
$ ls
$ tree
.
├── docker-compose.yml
├── redis-deployment.yaml
├── redis-rc.yaml
├── redis-svc.yaml
├── web-deployment.yaml
└── web-rc.yaml
```

The `*deployment.yaml` files contain the Deployments objects

```bash
$ kompose convert --ds -y
$ tree .
  .
  ├── redis-daemonset.yaml
  ├── redis-rc.yaml
  ├── redis-svc.yaml
  ├── web-daemonset.yaml
  ├── web-rc.yaml
  └── web-svc.yaml
```

The `*daemonset.yaml` files contain the DaemonSet objects

```bash
$ kompose convert --rs -y
$ tree .
.
├── redis-rc.yaml
├── redis-replicaset.yaml
├── redis-svc.yaml
├── web-rc.yaml
├── web-replicaset.yaml
└── web-svc.yaml

```

The `*replicaset.yaml` files contain the ReplicaSet objects

If you want to generate a Chart to be used with [Helm](https://github.com/kubernetes/helm) simply do:

```bash
$ kompose convert -c -y
$ tree docker-compose/
docker-compose/
├── Chart.yaml
├── README.md
└── templates
    ├── redis-rc.yaml
    ├── redis-svc.yaml
    └── web-rc.yaml
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

## Building

### Building with `go`

- You need `go` v1.5
- You need to set export `GO15VENDOREXPERIMENT=1` environment variable
- If your working copy is not in your `GOPATH`, you need to set it
accordingly.

```bash
$ go build -o kompose ./cli/main
```

### Building multi-platform binaries with make

- You need `make`

```
$ make binary
```

## Contributing and Issues

`kompose` is a work in progress, we will see how far it takes us. We welcome any pull request to make it even better.
If you find any issues, please [file it](https://github.com/skippbox/kompose2/issues).

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- Slack: #kubernetes-dev
- Mailing List: https://groups.google.com/forum/#!forum/kubernetes-dev

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
