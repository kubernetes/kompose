# User Guide

* TOC
{:toc}

Kompose has support for two providers: OpenShift and Kubernetes.
You can choose a targeted provider using global option `--provider`. If no provider is specified, Kubernetes is set by default.


## Kompose Convert

Kompose supports conversion of V1, V2, and V3 Docker Compose files into Kubernetes and OpenShift objects.

### Kubernetes

```sh
$ kompose --file docker-voting.yml convert
WARN Unsupported key networks - ignoring
WARN Unsupported key build - ignoring
INFO Kubernetes file "worker-svc.yaml" created
INFO Kubernetes file "db-svc.yaml" created
INFO Kubernetes file "redis-svc.yaml" created
INFO Kubernetes file "result-svc.yaml" created
INFO Kubernetes file "vote-svc.yaml" created
INFO Kubernetes file "redis-deployment.yaml" created
INFO Kubernetes file "result-deployment.yaml" created
INFO Kubernetes file "vote-deployment.yaml" created
INFO Kubernetes file "worker-deployment.yaml" created
INFO Kubernetes file "db-deployment.yaml" created

$ ls
db-deployment.yaml  docker-compose.yml         docker-gitlab.yml  redis-deployment.yaml  result-deployment.yaml  vote-deployment.yaml  worker-deployment.yaml
db-svc.yaml         docker-voting.yml          redis-svc.yaml     result-svc.yaml        vote-svc.yaml           worker-svc.yaml
```

You can also provide multiple docker-compose files at the same time:

```sh
$ kompose -f docker-compose.yml -f docker-guestbook.yml convert
INFO Kubernetes file "frontend-service.yaml" created         
INFO Kubernetes file "mlbparks-service.yaml" created         
INFO Kubernetes file "mongodb-service.yaml" created          
INFO Kubernetes file "redis-master-service.yaml" created     
INFO Kubernetes file "redis-slave-service.yaml" created      
INFO Kubernetes file "frontend-deployment.yaml" created      
INFO Kubernetes file "mlbparks-deployment.yaml" created      
INFO Kubernetes file "mongodb-deployment.yaml" created       
INFO Kubernetes file "mongodb-claim0-persistentvolumeclaim.yaml" created 
INFO Kubernetes file "redis-master-deployment.yaml" created  
INFO Kubernetes file "redis-slave-deployment.yaml" created   

$ ls
mlbparks-deployment.yaml  mongodb-service.yaml                       redis-slave-service.jsonmlbparks-service.yaml  
frontend-deployment.yaml  mongodb-claim0-persistentvolumeclaim.yaml  redis-master-service.yaml
frontend-service.yaml     mongodb-deployment.yaml                    redis-slave-deployment.yaml
redis-master-deployment.yaml
``` 

When multiple docker-compose files are provided the configuration is merged. Any configuration that is common will be over ridden by subsequent file.
 
### OpenShift

```sh
$ kompose --provider openshift --file docker-voting.yml convert
WARN [worker] Service cannot be created because of missing port.
INFO OpenShift file "vote-service.yaml" created             
INFO OpenShift file "db-service.yaml" created               
INFO OpenShift file "redis-service.yaml" created            
INFO OpenShift file "result-service.yaml" created           
INFO OpenShift file "vote-deploymentconfig.yaml" created    
INFO OpenShift file "vote-imagestream.yaml" created         
INFO OpenShift file "worker-deploymentconfig.yaml" created  
INFO OpenShift file "worker-imagestream.yaml" created       
INFO OpenShift file "db-deploymentconfig.yaml" created      
INFO OpenShift file "db-imagestream.yaml" created           
INFO OpenShift file "redis-deploymentconfig.yaml" created   
INFO OpenShift file "redis-imagestream.yaml" created        
INFO OpenShift file "result-deploymentconfig.yaml" created  
INFO OpenShift file "result-imagestream.yaml" created  
```

It also supports creating buildconfig for build directive in a service. By default, it uses the remote repo for the current git branch as the source repo, and the current branch as the source branch for the build. You can specify a different source repo and branch using ``--build-repo`` and ``--build-branch`` options respectively.

```sh
$ kompose --provider openshift --file buildconfig/docker-compose.yml convert
WARN [foo] Service cannot be created because of missing port. 
INFO OpenShift Buildconfig using git@github.com:rtnpro/kompose.git::master as source. 
INFO OpenShift file "foo-deploymentconfig.yaml" created     
INFO OpenShift file "foo-imagestream.yaml" created          
INFO OpenShift file "foo-buildconfig.yaml" created 
```

**Note**: If you are manually pushing the Openshift artifacts using ``oc create -f``, you need to ensure that you push the imagestream artifact before the buildconfig artifact, to workaround this Openshift issue: https://github.com/openshift/origin/issues/4518 .

## Kompose Up

Kompose supports a straightforward way to deploy your "composed" application to Kubernetes or OpenShift via `kompose up`.


