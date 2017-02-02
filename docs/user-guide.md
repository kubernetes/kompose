---
layout: default
permalink: /user-guide/
---

# User Guide

- [Usage](#user-guide)
  * [Kompose convert](#kompose-convert)
  * [Kompose up](#kompose-up)
  * [Kompose down](#kompose-down)
- [Alternate formats](#alternate-formats)
- [Unsupported docker-compose configuration options](#unsupported-docker-compose-configuration-options)


Kompose has support for two providers: OpenShift and Kubernetes.
You can choose targeted provider either using global option `--provider`, or by setting environment variable `PROVIDER`.
By setting environment variable `PROVIDER` you can permanently switch to OpenShift provider without need to always specify `--provider openshift` option.
If no provider is specified Kubernetes is default provider.


## Kompose convert

Currently Kompose supports to transform either Docker Compose file (both of v1 and v2) and [experimental Distributed Application Bundles](https://blog.docker.com/2016/06/docker-app-bundle/) into Kubernetes and OpenShift objects.
There is a couple of sample files in the `examples/` directory for testing.
You will convert the compose or dab file to Kubernetes or OpenShift objects with `kompose convert`.

### Kubernetes
```console
$ cd examples/

$ ls
docker-compose.yml  docker-compose-bundle.dab  docker-gitlab.yml  docker-voting.yml

$ kompose -f docker-gitlab.yml convert
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
$ kompose --file docker-voting.yml convert
WARN Unsupported key networks - ignoring
WARN Unsupported key build - ignoring
file "worker-svc.yaml" created
file "db-svc.yaml" created
file "redis-svc.yaml" created
file "result-svc.yaml" created
file "vote-svc.yaml" created
file "redis-deployment.yaml" created
file "result-deployment.yaml" created
file "vote-deployment.yaml" created
file "worker-deployment.yaml" created
file "db-deployment.yaml" created

$ ls
db-deployment.yaml  docker-compose.yml         docker-gitlab.yml  redis-deployment.yaml  result-deployment.yaml  vote-deployment.yaml  worker-deployment.yaml
db-svc.yaml         docker-compose-bundle.dab  docker-voting.yml  redis-svc.yaml         result-svc.yaml         vote-svc.yaml         worker-svc.yaml
```

You can also provide multiple docker-compose files at the same time:

```console
$ kompose -f docker-compose.yml -f docker-guestbook.yml convert
file "frontend-service.yaml" created         
file "mlbparks-service.yaml" created         
file "mongodb-service.yaml" created          
file "redis-master-service.yaml" created     
file "redis-slave-service.yaml" created      
file "frontend-deployment.yaml" created      
file "mlbparks-deployment.yaml" created      
file "mongodb-deployment.yaml" created       
file "mongodb-claim0-persistentvolumeclaim.yaml" created 
file "redis-master-deployment.yaml" created  
file "redis-slave-deployment.yaml" created   

$ ls
mlbparks-deployment.yaml  mongodb-service.yaml                       redis-slave-service.jsonmlbparks-service.yaml  
frontend-deployment.yaml  mongodb-claim0-persistentvolumeclaim.yaml  redis-master-service.yaml
frontend-service.yaml     mongodb-deployment.yaml                    redis-slave-deployment.yaml
redis-master-deployment.yaml
``` 

When multiple docker-compose files are provided the configuration is merged. Any configuration that is common will be over ridden by subsequent file.
 
Using `--bundle, --dab` to specify a DAB file as below:

```console
$ kompose --bundle docker-compose-bundle.dab convert
WARN: Unsupported key networks - ignoring
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "web-deployment.yaml" created
file "redis-deployment.yaml" created
```

### OpenShift

```console
$ kompose --provider openshift --file docker-voting.yml convert
WARN [worker] Service cannot be created because of missing port.
INFO file "vote-service.yaml" created             
INFO file "db-service.yaml" created               
INFO file "redis-service.yaml" created            
INFO file "result-service.yaml" created           
INFO file "vote-deploymentconfig.yaml" created    
INFO file "vote-imagestream.yaml" created         
INFO file "worker-deploymentconfig.yaml" created  
INFO file "worker-imagestream.yaml" created       
INFO file "db-deploymentconfig.yaml" created      
INFO file "db-imagestream.yaml" created           
INFO file "redis-deploymentconfig.yaml" created   
INFO file "redis-imagestream.yaml" created        
INFO file "result-deploymentconfig.yaml" created  
INFO file "result-imagestream.yaml" created  
```

In similar way you can convert DAB files to OpenShift.
```console
$ kompose --bundle docker-compose-bundle.dab --provider openshift convert
WARN: Unsupported key networks - ignoring
INFO file "redis-svc.yaml" created
INFO file "web-svc.yaml" created
INFO file "web-deploymentconfig.yaml" created
INFO file "web-imagestream.yaml" created           
INFO file "redis-deploymentconfig.yaml" created
INFO file "redis-imagestream.yaml" created
```

It also supports creating buildconfig for build directive in a service. By default, it uses the remote repo for the current git branch as the source repo, and the current branch as the source branch for the build. You can specify a different source repo and branch using ``--build-repo`` and ``--build-branch`` options respectively.

```console
kompose --provider openshift --file buildconfig/docker-compose.yml convert
WARN [foo] Service cannot be created because of missing port. 
INFO Buildconfig using git@github.com:rtnpro/kompose.git::master as source. 
INFO file "foo-deploymentconfig.yaml" created     
INFO file "foo-imagestream.yaml" created          
INFO file "foo-buildconfig.yaml" created 
```

**Note**: If you are manually pushing the Openshift artifacts using ``oc create -f``, you need to ensure that you push the imagestream artifact before the buildconfig artifact, to workaround this Openshift issue: https://github.com/openshift/origin/issues/4518 .

## Kompose up

Kompose supports a straightforward way to deploy your "composed" application to Kubernetes or OpenShift via `kompose up`.


### Kubernetes
```console
$ kompose --file ./examples/docker-guestbook.yml up
We are going to create Kubernetes deployments and services for your Dockerized application.
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead.

INFO Successfully created service: redis-master   
INFO Successfully created service: redis-slave    
INFO Successfully created service: frontend       
INFO Successfully created deployment: redis-master
INFO Successfully created deployment: redis-slave
INFO Successfully created deployment: frontend    

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods' for details.

$ kubectl get deployment,svc,pods
NAME                            DESIRED       CURRENT       UP-TO-DATE   AVAILABLE   AGE
frontend                        1             1             1            1           4m
redis-master                    1             1             1            1           4m
redis-slave                     1             1             1            1           4m
NAME                            CLUSTER-IP    EXTERNAL-IP   PORT(S)      AGE
frontend                        10.0.174.12   <none>        80/TCP       4m
kubernetes                      10.0.0.1      <none>        443/TCP      13d
redis-master                    10.0.202.43   <none>        6379/TCP     4m
redis-slave                     10.0.1.85     <none>        6379/TCP     4m
NAME                            READY         STATUS        RESTARTS     AGE
frontend-2768218532-cs5t5       1/1           Running       0            4m
redis-master-1432129712-63jn8   1/1           Running       0            4m
redis-slave-2504961300-nve7b    1/1           Running       0            4m
```
Note:
- You must have a running Kubernetes cluster with a pre-configured kubectl context.
- Only deployments and services are generated and deployed to Kubernetes. If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead.

### OpenShift
```console
$kompose --file ./examples/docker-guestbook.yml --provider openshift up
We are going to create OpenShift DeploymentConfigs and Services for your Dockerized application.
If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead.

INFO Successfully created service: redis-slave    
INFO Successfully created service: frontend       
INFO Successfully created service: redis-master   
INFO Successfully created deployment: redis-slave
INFO Successfully created ImageStream: redis-slave
INFO Successfully created deployment: frontend    
INFO Successfully created ImageStream: frontend   
INFO Successfully created deployment: redis-master
INFO Successfully created ImageStream: redis-master

Your application has been deployed to OpenShift. You can run 'oc get dc,svc,is' for details.

$ oc get dc,svc,is
NAME               REVISION                              DESIRED       CURRENT    TRIGGERED BY
dc/frontend        0                                     1             0          config,image(frontend:v4)
dc/redis-master    0                                     1             0          config,image(redis-master:e2e)
dc/redis-slave     0                                     1             0          config,image(redis-slave:v1)
NAME               CLUSTER-IP                            EXTERNAL-IP   PORT(S)    AGE
svc/frontend       172.30.46.64                          <none>        80/TCP     8s
svc/redis-master   172.30.144.56                         <none>        6379/TCP   8s
svc/redis-slave    172.30.75.245                         <none>        6379/TCP   8s
NAME               DOCKER REPO                           TAGS          UPDATED
is/frontend        172.30.12.200:5000/fff/frontend                     
is/redis-master    172.30.12.200:5000/fff/redis-master                 
is/redis-slave     172.30.12.200:5000/fff/redis-slave    v1  
```

Note:
- You must have a running OpenShift cluster with a pre-configured `oc` context (`oc login`)

## Kompose down

Once you have deployed "composed" application to Kubernetes, `kompose down` will help you to take the application out by deleting its deployments and services. If you need to remove other resources, use the 'kubectl' command.

```console
$ kompose --file docker-guestbook.yml down
INFO Successfully deleted service: redis-master   
INFO Successfully deleted deployment: redis-master
INFO Successfully deleted service: redis-slave    
INFO Successfully deleted deployment: redis-slave
INFO Successfully deleted service: frontend       
INFO Successfully deleted deployment: frontend
```
Note:
- You must have a running Kubernetes cluster with a pre-configured kubectl context.

## Alternate formats

The default `kompose` transformation will generate Kubernetes [Deployments](http://kubernetes.io/docs/user-guide/deployments/) and [Services](http://kubernetes.io/docs/user-guide/services/), in yaml format. You have alternative option to generate json with `-j`. Also, you can alternatively generate [Replication Controllers](http://kubernetes.io/docs/user-guide/replication-controller/) objects, [Deamon Sets](http://kubernetes.io/docs/admin/daemons/), or [Helm](https://github.com/helm/helm) charts.

```console
$ kompose convert -j
file "redis-svc.json" created
file "web-svc.json" created
file "redis-deployment.json" created
file "web-deployment.json" created
```
The `*-deployment.json` files contain the Deployment objects.

```console
$ kompose convert --rc 
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-rc.yaml" created
file "web-rc.yaml" created
```

The `*-rc.yaml` files contain the Replication Controller objects. If you want to specify replicas (default is 1), use `--replicas` flag: `$ kompose convert --rc --replicas 3`

```console
$ kompose convert --ds 
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "redis-daemonset.yaml" created
file "web-daemonset.yaml" created
```

The `*-daemonset.yaml` files contain the Daemon Set objects

If you want to generate a Chart to be used with [Helm](https://github.com/kubernetes/helm) simply do:

```console
$ kompose convert -c 
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
"build", "cgroup_parent", "devices", "depends_on", "dns", "dns_search", "domainname", "env_file", "extends", "external_links", "extra_hosts", "hostname", "ipc", "logging", "mac_address", "mem_limit", "memswap_limit", "network_mode", "networks", "pid", "security_opt", "shm_size", "stop_signal", "volume_driver", "uts", "read_only", "stdin_open", "tty", "user", "ulimits", "dockerfile", "net"
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

$ kompose -f nginx.yml convert
WARN Unsupported key build - ignoring             
WARN Unsupported key cap_add - ignoring           
WARN Unsupported key dockerfile - ignoring
```

## Labels

`kompose` supports Kompose-specific labels within the `docker-compose.yml` file in order to explicitly define a service's behavior upon conversion.

- kompose.service.type defines the type of service to be created.

For example:

```yaml
version: "2"
services: 
  nginx:
    image: nginx
    dockerfile: foobar
    build: ./foobar
    cap_add:
      - ALL
    container_name: foobar
    labels: 
      kompose.service.type: nodeport
```

- kompose.service.expose defines if the service needs to be made accessible from outside the cluster or not. If the value is set to "true", the provider sets the endpoint automatically, and for any other value, the value is set as the hostname. If multiple ports are defined in a service, the first one is chosen to be the exposed.
    - For the Kubernetes provider, an ingress resource is created and it is assumed that an ingress controller has already been configured.
    - For the OpenShift provider, a route is created.

For example:

```yaml
version: "2"
services:
  web:
    image: tuna/docker-counter23
    ports:
     - "5000:5000"
    links:
     - redis
    labels:
      kompose.service.expose: "counter.example.com"
  redis:
    image: redis:3.0
    ports:
     - "6379"
```

The currently supported options are:

| Key                  | Value                               |
|----------------------|-------------------------------------|
| kompose.service.type | nodeport / clusterip / loadbalancer |
| kompose.service.expose| true / hostname |


## Restart

If you want to create normal pods without controllers you can use `restart` construct of docker-compose to define that. Follow table below to see what heppens on the `restart` value.

| `docker-compose` `restart` | object created    | Pod `restartPolicy` |
|----------------------------|-------------------|---------------------|
| `""`                       | controller object | `Always`            |
| `always`                   | controller object | `Always`            |
| `on-failure`               | Pod               | `OnFailure`         |
| `no`                       | Pod               | `Never`             |

**Note**: controller object could be `deployment` or `replicationcontroller`, etc.

For e.g. `pival` service will become pod down here. This container calculated value of `pi`.

```yaml
version: '2'

services:
  pival:
    image: perl
    command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
    restart: "on-failure"
```
