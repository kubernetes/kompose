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

INFO[0000] Successfully created Service: redis          
INFO[0000] Successfully created Service: web            
INFO[0000] Successfully created Deployment: redis       
INFO[0000] Successfully created Deployment: web         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.
```

__Alternatively, you can run `kompose convert` and deploy with `kubectl`__

__2.1. Run `kompose convert` in the same directory__

```bash
▶ kompose convert                           
INFO[0000] file "frontend-service.json" created         
INFO[0000] file "redis-master-service.json" created     
INFO[0000] file "redis-slave-service.json" created      
INFO[0000] file "frontend-deployment.json" created      
INFO[0000] file "redis-master-deployment.json" created  
INFO[0000] file "redis-slave-deployment.json" created   
```

__2.2. And start it on Kubernetes!__

```bash
▶ kubectl create -f frontend-service.json,redis-master-service.json,redis-slave-service.json,frontend-deployment.json,redis-master-deployment.json,redis-slave-deployment.json
service "frontend" created
service "redis-master" created
service "redis-slave" created
deployment "frontend" created
deployment "redis-master" created
deployment "redis-slave" created
```
