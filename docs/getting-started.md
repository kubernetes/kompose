# Getting Started

* TOC
{:toc}

This is how you'll get started with Kompose!

There are three different guides depending on your container orchestrator as well as operating system.

For beginners and the most compatibility, follow the _Minikube and Kompose_ guide.

## Minikube and Kompose

In this guide, we'll deploy a sample `docker-compose.yaml` file to a Kubernetes cluster.

Requirements:
  - [minikube](https://github.com/kubernetes/minikube)
  - [kompose](https://github.com/kubernetes/kompose)

__Start `minikube`:__

If you don't already have a Kubernetes cluster running, [minikube](https://github.com/kubernetes/minikube) is the best way to get started.

```sh
$ minikube start
Starting local Kubernetes v1.7.5 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster
```

__Download an [example Docker Compose file](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml), or use your own:__

```sh
wget https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml
```

__Convert your Docker Compose file to Kubernetes:__

Run `kompose convert` in the same directory as your `docker-compose.yaml` file.

```sh
$ kompose convert                           
INFO Kubernetes file "frontend-service.yaml" created         
INFO Kubernetes file "redis-master-service.yaml" created     
INFO Kubernetes file "redis-slave-service.yaml" created      
INFO Kubernetes file "frontend-deployment.yaml" created      
INFO Kubernetes file "redis-master-deployment.yaml" created  
INFO Kubernetes file "redis-slave-deployment.yaml" created 
```

Alternatively, you can convert and deploy directly to Kubernetes with `kompose up`.

```sh
$ kompose up
We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. 

INFO Successfully created Service: redis          
INFO Successfully created Service: web            
INFO Successfully created Deployment: redis       
INFO Successfully created Deployment: web         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.
```


__Access the newly deployed service:__

Now that your service has been deployed, let's access it.

If you're using `minikube` you may access it via the `minikube service` command.

```sh
$ minikube service frontend
```

Otherwise, use `kubectl` to see what IP the service is using:

```sh
$ kubectl describe svc frontend
Name:                   frontend
Namespace:              default
Labels:                 service=frontend
Selector:               service=frontend
Type:                   LoadBalancer
IP:                     10.0.0.183
LoadBalancer Ingress:   123.45.67.89
Port:                   80      80/TCP
NodePort:               80      31144/TCP
Endpoints:              172.17.0.4:80
Session Affinity:       None
No events.

```

Note: If you're using a cloud provider, your IP will be listed next to `LoadBalancer Ingress`.

If you have yet to expose your service (for example, within GCE), use the command:

```sh
kubectl expose deployment frontend --type="LoadBalancer" 
```

To check functionality, you may also `curl` the URL.

```sh
$ curl http://123.45.67.89
```

## Minishift and Kompose

In this guide, we'll deploy a sample `docker-compose.yaml` file to an OpenShift cluster.

Requirements:
  - [minishift](https://github.com/minishift/minishift)
  - [kompose](https://github.com/kubernetes/kompose)
  - An OpenShift route created

__Note:__ The service will NOT be accessible until you create an OpenShift route with `oc expose`. You must also have a virtualization environment setup. By default, `minishift` uses KVM.

__Start `minishift`:__

[Minishift](https://github.com/minishift/minishift) is a tool that helps run OpenShift locally using a single-node cluster inside of a VM. Similar to [minikube](https://github.com/kubernetes/minikube).

```sh
$ minishift start
Starting local OpenShift cluster using 'kvm' hypervisor...
-- Checking OpenShift client ... OK
-- Checking Docker client ... OK
-- Checking Docker version ... OK
-- Checking for existing OpenShift container ... OK
...
```

__Download an [example Docker Compose file](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml), or use your own:__

```sh
wget https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml
```

__Convert your Docker Compose file to OpenShift:__

Run `kompose convert --provider=openshift` in the same directory as your `docker-compose.yaml` file.

```sh
$ kompose convert --provider=openshift
INFO OpenShift file "frontend-service.yaml" created 
INFO OpenShift file "redis-master-service.yaml" created 
INFO OpenShift file "redis-slave-service.yaml" created 
INFO OpenShift file "frontend-deploymentconfig.yaml" created 
INFO OpenShift file "frontend-imagestream.yaml" created 
INFO OpenShift file "redis-master-deploymentconfig.yaml" created 
INFO OpenShift file "redis-master-imagestream.yaml" created 
INFO OpenShift file "redis-slave-deploymentconfig.yaml" created 
INFO OpenShift file "redis-slave-imagestream.yaml" created 
```

Alternatively, you can convert and deploy directly to OpenShift with `kompose up --provider=openshift`.

```sh
$ kompose up --provider=openshift       
INFO We are going to create OpenShift DeploymentConfigs, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead. 
 
INFO Deploying application in "myproject" namespace 
INFO Successfully created Service: frontend       
INFO Successfully created Service: redis-master   
INFO Successfully created Service: redis-slave    
INFO Successfully created DeploymentConfig: frontend 
INFO Successfully created ImageStream: frontend   
INFO Successfully created DeploymentConfig: redis-master 
INFO Successfully created ImageStream: redis-master 
INFO Successfully created DeploymentConfig: redis-slave 
INFO Successfully created ImageStream: redis-slave 

Your application has been deployed to OpenShift. You can run 'oc get dc,svc,is,pvc' for details.
```

__Access the newly deployed service:__

After deployment, you must create an OpenShift route in order to access the service.

If you're using `minishift`, you'll use a combination of `oc` and `minishift` commands to access the service.

Create a route for the `frontend` service using `oc`:

```sh
$ oc expose service/frontend
route "frontend" exposed
```

Access the `frontend` service with `minishift`:

```sh
$ minishift openshift service frontend --namespace=myproject
Opening the service myproject/frontend in the default browser...
```

You can also access the GUI interface of OpenShift for an overview of the deployed containers:

```sh
$ minishift console
Opening the OpenShift Web console in the default browser...
```

## RHEL and Kompose

In this guide, we'll deploy a sample `docker-compose.yaml` file using both RHEL (Red Hat Enterprise Linux) and OpenShift.

Requirements:
  - Red Hat Enterprise Linux 7.4
  - [Red Hat Development Suite](https://developers.redhat.com/products/devsuite/overview/)
    - Which includes:
    - [minishift](https://github.com/minishift/minishift)
    - [kompose](https://github.com/kubernetes/kompose)

__Note:__ A KVM hypervisor must be setup in order to correctly use `minishift` on RHEL. You can set it up via the [CDK Documentation](https://access.redhat.com/documentation/en-us/red_hat_container_development_kit/3.1/html-single/getting_started_guide/index#setup-virtualization) under "Set up your virtualization environment".

__Install Red Hat Development Suite:__

Before we are able to use both `minishift` and `kompose`, DevSuite must be installed. A more concise [installation document is available](https://developers.redhat.com/products/devsuite/hello-world/#fndtn-rhel).

Change to root.

```sh
$ su -
```

Enable the Red Hat Developer Tools software repository.

```sh
$ subscription-manager repos --enable rhel-7-server-devtools-rpms
$ subscription-manager repos --enable rhel-server-rhscl-7-rpms
```

Add the Red Hat Developer Tools key to your system.

```sh
$ cd /etc/pki/rpm-gpg
$ wget -O RPM-GPG-KEY-redhat-devel https://www.redhat.com/security/data/a5787476.txt
$ rpm --import RPM-GPG-KEY-redhat-devel
```

Install Red Hat Development Suite and Kompose.

```sh
$ yum install rh-devsuite kompose -y
```

__Start `minishift`:__

Before we begin, we must do a few preliminary steps setting up `minishift`.

```sh
$ su -
$ ln -s /var/lib/cdk-minishift-3.0.0/minishift /usr/bin/minishift
$ minishift setup-cdk --force --default-vm-driver="kvm"
$ ln -s /home/$(whoami)/.minishift/cache/oc/v3.5.5.8/oc /usr/bin/oc
```

Now we may start `minishift`.

```sh
$ minishift start
Starting local OpenShift cluster using 'kvm' hypervisor...
-- Checking OpenShift client ... OK
-- Checking Docker client ... OK
-- Checking Docker version ... OK
-- Checking for existing OpenShift container ... OK
...
```

__Download an [example Docker Compose file](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml), or use your own:__

```sh
wget https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose.yaml
```

__Convert your Docker Compose file to OpenShift:__

Run `kompose convert --provider=openshift` in the same directory as your `docker-compose.yaml` file.

```sh
$ kompose convert --provider=openshift
INFO OpenShift file "frontend-service.yaml" created 
INFO OpenShift file "redis-master-service.yaml" created 
INFO OpenShift file "redis-slave-service.yaml" created 
INFO OpenShift file "frontend-deploymentconfig.yaml" created 
INFO OpenShift file "frontend-imagestream.yaml" created 
INFO OpenShift file "redis-master-deploymentconfig.yaml" created 
INFO OpenShift file "redis-master-imagestream.yaml" created 
INFO OpenShift file "redis-slave-deploymentconfig.yaml" created 
INFO OpenShift file "redis-slave-imagestream.yaml" created 
```

Alternatively, you can convert and deploy directly to OpenShift with `kompose up --provider=openshift`.

```sh
$ kompose up --provider=openshift       
INFO We are going to create OpenShift DeploymentConfigs, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead. 
 
INFO Deploying application in "myproject" namespace 
INFO Successfully created Service: frontend       
INFO Successfully created Service: redis-master   
INFO Successfully created Service: redis-slave    
INFO Successfully created DeploymentConfig: frontend 
INFO Successfully created ImageStream: frontend   
INFO Successfully created DeploymentConfig: redis-master 
INFO Successfully created ImageStream: redis-master 
INFO Successfully created DeploymentConfig: redis-slave 
INFO Successfully created ImageStream: redis-slave 

Your application has been deployed to OpenShift. You can run 'oc get dc,svc,is,pvc' for details.
```

__Access the newly deployed service:__

After deployment, you must create an OpenShift route in order to access the service.

If you're using `minishift`, you'll use a combination of `oc` and `minishift` commands to access the service.

Create a route for the `frontend` service using `oc`:

```sh
$ oc expose service/frontend
route "frontend" exposed
```

Access the `frontend` service with `minishift`:

```sh
$ minishift openshift service frontend --namespace=myproject
Opening the service myproject/frontend in the default browser...
```

You can also access the GUI interface of OpenShift for an overview of the deployed containers:

```sh
$ minishift console
Opening the OpenShift Web console in the default browser...
```
