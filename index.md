---
layout: default
---

# Kubernetes + Compose = Kompose

What's Kompose? It's a conversion tool for all things compose (namely Docker Compose) to container orchestrators (Kubernetes or OpenShift).

Whether you have a `docker-compose.yaml` or a `docker-compose.dab`, it doesn't matter. Kompose will get you up-and-running on Kubernetes.

In two simple steps, we'll take you from Docker Compose to Kubernetes.

__1. Take a sample docker-compose.yaml file__

```yaml
version: "2"

services:

  redis-master:
    image: gcr.io/google_containers/redis:e2e 
    ports:
      - "6379"

  redis-slave:
    image: gcr.io/google_samples/gb-redisslave:v1
    ports:
      - "6379"
    environment:
      - GET_HOSTS_FROM=dns

  frontend:
    image: gcr.io/google-samples/gb-frontend:v4
    ports:
      - "80:80"
    environment:
      - GET_HOSTS_FROM=dns
```

__2. Run `kompose up` in the same directory__ 

```bash
▶ kompose up
We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. 

INFO Successfully created Service: redis          
INFO Successfully created Service: web            
INFO Successfully created Deployment: redis       
INFO Successfully created Deployment: web         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.
```

__Alternatively, you can run `kompose convert` and deploy with `kubectl`__

__2.1. Run `kompose convert` in the same directory__

```bash
▶ kompose convert                           
INFO file "frontend-service.yaml" created         
INFO file "redis-master-service.yaml" created     
INFO file "redis-slave-service.yaml" created      
INFO file "frontend-deployment.yaml" created      
INFO file "redis-master-deployment.yaml" created  
INFO file "redis-slave-deployment.yaml" created   
```

__2.2. And start it on Kubernetes!__

```bash
▶ kubectl create -f frontend-service.yaml,redis-master-service.yaml,redis-slave-service.yaml,frontend-deployment.yaml,redis-master-deployment.yaml,redis-slave-deployment.yaml
service "frontend" created
service "redis-master" created
service "redis-slave" created
deployment "frontend" created
deployment "redis-master" created
deployment "redis-slave" created
```