### Kubernetes
```sh
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
NAME                               DESIRED       CURRENT       UP-TO-DATE   AVAILABLE   AGE
deploy/frontend                    1             1             1            1           4m
deploy/redis-master                1             1             1            1           4m
deploy/redis-slave                 1             1             1            1           4m

NAME                               CLUSTER-IP    EXTERNAL-IP   PORT(S)      AGE
svc/frontend                       10.0.174.12   <none>        80/TCP       4m
svc/kubernetes                     10.0.0.1      <none>        443/TCP      13d
svc/redis-master                   10.0.202.43   <none>        6379/TCP     4m
svc/redis-slave                    10.0.1.85     <none>        6379/TCP     4m

NAME                               READY         STATUS        RESTARTS     AGE
po/frontend-2768218532-cs5t5       1/1           Running       0            4m
po/redis-master-1432129712-63jn8   1/1           Running       0            4m
po/redis-slave-2504961300-nve7b    1/1           Running       0            4m
```
Note:
- You must have a running Kubernetes cluster with a pre-configured kubectl context.
- If you kubernetes api server url is different than `https://127.0.0.1:6443`, you can use the `--server` flag to customize it.
- Only deployments and services are generated and deployed to Kubernetes. If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead.

### OpenShift
```sh
$ kompose --file ./examples/docker-guestbook.yml --provider openshift up
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

## Kompose Down

Once you have deployed "composed" application to Kubernetes, `$ kompose down` will help you to take the application out by deleting its deployments and services. If you need to remove other resources, use the 'kubectl' command.

```sh
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

## Build and Push Docker Images

Kompose supports both building and pushing Docker images. When using the `build` key within your Docker Compose file, your image will:

  - Automatically be built with Docker using the `image` key specified within your file
  - Be pushed to the correct Docker repository using local credentials (located at `.docker/config`)
  
If you do not want to push the image you have built, you and use the `--push-image=false` flag to disable it.

Using an [example Docker Compose file](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/buildconfig/docker-compose.yml):

```yaml
version: "2"

services:
    foo:
        build: "./build"
        image: docker.io/foo/bar
```

Using `kompose up` with a `build` key:

```sh
$ kompose up
INFO Build key detected. Attempting to build and push image 'docker.io/foo/bar' 
INFO Building image 'docker.io/foo/bar' from directory 'build' 
INFO Image 'docker.io/foo/bar' from directory 'build' built successfully 
INFO Pushing image 'foo/bar:latest' to registry 'docker.io' 
INFO Attempting authentication credentials 'https://index.docker.io/v1/ 
INFO Successfully pushed image 'foo/bar:latest' to registry 'docker.io' 
INFO We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. 
 
INFO Deploying application in "default" namespace 
INFO Successfully created Service: foo            
INFO Successfully created Deployment: foo         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.
```

In order to disable the functionality, or choose to use BuildConfig generation (with OpenShift) `--build (local|build-config|none)` can be passed.

```sh
# Disable building/pushing Docker images
$ kompose up --build none

# Generate Build Config artifacts for OpenShift
$ kompose up --provider openshift --build build-config
```

## Alternative Conversions

