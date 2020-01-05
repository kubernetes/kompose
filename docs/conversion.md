# Conversion Matrix

This document outlines all possible conversion details regarding `docker-compose.yaml` values to Kubernetes / OpenShift artifacts. This covers *major* versions of Docker Compose such as 1, 2 and 3.

The current table covers all **current** possible Docker Compose keys.

__Note:__ We don't support anything 3.4 and above at the moment.

__Glossary:__

- __✓:__ Converts
- __-:__ Not in this Docker Compose Version
- __n:__ Not yet implemented
- __x:__ Not applicable / no 1-1 conversion

| Keys                   | V1 | V2 | V3 | Kubernetes / OpenShift                                      | Notes                                                                                                          |
|------------------------|----|----|----|-------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| build                  | ✓  | ✓  | ✓  |                                                             | Builds/Pushes to Docker repository. See `--build` parameter                                                    |
| build: context         | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
| build: dockerfile      | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
| build: args            | n  | n  | n  |                                                             |                                                                                                                |
| build: cache_from      | -  | -  | n  |                                                             |                                                                                                                |
| cap_add, cap_drop      | ✓  | ✓  | ✓  | Pod.Spec.Container.SecurityContext.Capabilities.Add/Drop    |                                                                                                                |
| command                | ✓  | ✓  | ✓  | Pod.Spec.Container.Command                                  |                                                                                                                |
| configs                | n  | n  | ✓  |                                                             |                                                                                                                |
| configs: short-syntax  | n  | n  | ✓  |                                                             | Only create configMap                                                                                          |
| configs: long-syntax   | n  | n  | ✓  |                                                             | If target path is /, ignore this and only create configMap                                                     |
| cgroup_parent          | x  | x  | x  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/11986               |
| container_name         | ✓  | ✓  | ✓  | Metadata.Name + Deployment.Spec.Containers.Name             |                                                                                                                |
| credential_spec        | x  | x  | x  |                                                             | Only applicable to Windows containers                                                                          |
| deploy                 | -  | -  | ✓  |                                                             |                                                                                                                |
| deploy: mode           | -  | -  | ✓  |                                                             |                                                                                                                |
| deploy: replicas       | -  | -  | ✓  | Deployment.Spec.Replicas / DeploymentConfig.Spec.Replicas   |                                                                                                                |
| deploy: placement      | -  | -  | ✓  | Pod.Spec.NodeSelector                                       |                                                                                                                |
| deploy: update_config  | -  | -  | ✓  | Workload.Spec.Strategy                                      | Deployment / DeploymentConfig                                                                                                               |
| deploy: resources      | -  | -  | ✓  | Containers.Resources.Limits.Memory / Containers.Resources.Limits.CPU | Support for memory as well as cpu                                                                     |
| deploy: restart_policy | -  | -  | ✓  | Pod generation                                              | This generated a Pod, see the [user guide on restart](http://kompose.io/user-guide/#restart)                   |
| deploy: labels         | -  | -  | ✓  | Workload.Metadata.Labels                                    | Only applied to workload resource                       |                                                                                                                |
| devices                | x  | x  | x  |                                                             | Not supported within Kubernetes, See issue https://github.com/kubernetes/kubernetes/issues/5607                |
| depends_on             | x  | x  | x  |                                                             |                                                                                                                |
| dns                    | x  | x  | x  |                                                             | Not used within Kubernetes. Kubernetes uses a managed DNS server                                               |
| dns_search             | x  | x  | x  |                                                             | See `dns` key                                                                                                  |
| domainname             | ✓  | ✓  | ✓  | Pod.Spec.SubDomain                                          |
| tmpfs                  | ✓  | ✓  | ✓  | Pod.Spec.Containers.Volumes.EmptyDir                        | Creates emptyDirvolume with medium set to Memory & mounts given directory inside container                     |
| entrypoint             | ✓  | ✓  | ✓  | Pod.Spec.Container.Command                                  | Same as command                                                                                                |
| env_file               | n  | n  | ✓  |                                                             |                                                                                                                |
| environment            | ✓  | ✓  | ✓  | Pod.Spec.Container.Env                                      |                                                                                                                |
| expose                 | ✓  | ✓  | ✓  | Service.Spec.Ports 
| endpoint_mode          | n  | n  | ✓  |                                                             | If endpoint_mode=vip, the created Service will be forced to set to NodePort type                               |
| extends                | ✓  | ✓  | ✓  |                                                             | Extends by utilizing the same image supplied                                                                   |
| external_links         | x  | x  | x  |                                                             | Kubernetes uses a flat-structure for all containers and thus external_links does not have a 1-1 conversion     |
| extra_hosts            | n  | n  | n  |                                                             |                                                                                                                |
| group_add              | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
| healthcheck            | -  | n  | ✓  |                                                             |                                                                                                                |
| hostname               | ✓  | ✓  | ✓  | Pod.Spec.HostName                                           |                                                                                                                |
| image                  | ✓  | ✓  | ✓  | Deployment.Spec.Containers.Image                            |                                                                                                                |
| isolation              | x  | x  | x  |                                                             | Not applicable as this applies to Windows with HyperV support                                                  |
| labels                 | ✓  | ✓  | ✓  | Metadata.Annotations                                        |                                                                                                                |
| links                  | x  | x  | x  |                                                             | All containers in the same pod are accessible in Kubernetes                                                    |
| logging                | x  | x  | x  |                                                             | Kubernetes has built-in logging support at the node-level                                                      |
| network_mode           | x  | x  | x  |                                                             | Kubernetes uses its own cluster networking                                                                    |
| networks               | ✓  | ✓  | ✓  |                                                             | See `networks` key                                                                                             |
| networks: aliases      | x  | x  | x  |                                                             | See `networks` key                                                                                             |
| networks: addresses    | x  | x  | x  |                                                             | See `networks` key                                                                                             |
| pid                    | ✓  | ✓  | ✓  | Pod.Spec.HostPID                                            |                                                                                                                |
| ports                  | ✓  | ✓  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| ports: short-syntax    | ✓  | ✓  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| ports: long-syntax     | -  | -  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| secrets                | -  | -  | ✓  | Secret                                                      | External Secret is not Supported                                                                               |
| secrets: short-syntax  | -  | -  | ✓  | Secret                                                      | External Secret is not Supported                                                                               |
| secrets: long-syntax   | -  | -  | ✓  | Secret                                                      | External Secret is not Supported                                                                               |
| security_opt           | x  | x  | x  |                                                             | Kubernetes uses its own container naming scheme                                                                |
| stop_grace_period      | ✓  | ✓  | ✓  | Pod.Spec.TerminationGracePeriodSeconds                      |                                                                                                                |
| stop_signal            | x  | x  | x  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/30051               |
| sysctls                | n  | n  | n  |                                                             |                                                                                                                |
| ulimits                | x  | x  | x  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/3595                |
| userns_mode            | x  | x  | x  |                                                             | Not supported within Kubernetes and ignored in Docker Compose Version 3                                        |
| volumes                | ✓  | ✓  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volumes: short-syntax  | ✓  | ✓  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volumes: long-syntax   | -  | -  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| restart                | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
|                        |    |    |    |                                                             |                                                                                                                |
| __Volume__             | x  | x  | x  |                                                             |                                                                                                                |
| driver                 | x  | x  | x  |                                                             |                                                                                                                |
| driver_opts            | x  | x  | x  |                                                             |                                                                                                                |
| external               | x  | x  | x  |                                                             |                                                                                                                |
| labels                 | x  | x  | x  |                                                             |                                                                                                                |
|                        |    |    |    |                                                             |                                                                                                                |
| __Network__            | x  | x  | x  |                                                             |                                                                                                                |
| driver                 | x  | x  | x  |                                                             |                                                                                                                |
| driver_opts            | x  | x  | x  |                                                             |                                                                                                                |
| enable_ipv6            | x  | x  | x  |                                                             |                                                                                                                |
| ipam                   | x  | x  | x  |                                                             |                                                                                                                |
| internal               | x  | x  | x  |                                                             |                                                                                                                |
| labels                 | x  | x  | x  |                                                             |                                                                                                                |
| external               | x  | x  | x  |                                                             |                                                                                                                |
