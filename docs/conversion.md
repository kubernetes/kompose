---
layout: default
title: Conversion
permalink: /conversion/
redirect_from: 
  - /docs/conversion.md/
  - /docs/conversion/
---

# Conversion Matrix

* TOC
{:toc}

This document outlines all possible conversion details regarding `docker-compose.yaml` values to Kubernetes / OpenShift artifacts.

## Version Support

Under the hood, we're using [compose-go](https://github.com/compose-spec/compose-go), the reference library for parsing Compose files. We should be able to load all versions of Compose files.
We're doing our best to keep it up to date as soon as possible in our releases to be compatible with the latest features defined in the [Compose specification](https://github.com/compose-spec/compose-spec/blob/master/spec.md). If you absolutely need a feature we don't support yet, please open a PR!

## Conversion Table

**Glossary:**

- **✓:** Converts
- **-:** Not in this Docker Compose Version
- **n:** Not yet implemented
- **x:** Not applicable / no 1-1 conversion

| Keys                   | V1 | V2 | V3 | Kubernetes / OpenShift                                               | Notes                                                                                                                             |
| ---------------------- | -- | -- | -- | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| build                  | ✓  | ✓  | ✓  |                                                                      | Builds/Pushes to Docker repository. See [user guide on build and push image](https://kompose.io/user-guide/#build-and-push-image) |
| build: context         | ✓  | ✓  | ✓  |                                                                      |                                                                                                                                   |
| build: dockerfile      | ✓  | ✓  | ✓  |                                                                      |                                                                                                                                   |
| build: args            | n  | n  | n  |                                                                      |                                                                                                                                   |
| build: cache_from      | -  | -  | n  |                                                                      |                                                                                                                                   |
| cap_add                | ✓  | ✓  | ✓  | Container.SecurityContext.Capabilities.Add                           |                                                                                                                                   |
| cap_drop               | ✓  | ✓  | ✓  | Container.SecurityContext.Capabilities.Drop                          |                                                                                                                                   |
| command                | ✓  | ✓  | ✓  | Container.Args                                                       |                                                                                                                                   |
| configs                | n  | n  | ✓  |                                                                      |                                                                                                                                   |
| configs: short-syntax  | n  | n  | ✓  |                                                                      | Only create configMap                                                                                                             |
| configs: long-syntax   | n  | n  | ✓  |                                                                      | If target path is /, ignore this and only create configMap                                                                        |
| cgroup_parent          | x  | x  | x  |                                                                      | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/11986                                  |
| container_name         | ✓  | ✓  | ✓  | Metadata.Name + Deployment.Spec.Containers.Name                      |                                                                                                                                   |
| credential_spec        | x  | x  | x  |                                                                      | Only applicable to Windows containers                                                                                             |
| deploy                 | -  | -  | ✓  |                                                                      |                                                                                                                                   |
| deploy: mode           | -  | -  | ✓  |                                                                      |                                                                                                                                   |
| deploy: replicas       | -  | -  | ✓  | Deployment.Spec.Replicas / DeploymentConfig.Spec.Replicas            |                                                                                                                                   |
| deploy: placement      | -  | -  | ✓  | Affinity                                                             |                                                                                                                                   |
| deploy: update_config  | -  | -  | ✓  | Workload.Spec.Strategy                                               | Deployment / DeploymentConfig                                                                                                     |
| deploy: resources      | -  | -  | ✓  | Containers.Resources.Limits.Memory / Containers.Resources.Limits.CPU | Support for memory as well as cpu                                                                                                 |
| deploy: restart_policy | -  | -  | ✓  | Pod generation                                                       | This generated a Pod, see the [user guide on restart](http://kompose.io/user-guide/#restart)                                      |
| deploy: labels         | -  | -  | ✓  | Workload.Metadata.Labels                                             | Only applied to workload resource                                                                                                 |
| devices                | x  | x  | x  |                                                                      | Not supported within Kubernetes, See issue https://github.com/kubernetes/kubernetes/issues/5607                                   |
| depends_on             | x  | x  | x  |                                                                      |                                                                                                                                   |
| dns                    | x  | x  | x  |                                                                      | Not used within Kubernetes. Kubernetes uses a managed DNS server                                                                  |
| dns_search             | x  | x  | x  |                                                                      | See `dns` key                                                                                                                     |
| domainname             | ✓  | ✓  | ✓  | SubDomain                                                            |                                                                                                                                   |
| tmpfs                  | ✓  | ✓  | ✓  | Containers.Volumes.EmptyDir                                          | Creates emptyDirvolume with medium set to Memory & mounts given directory inside container                                        |
| entrypoint             | ✓  | ✓  | ✓  | Container.Command                                                    |                                                                                                                                   |
| env_file               | n  | n  | ✓  |                                                                      |                                                                                                                                   |
| environment            | ✓  | ✓  | ✓  | Container.Env                                                        |                                                                                                                                   |
| expose                 | ✓  | ✓  | ✓  | Service.Spec.Ports                                                   |                                                                                                                                   |
| endpoint_mode          | n  | n  | ✓  |                                                                      | If endpoint_mode=vip, the created Service will be forced to set to NodePort type                                                  |
| extends                | ✓  | ✓  | ✓  |                                                                      | Extends by utilizing the same image supplied                                                                                      |
| external_links         | x  | x  | x  |                                                                      | Kubernetes uses a flat-structure for all containers and thus external_links does not have a 1-1 conversion                        |
| extra_hosts            | n  | n  | n  |                                                                      |                                                                                                                                   |
| group_add              | ✓  | ✓  | ✓  |                                                                      |                                                                                                                                   |
| healthcheck            | -  | n  | ✓  |                                                                      |                                                                                                                                   |
| hostname               | ✓  | ✓  | ✓  | HostName                                                             |                                                                                                                                   |
| image                  | ✓  | ✓  | ✓  | Deployment.Spec.Containers.Image                                     |                                                                                                                                   |
| isolation              | x  | x  | x  |                                                                      | Not applicable as this applies to Windows with HyperV support                                                                     |
| labels                 | ✓  | ✓  | ✓  | Metadata.Annotations                                                 |                                                                                                                                   |
| links                  | x  | x  | x  |                                                                      | All containers in the same pod are accessible in Kubernetes                                                                       |
| logging                | x  | x  | x  |                                                                      | Kubernetes has built-in logging support at the node-level                                                                         |
| network_mode           | x  | x  | x  |                                                                      | Kubernetes uses its own cluster networking                                                                                        |
| networks               | ✓  | ✓  | ✓  |                                                                      | See `networks` key                                                                                                                |
| networks: aliases      | x  | x  | x  |                                                                      | See `networks` key                                                                                                                |
| networks: addresses    | x  | x  | x  |                                                                      | See `networks` key                                                                                                                |
| pid                    | ✓  | ✓  | ✓  | HostPID                                                              |                                                                                                                                   |
| ports                  | ✓  | ✓  | ✓  | Service.Spec.Ports                                                   |                                                                                                                                   |
| ports: short-syntax    | ✓  | ✓  | ✓  | Service.Spec.Ports                                                   |                                                                                                                                   |
| ports: long-syntax     | -  | -  | ✓  | Service.Spec.Ports                                                   |                                                                                                                                   |
| secrets                | -  | -  | ✓  | Secret                                                               | External Secret is not Supported                                                                                                  |
| secrets: short-syntax  | -  | -  | ✓  | Secret                                                               | External Secret is not Supported                                                                                                  |
| secrets: long-syntax   | -  | -  | ✓  | Secret                                                               | External Secret is not Supported                                                                                                  |
| security_opt           | x  | x  | x  |                                                                      | Kubernetes uses its own container naming scheme                                                                                   |
| stop_grace_period      | ✓  | ✓  | ✓  | TerminationGracePeriodSeconds                                        |                                                                                                                                   |
| stop_signal            | x  | x  | x  |                                                                      | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/30051                                  |
| sysctls                | n  | n  | n  |                                                                      |                                                                                                                                   |
| ulimits                | x  | x  | x  |                                                                      | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/3595                                   |
| userns_mode            | x  | x  | x  |                                                                      | Not supported within Kubernetes and ignored in Docker Compose Version 3                                                           |
| volumes                | ✓  | ✓  | ✓  | PersistentVolumeClaim                                                | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster                    |
| volumes: short-syntax  | ✓  | ✓  | ✓  | PersistentVolumeClaim                                                | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster                    |
| volumes: long-syntax   | -  | -  | ✓  | PersistentVolumeClaim                                                | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster                    |
| restart                | ✓  | ✓  | ✓  |                                                                      |                                                                                                                                   |
|                        |    |    |    |                                                                      |                                                                                                                                   |
| **Volume**             | x  | x  | x  |                                                                      |                                                                                                                                   |
| driver                 | x  | x  | x  |                                                                      |                                                                                                                                   |
| driver_opts            | x  | x  | x  |                                                                      |                                                                                                                                   |
| external               | x  | x  | x  |                                                                      |                                                                                                                                   |
| labels                 | x  | x  | x  |                                                                      |                                                                                                                                   |
|                        |    |    |    |                                                                      |                                                                                                                                   |
| **Network**            | x  | x  | x  |                                                                      |                                                                                                                                   |
| driver                 | x  | x  | x  |                                                                      |                                                                                                                                   |
| driver_opts            | x  | x  | x  |                                                                      |                                                                                                                                   |
| enable_ipv6            | x  | x  | x  |                                                                      |                                                                                                                                   |
| ipam                   | x  | x  | x  |                                                                      |                                                                                                                                   |
| internal               | x  | x  | x  |                                                                      |                                                                                                                                   |
| labels                 | x  | x  | x  |                                                                      |                                                                                                                                   |
| external               | x  | x  | x  |                                                                      |                                                                                                                                   |
