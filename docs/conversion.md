---
layout: default
permalink: /conversion/
---

# Conversion

This document outlines all the conversion details regarding `docker-compose.yaml` values to Kubernetes / OpenShift artifacts.

| Value | Version | Supported | K8s / OpenShift | Notes |
|-------------------|---------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| __SERVICE__ |  |  |  |  |
| build |  | Y | OpenShift: [BuildConfig](https://docs.openshift.com/enterprise/3.1/dev_guide/builds.html#defining-a-buildconfig) | Converts, but local builds are not yet supported. See issue [97](https://github.com/kubernetes-incubator/kompose/issues/97) |
| cap_add, cap_drop |  | N |  |  |
| command |  | Y | [Pod.Spec.Container.Command](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_container) |  |
| cgroup_parent |  | N |  | No compatibility with Kubernetes / OpenShift. Limited use-cases with Docker. |
| container_name |  | Y | Mapped to both [Metadata.Name](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/identifiers.md) and [Deployment.Spec.Containers.Name](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/identifiers.md) |  |
| deploy | v3 | N |  | Upcoming support started |
| devices |  | N |  | Not supported within Kubernetes, see this [issue](https://github.com/kubernetes/kubernetes/issues/5607) |
| depends_on |  | N |  |  |
| dns |  | N |  |  |
| dns_search |  | N |  |  |
| tmpfs |  | N |  |  |
| entrypoint |  | Y | [Pod.Spec.Container.Command](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_container) | Same as `command` |
| env_file |  | N |  |  |
| environment |  | Y | [Pod.Spec.Container.Env](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_envvar) |  |
| expose |  | Y | [Service.Spec.Ports](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_containerport) |  |
| extends | v2 | N |  |  |
| external_links |  | N |  |  |
| extra_hosts |  | N |  |  |
| group_add |  | N |  |  |
| healthcheck | v2.1/v3 | N |  |  |
| image |  | Y | [Deployment.Spec.Containers.Image](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_container) |  |
| isolation |  | N |  |  |
| labels |  | Y | [Metadata.Annotations](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_objectmeta) |  |
| links |  | N |  |  |
| logging |  | N |  |  |
| network_mode |  | N |  |  |
| networks |  | N |  |  |
| pid |  | N |  |  |
| ports |  | Y | [Service.Spec.Ports](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_containerport) |  |
| security_opt |  | N |  |  |
| stop_grace_period |  | N |  |  |
| stop_signal |  | N |  |  |
| sysctls |  | N |  |  |
| ulimits |  | N |  | See this [issue](https://github.com/kubernetes/kubernetes/issues/3595) on the k8s repo |
| userns_mode |  | N |  |  |
| volumes |  | Y | [PersistentVolumeClaim](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_PersistentVolumeClaim) | Creates a PersistentVolumeClaim. Can only be created if there is already a PersistentVolume within the cluster |
| volume_driver | v2 | N |  |  |
| volumes_from | v2 | N |  |  |
| cpu_shares | v2 | N |  |  |
| cpu_quota | v2 | N |  |  |
| cpuset | v2 | N |  |  |
| mem_limit | v2 | Y | [...Containers.Resources.Limits.Memory](https://kubernetes.io/docs/api-reference/v1/definitions/#_v1_resourcefieldselector) |  |
| memswap_limit | v2 | N |  | Use `mem_limit` |
| __VOLUME__ |  |  |  |  |
| driver |  | N |  |  |
| driver_opts |  | N |  |  |
| external |  | N |  |  |
| labels |  | N |  |  |
| __NETWORK__ |  |  |  |  |
| driver |  | N |  |  |
| driver_opts |  | N |  |  |
| enable_ipv6 |  | N |  |  |
| ipam |  | N |  |  |
| internal |  | N |  |  |
| labels |  | N |  |  |
| external |  | N |  |  |
