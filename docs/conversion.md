# Conversion Matrix

This document outlines all possible conversion details regarding `docker-compose.yaml` values to Kubernetes / OpenShift artifacts. This covers *major* versions of Docker Compose such as 1, 2 and 3.

The current table covers all **current** possible Docker Compose keys.

__Note:__ due to the fast-pace nature of Docker Compose version revisions, minor versions such as 2.1, 2.2, 3.1, 3.2 or 3.3 are not supported until they are cut into a major version release such as 2 or 3.

__Glossary:__

- __✓:__ Converts
- __-:__ Not in this Docker Compose Version
- __N:__ Not yet implemented
- __X:__ Not applicable / no 1-1 conversion

| Keys                   | V1 | V2 | V3 | Kubernetes / OpenShift                                      | Notes                                                                                                          |
|------------------------|----|----|----|-------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| build                  | ✓  | ✓  | N  | Builds/Pushes to Docker repository. See `--build` parameter | Only supported on Version 1/2 of Docker Compose                                                                |
| build: context         | ✓  | ✓  | N  |                                                             |                                                                                                                |
| build: dockerfile      | ✓  | ✓  | N  |                                                             |                                                                                                                |
| build: args            | N  | N  | N  |                                                             |                                                                                                                |
| build: cache_from      | -  | -  | N  |                                                             |                                                                                                                |
| cap_add, cap_drop      | ✓  | ✓  | ✓  | Pod.Spec.Container.SecurityContext.Capabilities.Add/Drop    |                                                                                                                |
| command                | ✓  | ✓  | ✓  | Pod.Spec.Container.Command                                  |                                                                                                                |
| configs                | N  | N  | N  |                                                             |                                                                                                                |
| configs: short-syntax  | N  | N  | N  |                                                             |                                                                                                                |
| configs: long-syntax   | N  | N  | N  |                                                             |                                                                                                                |
| cgroup_parent          | X  | X  | X  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/11986               |
| container_name         | ✓  | ✓  | ✓  | Metadata.Name + Deployment.Spec.Containers.Name             |                                                                                                                |
| credential_spec        | N  | N  | N  |                                                             |                                                                                                                |
| deploy                 | -  | -  | ✓  |                                                             |                                                                                                                |
| deploy: mode           | -  | -  | ✓  |                                                             |                                                                                                                |
| deploy: replicas       | -  | -  | ✓  | Deployment.Spec.Replicas / DeploymentConfig.Spec.Replicas   |                                                                                                                |
| deploy: placement      | -  | -  | N  |                                                             |                                                                                                                |
| deploy: update_config  | -  | -  | N  |                                                             |                                                                                                                |
| deploy: resources      | -  | -  | ✓  | Containers.Resources.Limits.Memory                          | Support for memory but not CPU                                                                                 |
| deploy: restart_policy | -  | -  | ✓  | Pod generation                                              | This generated a Pod, see the [user guide on restart](http://kompose.io/user-guide/#restart)                   |
| deploy: labels         | -  | -  | N  |                                                             |                                                                                                                |
| devices                | X  | X  | X  |                                                             | Not supported within Kubernetes, See issue https://github.com/kubernetes/kubernetes/issues/5607                |
| depends_on             | X  | X  | X  |                                                             |                                                                                                                |
| dns                    | X  | X  | X  |                                                             | Not used within Kubernetes. Kubernetes uses a managed DNS server                                               |
| dns_search             | X  | X  | X  |                                                             | See `dns` key                                                                                                  |
| tmpfs                  | ✓  | ✓  | ✓  | Pod.Spec.Containers.Volumes.EmptyDir                        | Creates emptyDirvolume with medium set to Memory & mounts given directory inside container                     |
| entrypoint             | ✓  | ✓  | ✓  | Pod.Spec.Container.Command                                  | Same as command                                                                                                |
| env_file               | N  | N  | N  |                                                             |                                                                                                                |
| environment            | ✓  | ✓  | ✓  | Pod.Spec.Container.Env                                      |                                                                                                                |
| expose                 | ✓  | ✓  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| extends                | ✓  | ✓  | ✓  |                                                             | Extends by utilizing the same image supplied                                                                   |
| external_links         | X  | X  | X  |                                                             | Kubernetes uses a flat-structure for all containers and thus external_links does not have a 1-1 conversion     |
| extra_hosts            | N  | N  | N  |                                                             |                                                                                                                |
| group_add              | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
| healthcheck            | -  | N  | N  |                                                             |                                                                                                                |
| image                  | ✓  | ✓  | ✓  | Deployment.Spec.Containers.Image                            |                                                                                                                |
| isolation              | X  | X  | X  |                                                             | Not applicable as this applies to Windows with HyperV support                                                  |
| labels                 | ✓  | ✓  | ✓  | Metadata.Annotations                                        |                                                                                                                |
| links                  | X  | X  | x  |                                                             | All containers in the same pod are accessible in Kubernetes                                                    |
| logging                | X  | x  | X  |                                                             | Kubernetes has built-in logging support at the node-level                                                      |
| network_mode           | X  | X  | X  |                                                             | Kubernetes uses it's own cluster networking                                                                    |
| networks               | X  | X  | X  |                                                             | See `networks` key                                                                                             |
| networks: aliases      | X  | X  | X  |                                                             | See `networks` key                                                                                             |
| networks: addresses    | X  | X  | X  |                                                             | See `networks` key                                                                                             |
| pid                    | ✓  | ✓  | ✓  | Pod.Spec.HostPID                                            |                                                                                                                |
| ports                  | ✓  | ✓  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| ports: short-syntax    | ✓  | ✓  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| ports: long-syntax     | -  | -  | ✓  | Service.Spec.Ports                                          |                                                                                                                |
| secrets                | -  | -  | N  |                                                             |                                                                                                                |
| secrets: short-syntax  | -  | -  | N  |                                                             |                                                                                                                |
| secrets: long-syntax   | -  | -  | N  |                                                             |                                                                                                                |
| security_opt           | X  | X  | X  |                                                             | Kubernetes uses it's own container naming scheme                                                               |
| stop_grace_period      | ✓  | ✓  | ✓  | Pod.Spec.TerminationGracePeriodSeconds                      |                                                                                                                |
| stop_signal            | X  | X  | X  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/30051               |
| sysctls                | N  | N  | N  |                                                             |                                                                                                                |
| ulimits                | X  | X  | X  |                                                             | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/3595                |
| userns_mode            | X  | X  | X  |                                                             | Not supported within Kubernetes and ignored in Docker Compose Version 3                                        |
| volumes                | ✓  | ✓  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volumes: short-syntax  | ✓  | ✓  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volumes: long-syntax   | -  | -  | ✓  | PersistentVolumeClaim                                       | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| restart                | ✓  | ✓  | ✓  |                                                             |                                                                                                                |
|                        |    |    |    |                                                             |                                                                                                                |
| __Volume__             | X  | X  | X  |                                                             |                                                                                                                |
| driver                 | X  | X  | X  |                                                             |                                                                                                                |
| driver_opts            | X  | X  | X  |                                                             |                                                                                                                |
| external               | X  | X  | X  |                                                             |                                                                                                                |
| labels                 | X  | X  | X  |                                                             |                                                                                                                |
|                        |    |    |    |                                                             |                                                                                                                |
| __Network__            | X  | X  | X  |                                                             |                                                                                                                |
| driver                 | X  | X  | X  |                                                             |                                                                                                                |
| driver_opts            | X  | X  | X  |                                                             |                                                                                                                |
| enable_ipv6            | X  | X  | X  |                                                             |                                                                                                                |
| ipam                   | X  | X  | X  |                                                             |                                                                                                                |
| internal               | X  | X  | X  |                                                             |                                                                                                                |
| labels                 | X  | X  | X  |                                                             |                                                                                                                |
| external               | X  | X  | X  |                                                             |                                                                                                                |
