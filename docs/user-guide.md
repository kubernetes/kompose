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

$ kompose -f docker-gitlab.yml convert -y
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
db-svc.json         docker-compose-bundle.dab  docker-voting.yml  redis-svc.json         result-svc.json         vote-svc.json         worker-svc.json
```

Using `--bundle, --dab` to specify a DAB file as below:

```console
$ kompose --bundle docker-compose-bundle.dab convert
WARN[0000]: Unsupported key networks - ignoring
file "redis-svc.json" created
file "web-svc.json" created
file "web-deployment.json" created
file "redis-deployment.json" created
```

### OpenShift

```console
$ kompose --provider openshift --file docker-voting.yml convert
WARN[0000] [worker] Service cannot be created because of missing port.
INFO[0000] file "vote-service.json" created             
INFO[0000] file "db-service.json" created               
INFO[0000] file "redis-service.json" created            
INFO[0000] file "result-service.json" created           
INFO[0000] file "vote-deploymentconfig.json" created    
INFO[0000] file "vote-imagestream.json" created         
INFO[0000] file "worker-deploymentconfig.json" created  
INFO[0000] file "worker-imagestream.json" created       
INFO[0000] file "db-deploymentconfig.json" created      
INFO[0000] file "db-imagestream.json" created           
INFO[0000] file "redis-deploymentconfig.json" created   
INFO[0000] file "redis-imagestream.json" created        
INFO[0000] file "result-deploymentconfig.json" created  
INFO[0000] file "result-imagestream.json" created  
```

In similar way you can convert DAB files to OpenShift.
```console$
$ kompose --bundle docker-compose-bundle.dab --provider openshift convert
WARN[0000]: Unsupported key networks - ignoring
INFO[0000] file "redis-svc.json" created
INFO[0000] file "web-svc.json" created
INFO[0000] file "web-deploymentconfig.json" created
INFO[0000] file "web-imagestream.json" created           
INFO[0000] file "redis-deploymentconfig.json" created
INFO[0000] file "redis-imagestream.json" created
```

## Kompose up

Kompose supports a straightforward way to deploy your "composed" application to Kubernetes or OpenShift via `kompose up`.


### Kubernetes
```console
$ kompose --file ./examples/docker-guestbook.yml up
We are going to create Kubernetes deployments and services for your Dockerized application.
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead.

INFO[0000] Successfully created service: redis-master   
INFO[0000] Successfully created service: redis-slave    
INFO[0000] Successfully created service: frontend       
INFO[0001] Successfully created deployment: redis-master
INFO[0001] Successfully created deployment: redis-slave
INFO[0001] Successfully created deployment: frontend    

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

INFO[0000] Successfully created service: redis-slave    
INFO[0000] Successfully created service: frontend       
INFO[0000] Successfully created service: redis-master   
INFO[0000] Successfully created deployment: redis-slave
INFO[0000] Successfully created ImageStream: redis-slave
INFO[0000] Successfully created deployment: frontend    
INFO[0000] Successfully created ImageStream: frontend   
INFO[0000] Successfully created deployment: redis-master
INFO[0000] Successfully created ImageStream: redis-master

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
INFO[0000] Successfully deleted service: redis-master   
INFO[0004] Successfully deleted deployment: redis-master
INFO[0004] Successfully deleted service: redis-slave    
INFO[0008] Successfully deleted deployment: redis-slave
INFO[0009] Successfully deleted service: frontend       
INFO[0013] Successfully deleted deployment: frontend
```
Note:
- You must have a running Kubernetes cluster with a pre-configured kubectl context.

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
WARN[0000] Unsupported key build - ignoring             
WARN[0000] Unsupported key cap_add - ignoring           
WARN[0000] Unsupported key dockerfile - ignoring
```

## Labels

`kompose` supports Kompose-specific labels within the `docker-compose.yml` file in order to explicitly imply a service type upon conversion.

The currently supported options are:

| Key                  | Value                               |
|----------------------|-------------------------------------|
| kompose.service.type | nodeport / clusterip / loadbalancer |


For example:

```yaml
--- 
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
