# Conversion Matrix

This document outlines all possible conversion details regarding `docker-compose.yaml` values to Kubernetes / OpenShift artifacts. This covers *major* versions of Docker Compose such as 1, 2 and 3.

__Note:__ due to the fast-pace nature of Docker Compose version revisions, minor versions such as 2.1, 2.2, 3.1, 3.2 or 3.3 are not supported until they are cut into a major version release such as 2 or 3.

__Glossary:__
__Y:__ Converts
__N:__ Not yet implemented
__N/A:__ Not applicable / no 1-1 conversion

| Value             | Support | K8s / OpenShift                                                  | Notes                                                                                                          |
|-------------------|---------|------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| __Service__       |         |                                                                  |                                                                                                                |
| build             | Y       | Builds/Pushes to Docker repository. See `--build` parameter      | Only supported on Version 1/2 of Docker Compose                                                                |
| cap_add, cap_drop | Y       | Pod.Spec.Container.SecurityContext.Capabilities.Add/Drop         |                                                                                                                |
| command           | Y       | Pod.Spec.Container.Command                                       |                                                                                                                |
| cgroup_parent     | N/A     |                                                                  | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/11986               |
| container_name    | Y       | Metadata.Name + Deployment.Spec.Containers.Name                  |                                                                                                                |
| devices           | N/A     |                                                                  | Not supported within Kubernetes, See issue https://github.com/kubernetes/kubernetes/issues/5607                |
| depends_on        | N/A     |                                                                  | No 1-1 mapping, Kubernetes uses a flat architecture                                                            |
| dns               | N/A     |                                                                  | Not used within Kubernetes. Kubernetes uses a managed DNS server                                               |
| dns_search        | N/A     |                                                                  | See `dns` key                                                                                                  |
| tmpfs             | Y       | Pod.Spec.Containers.Volumes.EmptyDir                             | Creates emptyDirvolume with medium set to Memory & mounts given directory inside container                     |
| entrypoint        | Y       | Pod.Spec.Container.Command                                       | Same as command                                                                                                |
| env_file          | N       |                                                                  |                                                                                                                |
| environment       | Y       | Pod.Spec.Container.Env                                           |                                                                                                                |
| expose            | Y       | Service.Spec.Ports                                               |                                                                                                                |
| extends           | Y       |                                                                  | Extends by utilizing the same image supplied                                                                   |
| external_links    | N/A     |                                                                  | Kubernetes uses a flat-structure for all containers and thus external_links does not have a 1-1 conversion     |
| extra_hosts       | N       |                                                                  |                                                                                                                |
| group_add         | N       |                                                                  |                                                                                                                |
| healthcheck       | N       |                                                                  |                                                                                                                |
| image             | Y       | Deployment.Spec.Containers.Image                                 |                                                                                                                |
| isolation         | N/A     |                                                                  | Not applicable as this applies to Windows with HyperV support                                                  |
| labels            | Y       | Metadata.Annotations                                             |                                                                                                                |
| links             | N/A     |                                                                  | All containers in the same pod are accessible in Kubernetes                                                    |
| logging           | N/A     |                                                                  | Kubernetes has built-in logging support at the node-level                                                      |
| network_mode      | N/A     |                                                                  | Kubernetes uses it's own cluster networking                                                                    |
| networks          | N/A     |                                                                  | See `networks` key                                                                                             |
| pid               | Y       | Pod.Spec.HostPID                                                 |                                                                                                                |
| ports             | Y       | Service.Spec.Ports                                               |                                                                                                                |
| security_opt      | N/A     |                                                                  | Kubernetes uses it's own container naming scheme                                                               |
| stop_grace_period | Y       | Pod.Spec.TerminationGracePeriodSeconds                           |                                                                                                                |
| stop_signal       | N/A     |                                                                  | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/30051               |
| sysctls           | N       |                                                                  |                                                                                                                |
| ulimits           | N/A     |                                                                  | Not supported within Kubernetes. See issue https://github.com/kubernetes/kubernetes/issues/3595                |
| userns_mode       | N/A     |                                                                  | Not supported within Kubernetes and ignored in Docker Compose Version 3                                        |
| volumes           | Y       | PersistentVolumeClaim                                            | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volume_driver     | N/A     |                                                                  | Different plugins for different volumes, see: https://kubernetes.io/docs/concepts/storage/volumes/             |
| volumes_from      | Y       | PersistentVolumeClaim                                            | Creates a PersistentVolumeClaim that is both shared by deployment and deployment config (OpenShift)            |
| cpu_shares        | Y       |                                                                  | No direct mapping, use `resources` key within Docker Compose Version 3 `deploy`                                |
| cpu_quota         | Y       |                                                                  | No direct mapping, use `resources` key within Docker Compose Version 3 `deploy`                                |
| cpuset            | Y       |                                                                  | No direct mapping, use `resources` key within Docker Compose Version 3 `deploy`                                |
| mem_limit         | Y       | Containers.Resources.Limits.Memory                               | Maps, but recommended to use `resources` key within Version 3 `deploy`                                         |
| memswap_limit     | N/A     |                                                                  | Removed in V3+, use mem_limit                                                                                  |
|                   |         |                                                                  |                                                                                                                |
| __Deploy__        |         |                                                                  |                                                                                                                |
| mode              | N       |                                                                  |                                                                                                                |
| replicas          | Y       | Deployment.Spec.Replicas / DeploymentConfig.Spec.Replicas        |                                                                                                                |
| placement         | N       |                                                                  |                                                                                                                |
| update_config     | N       |                                                                  |                                                                                                                |
| resources         | Y       | Containers.Resources.Limits / Containers.Resources.Requests      | Supports memory/cpu limits as well as memory/cpu requests                                                      |
| restart_policy    | Y       | Pod generation                                                   | This generated a Pod, see the [user guide on restart](http://kompose.io/user-guide/#restart)                   |
| labels            | N       |                                                                  |                                                                                                                |
|                   |         |                                                                  |                                                                                                                |
| __Volume__        | N/A     |                                                                  |                                                                                                                |
| driver            | N/A     |                                                                  |                                                                                                                |
| driver_opts       | N/A     |                                                                  |                                                                                                                |
| external          | N/A     |                                                                  |                                                                                                                |
| labels            | N/A     |                                                                  |                                                                                                                |
|                   |         |                                                                  |                                                                                                                |
| __Network__       | N/A     |                                                                  |                                                                                                                |
| driver            | N/A     |                                                                  |                                                                                                                |
| driver_opts       | N/A     |                                                                  |                                                                                                                |
| enable_ipv6       | N/A     |                                                                  |                                                                                                                |
| ipam              | N/A     |                                                                  |                                                                                                                |
| internal          | N/A     |                                                                  |                                                                                                                |
| labels            | N/A     |                                                                  |                                                                                                                |
| external          | N/A     |                                                                  |                                                                                                                |
