---
layout: default
permalink: /user-guide/
title: User Guide
redirect_from:
  - /docs/user-guide.md/
  - /docs/user-guide/
---

# User Guide

## Table of contents

* [Kompose conversion example](#kompose-conversion-example)
* [CLI Modifications](#cli-modifications)
* [Labels](#labels)
* [Restart Policy](#restart-policy)
* [Building and Pushing Images](#building-and-pushing-images)

## Kompose Conversion Example

Kompose has support for two providers: OpenShift and Kubernetes.
You can choose a targeted provider using global option `--provider`. If no provider is specified, Kubernetes is set by default.

### Kubernetes

```sh
$ kompose --file compose.yaml convert
```

You can also provide multiple compose files at the same time:

```sh
$ kompose -f compose.yaml -f compose.yaml convert
```

When multiple compose files are provided the configuration is merged. Any configuration that is common will be over ridden by subsequent file.

You can provide your compose files via environment variables as following:
```sh
$ COMPOSE_FILE="compose.yaml alternative-compose.yaml" kompose convert
```

### OpenShift

```sh
$ kompose --provider openshift --file compose.yaml convert
```

## CLI Modifications

On the command line, you can modify the output of the generated YAML. For example, using alternative controllers such as [Replication Controllers](http://kubernetes.io/docs/user-guide/replication-controller/) objects, [Daemon Sets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/), or [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).


Daemon Set example:

```sh
$ kompose convert --controller daemonSet
```

A full list of these options can be found on `kompose convert --help`.

## Labels

`kompose` supports Kompose-specific labels within the `compose.yaml` file to get you the rest of the way there.

Labels are an important kompose concept as they allow you to add Kubernetes modifications without having to edit the YAML afterwards. For example, adding an init container, or a custom readiness check.

| Key / Value | Description / Example |
|-----|-------------|
| [`kompose.controller.port.expose`](#komposecontrollerportexpose) | Expose as hostPort on the controller (not recommended) |
| `Boolean` | `false` |
| [`kompose.controller.type`](#komposecontrollertype) | Type of the controller |
| `String` | `deployment`, `daemonset`, `replicationcontroller`, `statefulset` |
| [`kompose.cronjob.backoff_limit`](#komposecronjobbackoff_limit) | Number of retries before marked as failed |
| `Integer` | `6` |
| [`kompose.cronjob.concurrency_policy`](#komposecronjobconcurrency_policy) | Handling of concurrent jobs |
| `String` | `Forbid`, `Allow`, `Never` |
| [`kompose.cronjob.schedule`](#komposecronjobschedule) | Schedule |
| `String` | `1 * * * *` |
| [`kompose.hpa.cpu`](#komposehpacpu) | CPU utilization percentage that triggers autoscaling |
| `Percentage` | `50%` |
| [`kompose.hpa.memory`](#komposehpamemory) | Memory utilization threshold that triggers autoscaling |
| `String` | `200Mi` |
| [`kompose.hpa.replicas.max`](#komposehpareplicasmax) | Max pod replicas for Horizontal Pod Autoscaler |
| `Integer` | `10` |
| [`kompose.hpa.replicas.min`](#komposehpareplicasmin) | Min pod replicas for Horizontal Pod Autoscaler |
| `Integer` | `2` |
| [`kompose.image-pull-policy`](#komposeimage-pull-policy) | Policy for pulling images |
| `String` | `Always`, `IfNotPresent`, `Never` |
| [`kompose.image-pull-secret`](#komposeimage-pull-secret) | Secret to be used for pulling images from a private registry |
| `String` | `myregistrykey` |
| [`kompose.init.containers.command`](#komposeinitcontainerscommand) | Command to be executed |
| `Array` | `["printenv"]` |
| [`kompose.init.containers.image`](#komposeinitcontainersimage) | Image to be used |
| `String` | `busybox` |
| [`kompose.init.containers.name`](#komposeinitcontainersname) | Name assigned |
| `String` | `init-mydb` |
| [`kompose.security-context.fsgroup`](#komposesecurity-contextfsgroup) | Filesystem group ID for the pods' volumes |
| `Integer` | `1001` |
| [`kompose.service.external-traffic-policy`](#komposeserviceexternal-traffic-policy) | Policy to route external traffic |
| `String` | `cluster`, `local` |
| [`kompose.service.expose`](#komposeserviceexpose) | Creates a Ingress or Route. Accepts domain or 'true' for auto-generating a domain. |
| `String` | `true,domain1.com,domain2.com` |
| [`kompose.service.expose.ingress-class-name`](#komposeserviceexposeingress-class-name) | Ingress class to be used for exposing services |
| `String` | `nginx` |
| [`kompose.service.expose.tls-secret`](#komposeserviceexposetls-secret) | TLS secret for securing ingress |
| `String` | `my-tls-secret` |
| [`kompose.service.group`](#komposeservicegroup) | Label to group multiple containers in a single pod |
| `String` | `mygroup` |
| [`kompose.service.healthcheck.liveness.http_get_path`](#komposeservicehealthchecklivenesshttp_get_path) | HTTP GET path for liveness probe |
| `String` | `/health` |
| [`kompose.service.healthcheck.liveness.http_get_port`](#komposeservicehealthchecklivenesshttp_get_port) | HTTP GET port for liveness probe |
| `Integer` | `8080` |
| [`kompose.service.healthcheck.liveness.tcp_port`](#komposeservicehealthchecklivenesstcp_port) | TCP socket port for liveness probe |
| `Integer` | `3306` |
| [`kompose.service.healthcheck.readiness.disable`](#komposeservicehealthcheckreadinessdisable) | Whether to disable the readiness probe |
| `Boolean` | `true` |
| [`kompose.service.healthcheck.readiness.http_get_path`](#komposeservicehealthcheckreadinesshttp_get_path) | HTTP GET path for readiness probe |
| `String` | `/ready` |
| [`kompose.service.healthcheck.readiness.http_get_port`](#komposeservicehealthcheckreadinesshttp_get_port) | HTTP GET port for readiness probe |
| `Integer` | `8081` |
| [`kompose.service.healthcheck.readiness.interval`](#komposeservicehealthcheckreadinessinterval) | Interval between readiness checks |
| `Duration` | `10s` |
| [`kompose.service.healthcheck.readiness.retries`](#komposeservicehealthcheckreadinessretries) | Number of times readiness probe should retry before failing |
| `Integer` | `3` |
| [`kompose.service.healthcheck.readiness.start_period`](#komposeservicehealthcheckreadinessstart_period) | Initial delay before starting the readiness probe |
| `Duration` | `30s` |
| [`kompose.service.healthcheck.readiness.tcp_port`](#komposeservicehealthcheckreadinesstcp_port) | TCP socket port for readiness probe |
| `Integer` | `3307` |
| [`kompose.service.healthcheck.readiness.test`](#komposeservicehealthcheckreadinesstest) | Command or script run by the readiness probe |
| `Array` | `["CMD", "echo", "OK"]` |
| [`kompose.service.healthcheck.readiness.timeout`](#komposeservicehealthcheckreadinesstimeout) | Timeout for a single readiness probe |
| `Duration` | `5s` |
| [`kompose.service.nodeport.port`](#komposeservicenodeportport) | Specific port number to be used as NodePort |
| `Integer` | `30000` |
| [`kompose.service.type`](#komposeservicetype) | Type of service |
| `String` | `nodeport`, `clusterip`, `loadbalancer`, `headless` |
| [`kompose.volume.size`](#komposevolumesize) | Size of the volume |
| `String` | `1Gi` |
| [`kompose.volume.storage-class-name`](#komposevolumestorage-class-name) | StorageClassName for provisioning volumes |
| `String` | `standard` |
| [`kompose.volume.sub-path`](#komposevolumesub-path) | Subpath inside the mounted volume |
| `String` | `/data` |
| [`kompose.volume.type`](#komposevolumetype) | Type of Kubernetes volume |
| `String` | `configMap`, `persistentVolumeClaim`, `emptyDir`, `hostPath` |

### kompose.controller.port.expose

```yaml
services:
  web:
    image: wordpress:latest
    ports:
      - '80:80'
    labels:
      kompose.controller.expose.port: true
```

### kompose.controller.type

```yaml
services:
  web:
    image: wordpress:latest
    ports:
      - '80:80'
    labels:
      kompose.controller.type: deployment
```

### kompose.cronjob.backoff_limit

```yaml
services:
  cron-job:
    image: busybox
    labels:
      kompose.cronjob.backoff_limit: 3
```

### kompose.cronjob.concurrency_policy

```yaml
services:
  periodic-task:
    image: busybox
    labels:
      kompose.cronjob.concurrency_policy: Forbid
```

### kompose.cronjob.schedule

```yaml
services:
  cron-job:
    image: busybox
    labels:
      kompose.cronjob.schedule: "*/5 * * * *"
```

### kompose.hpa.cpu

```yaml
services:
  web:
    image: nginx
    labels:
      kompose.hpa.cpu: 80
```

### kompose.hpa.memory

```yaml
services:
  db:
    image: mysql
    labels:
      kompose.hpa.memory: 512Mi
```

### kompose.hpa.replicas.max

```yaml
services:
  api:
    image: custom-api
    labels:
      kompose.hpa.replicas.max: 10
```

### kompose.hpa.replicas.min

```yaml
services:
  api:
    image: custom-api
    labels:
      kompose.hpa.replicas.min: 2
```

### kompose.image-pull-policy

```yaml
services:
  example-service:
    image: example-image
    labels:
      kompose.image-pull-policy: "IfNotPresent"
```

### kompose.image-pull-secret

```yaml
services:
  private-service:
    image: private-repo/image:tag
    labels:
      kompose.image-pull-secret: "my-private-registry-key"
```

### kompose.init.containers.command

```yaml
services:
  init-service:
    image: busybox
    labels:
      kompose.init.containers.command: ["echo", "Initializing..."]
```

### kompose.init.containers.image

```yaml
services:
  init-service:
    image: busybox
    labels:
      kompose.init.containers.image: busybox
```

### kompose.init.containers.name

```yaml
services:
  init-service:
    image: busybox
    labels:
      kompose.init.containers.name: "initial-setup"
```

### kompose.security-context.fsgroup

```yaml
services:
  secured-service:
    image: nginx
    labels:
      kompose.security-context.fsgroup: 2000
```

### kompose.service.external-traffic-policy

```yaml
services:
  front-end:
    image: quay.io/kompose/web
    ports:
      - 8080:8080
    labels:
      kompose.service.external-traffic-policy: local
```

### kompose.service.expose

```yaml
services:
  web-app:
    image: nginx
    ports:
      - 80:80
    labels:
      kompose.service.expose: "example.com"
```

### kompose.service.expose.ingress-class-name

```yaml
services:
  web:
    image: nginx
    ports:
      - 80:80
    labels:
      kompose.service.expose.ingress-class-name: "nginx"
```

### kompose.service.expose.tls-secret

```yaml
services:
  web:
    image: nginx
    ports:
      - 443:443
    labels:
      kompose.service.expose.tls-secret: "my-ssl-secret"
```

### kompose.service.group

```yaml
version: "3"

services:
  nginx:
    image: nginx
    depends_on:
      - logs
    labels:
      - kompose.service.group=sidecar

  logs:
    image: busybox
    command: ["tail -f /var/log/nginx/access.log"]
    labels:
      - kompose.service.group=sidecar
```

### kompose.service.healthcheck.liveness.http_get_path

```yaml
services:
  web:
    image: custom-web
    ports:
      - "8080:8080"
    labels:
      kompose.service.healthcheck.liveness.http_get_path: /health
```

### kompose.service.healthcheck.liveness.http_get_port

```yaml
services:
  web:
    image: custom-web
    ports:
      - "8080:8080"
    labels:
      kompose.service.healthcheck.liveness.http_get_port: 8080
```

### kompose.service.healthcheck.liveness.tcp_port

```yaml
services:
  db:
    image: mysql
    ports:
      - "3306:3306"
    labels:
      kompose.service.healthcheck.liveness.tcp_port: 3306
```

### kompose.service.healthcheck.readiness.disable

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.disable: true
```

### kompose.service.healthcheck.readiness.http_get_path

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.http_get_path: /ready
```

### kompose.service.healthcheck.readiness.http_get_port

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.http_get_port: 8081
```

### kompose.service.healthcheck.readiness.interval

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.interval: 10s
```

### kompose.service.healthcheck.readiness.retries

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.retries: 3
```

### kompose.service.healthcheck.readiness.start_period

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.start_period: 30s
```

### kompose.service.healthcheck.readiness.tcp_port

```yaml
services:
  db:
    image: mysql
    labels:
      kompose.service.healthcheck.readiness.tcp_port: 3307
```

### kompose.service.healthcheck.readiness.test

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.test: ["CMD", "curl", "-f", "http://localhost:8081/ready"]
```

### kompose.service.healthcheck.readiness.timeout

```yaml
services:
  web:
    image: custom-web
    labels:
      kompose.service.healthcheck.readiness.timeout: 5s
```

### kompose.service.nodeport.port

```yaml
services:
  web:
    image: nginx
    ports:
      - "30000:80"
    labels:
      kompose.service.nodeport.port: 30000
```

### kompose.service.type

```yaml
services:
  web:
    image: nginx
    ports:
      - "80:80"
    labels:
      kompose.service.type: nodeport
```

### kompose.volume.size

```yaml
services:
  db:
    image: postgres:10.1
    labels:
      kompose.volume.size: 1Gi
    volumes:
      - db-data:/var/lib/postgresql/data
```

### kompose.volume.storage-class-name

```yaml
services:
  db:
    image: postgres:10.1
    labels:
      kompose.volume.storage-class-name: custom-storage-class-name
    volumes:
      - db-data:/var/lib/postgresql/data
```

### kompose.volume.sub-path

```yaml
services:
  pgadmin:
    image: postgres
    labels:
      kompose.volume.sub-path: pg-data
```

### kompose.volume.type

```yaml
services:
  db:
    image: postgres
    labels:
      kompose.volume.type: persistentVolumeClaim
    volumes:
      - db-data:/var/lib/postgresql/data
```

## Restart Policy

If you want to create normal pods without a controller you can use the `restart` construct of compose to define that. Follow the table below to see what happens on the `restart` value.

| `compose` `restart` | object created    | Pod `restartPolicy` |
|---------------------|-------------------|---------------------|
| `""`                | controller object | `Always`            |
| `always`            | controller object | `Always`            |
| `unless-stopped`    | controller object | `Always`            |
| `on-failure`        | Pod / CronJob     | `OnFailure`         |
| `no`                | Pod / CronJob     | `Never`             |

**Note**: controller object could be `deployment`, `replicationcontroller`, etc.

For example, the `pival` service will become a pod down here. This container calculated the value of `pi`.

```yaml
version: '2'

services:
  pival:
    image: perl
    command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
    restart: "on-failure"
```

For example, the `pival` service will become a cron job down here. This container calculated the value of `pi` every minute.

```yaml
version: '2'

services:
  pival:
    image: perl
    command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
    restart: "no"
    labels:
      kompose.cronjob.schedule: "* * * * *"
      kompose.cronjob.concurrency_policy: "Forbid"
      kompose.cronjob.backoff_limit: "0"
```

#### Warning about Deployment Configs

If the Compose file has a volume specified for a service, the Deployment (Kubernetes) or DeploymentConfig (OpenShift) strategy is changed to "Recreate" instead of "RollingUpdate" (default). This is done to avoid multiple instances of a service from accessing a volume at the same time.

If the Compose file has a service name with `_` or `.` in it (e.g., `web_service` or `web.service`), then it will be replaced by `-` and the service name will be renamed accordingly (e.g., `web-service`). Kompose does this because "Kubernetes" doesn't allow `_` in object names.

Please note that changing the service name might break some `compose` files.

## Building and Pushing Images

If the Compose file has `build` or `build:context, build:dockerfile` keys, build will run when `--build` specified.

And Image will push to _docker.io_ (default) when `--push-image=true` specified.

It is possible to push to a custom registry by specifying `--push-image-registry`, which will override the registry from the image name.

### Authentication on Registry

Kompose uses the docker authentication from file `$DOCKER_CONFIG/config.json`, `$HOME/.docker/config.json`, and `$HOME/.dockercfg` after `docker login`.

**This only works fine on Linux but macOS would fail when using `"credsStore": "osxkeychain"`.**

However, there is an approach to push successfully on macOS, by not using `osxkeychain` for `credsStore`. To disable `osxkeychain`:

- remove `credsStore` from the `config.json` file, and `docker login` again.
- for some docker desktop versions, there is a setting `Securely store Docker logins in macOS keychain`, which should be unchecked. Then restart docker desktop if needed, and `docker login` again.

Now `config.json` should contain base64 encoded passwords, then push image should succeed. Working, but not safe though! Use it at your risk.

For Windows, there is also `credsStore` which is `wincred`. Technically it will fail on authentication as macOS does, but you can try the approach above like macOS too.

### Custom Build and Push

If you want to customize the build and push processes and use another containers solution than Docker,
Kompose offers you the possibility to do that. You can use `--build-command` and `--push-command` flags
to achieve that.

e.g: `kompose -f convert --build-command 'whatever command --you-use' --push-command 'whatever command --you-use'`