The default `kompose` transformation will generate Kubernetes [Deployments](http://kubernetes.io/docs/user-guide/deployments/) and [Services](http://kubernetes.io/docs/user-guide/services/), in yaml format. You have alternative option to generate json with `-j`. Also, you can alternatively generate [Replication Controllers](http://kubernetes.io/docs/user-guide/replication-controller/) objects, [Daemon Sets](http://kubernetes.io/docs/admin/daemons/), or [Helm](https://github.com/helm/helm) charts.

```sh
$ kompose convert -j
INFO Kubernetes file "redis-svc.json" created
INFO Kubernetes file "web-svc.json" created
INFO Kubernetes file "redis-deployment.json" created
INFO Kubernetes file "web-deployment.json" created
```
The `*-deployment.json` files contain the Deployment objects.

```sh
$ kompose convert --controller replicationController
INFO Kubernetes file "redis-svc.yaml" created
INFO Kubernetes file "web-svc.yaml" created
INFO Kubernetes file "redis-replicationcontroller.yaml" created
INFO Kubernetes file "web-replicationcontroller.yaml" created
```

The `*-replicationcontroller.yaml` files contain the Replication Controller objects. If you want to specify replicas (default is 1), use `--replicas` flag: `$ kompose convert --controller replicationController --replicas 3`

```sh
$ kompose convert --controller daemonSet
INFO Kubernetes file "redis-svc.yaml" created
INFO Kubernetes file "web-svc.yaml" created
INFO Kubernetes file "redis-daemonset.yaml" created
INFO Kubernetes file "web-daemonset.yaml" created
```

The `*-daemonset.yaml` files contain the Daemon Set objects

If you want to generate a Chart to be used with [Helm](https://github.com/kubernetes/helm) simply do:

```sh
$ kompose convert -c 
INFO Kubernetes file "web-svc.yaml" created
INFO Kubernetes file "redis-svc.yaml" created
INFO Kubernetes file "web-deployment.yaml" created
INFO Kubernetes file "redis-deployment.yaml" created
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

## Labels

`kompose` supports Kompose-specific labels within the `docker-compose.yml` file to
explicitly define the generated resources' behavior upon conversion, like Service, PersistentVolumeClaim...

The currently supported options are:

| Key                  | Value                               |
|----------------------|-------------------------------------|
| kompose.service.type | nodeport / clusterip / loadbalancer / headless |
| kompose.service.expose | true / hostnames (separated by comma) |
| kompose.service.nodeport.port | port value (string) | 
| kompose.service.expose.tls-secret | secret name |
| kompose.volume.size | kubernetes supported volume size |
| kompose.controller.type | deployment / daemonset / replicationcontroller |
| kompose.image-pull-policy | kubernetes pods imagePullPolicy |
| kompose.image-pull-secret | kubernetes secret name for imagePullSecrets |

**Note**: `kompose.service.type` label should be defined with `ports` only (except for headless service), otherwise `kompose` will fail.



- `kompose.service.type` defines the type of service to be created.

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

- `kompose.service.expose` defines if the service needs to be made accessible from outside the cluster or not. If the value is set to "true", the provider sets the endpoint automatically, and for any other value, the value is set as the hostname. If multiple ports are defined in a service, the first one is chosen to be the exposed.
    - For the Kubernetes provider, an ingress resource is created and it is assumed that an ingress controller has already been configured. If the value is set to a comma sepatated list, multiple hostnames are supported.Hostname with path is also supported.
    - For the OpenShift provider, a route is created.
- `kompose.service.nodeport.port` defines the port value when service type is `nodeport`, this label should only be set when the service only contains 1 port. Usually kubernetes define a port range for node port values, kompose will not validate this.
- `kompose.service.expose.tls-secret` provides the name of the TLS secret to use with the Kubernetes ingress controller. This requires kompose.service.expose to be set.

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
      kompose.service.expose: "counter.example.com,foobar.example.com"
      kompose.service.expose.tls-secret: "example-secret"
  redis:
    image: redis:3.0
    ports:
     - "6379"
```

- `kompose.image-pull-secret` defines a kubernetes secret name for imagePullSecrets podspec field.
This secret will be used for pulling private images.
For example:

```yaml
version: '2'
services:
  tm-service:
    image: premium/private-image
    labels:
      kompose.image-pull-secret: "example-kubernetes-secret"
```

- `kompose.volume.size` defines the requests storage's size in the PersistentVolumeClaim

For example:

```yaml
version: '2'
services:
  db:
    image: postgres:10.1
    labels:
      kompose.volume.size: 1Gi
    volumes:
      - db-data:/var/lib/postgresql/data
```

- `kompose.controller.type` defines which controller type should convert for this service

For example:

```
web:
  image: wordpress:4.5
  ports:
    - '80'
  environment:
    WORDPRESS_AUTH_KEY: changeme
    WORDPRESS_SECURE_AUTH_KEY: changeme
    WORDPRESS_LOGGED_IN_KEY: changeme
    WORDPRESS_NONCE_KEY: changeme
    WORDPRESS_AUTH_SALT: changeme
    WORDPRESS_SECURE_AUTH_SALT: changeme
    WORDPRESS_LOGGED_IN_SALT: changeme
    WORDPRESS_NONCE_SALT: changeme
    WORDPRESS_NONCE_AA: changeme
  restart: always
  links:
    - 'db:mysql'
db:
  image: mysql:5.7
  environment:
    MYSQL_ROOT_PASSWORD: password
  restart: always
  labels:
    project.logs: /var/log/mysql
    kompose.controller.type: daemonset
```

Service `web` will be converted to `Deployment` as default, service `db` will be converted to `DaemonSet` because of `kompose.controller.type` label.

- `kompose.image-pull-policy` defines Kubernetes PodSpec imagePullPolicy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.

For example:

```yaml
version: '2'
services:
  example-service:
    image: example-image
    labels:
      kompose.image-pull-policy: "Never"
```

## Restart

If you want to create normal pods without controller you can use `restart` construct of docker-compose to define that. Follow table below to see what happens on the `restart` value.

| `docker-compose` `restart` | object created    | Pod `restartPolicy` |
|----------------------------|-------------------|---------------------|
| `""`                       | controller object | `Always`            |
| `always`                   | controller object | `Always`            |
| `unless-stopped`           | controller object | `Always`            |
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

#### Warning about Deployment Config's

If the Docker Compose file has a volume specified for a service, the Deployment (Kubernetes) or DeploymentConfig (OpenShift) strategy is changed to "Recreate" instead of "RollingUpdate" (default). This is done to avoid multiple instances of a service from accessing a volume at the same time.

If the Docker Compose file has service name with `_` or `.` in it (eg.`web_service` or `web.service`), then it will be replaced by `-` and the service name will be renamed accordingly (eg.`web-service`). Kompose does this because "Kubernetes" doesn't allow `_` in object name.

Please note that changing service name might break some `docker-compose` files.

## Docker Compose Versions

Kompose supports Docker Compose versions: 1, 2 and 3. We have limited support on versions 2.1 and 3.2 due to their experimental nature.

A full list on compatibility between all three versions is listed in our [conversion document](/docs/conversion.md) including a list of all incompatible Docker Compose keys.
