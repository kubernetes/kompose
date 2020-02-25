# Kubernetes + Compose = Kompose
## A conversion tool to go from Docker Compose to Kubernetes

### What's Kompose?

Kompose is a conversion tool for Docker Compose to container orchestrators such as Kubernetes (or OpenShift).

Why do developers love it?

  - Simplify your development process with Docker Compose and then deploy your containers to a production cluster
  - Convert your `docker-compose.yaml` with one simple command `kompose convert`
  - Immediately bring up your cluster with `kompose up`
  - Bring it back down with `kompose down`

### It's as simple as 1-2-3

1. [Use an example docker-compose.yaml file](https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose-v3.yaml) or your own
2. Run `kompose up`
3. Check your Kubernetes cluster for your newly deployed containers!

```sh
$ wget https://raw.githubusercontent.com/kubernetes/kompose/master/examples/docker-compose-v3.yaml -O docker-compose.yaml

$ kompose up
We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. 

INFO Successfully created Service: redis          
INFO Successfully created Service: web            
INFO Successfully created Deployment: redis       
INFO Successfully created Deployment: web         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.

$ kubectl get po
NAME                            READY     STATUS              RESTARTS   AGE
frontend-591253677-5t038        1/1       Running             0          10s
redis-master-2410703502-9hshf   1/1       Running             0          10s
redis-slave-4049176185-hr1lr    1/1       Running             0          10s
```

A more detailed guide is available in our [getting started guide](/docs/getting-started.md).

### Install Kompose on Linux, macOS or Windows

Grab the Kompose binary!

```sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

_Windows:_ Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-windows-amd64.exe) and add the binary to your PATH.
