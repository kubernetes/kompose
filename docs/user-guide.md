# User Guide

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
WARN[0000]: Unsupported key networks - ignoring
WARN[0000]: Unsupported key build - ignoring
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
WARN[0000]: Unsupported key networks - ignoring
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
"build", "cap_add", "cap_drop", "cpuset", "cpu_shares", "cpu_quota", "cgroup_parent", "devices", "depends_on", "dns",
"dns_search", "domainname", "entrypoint", "env_file", "expose", "extends", "external_links", "extra_hosts", "hostname", "ipc",
"logging", "mac_address", "mem_limit", "memswap_limit", "network_mode", "networks", "pid", "security_opt", "shm_size",
"stop_signal", "volume_driver", "volumes_from", "uts", "read_only", "stdin_open", "tty", "user", "ulimits", "dockerfile",
"net", "args"
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
WARN[0000] Unsupported key build - ignoring             
WARN[0000] Unsupported key cap_add - ignoring           
WARN[0000] Unsupported key dockerfile - ignoring
```
