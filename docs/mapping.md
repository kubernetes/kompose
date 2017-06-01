| Kompose-Object | Description | Docker-Compose Object | Kubernetes Object |
|--------|-------------|-----------------------|-------------------|
| build  | Builds a container using linux container | build | buildconfig  |
| Context  | Either a path to the directory container _dockerfile_ or a URL to git repository | context | BuildConfig.spec.source.contextDir |
| Dockerfile | Alternate _Dockerfile_ | dockerfile | N/A This is the default file used by the BuildConfig/Docker Strategy |
| command | Override default command | command | Pod.spec.container.command |
| container_name | Specify a custom container name, rather than generated default name | container_name | ObjectMeta/name |
| depends_on | Express dependency between services | depends_on | directed service dependencies |
| environment | Add environment variable | environment | Container/env |
| expose | Expose ports without publishing them to host | expose | Service |
| extend | Extend another service, in current file or another, optionally overriding configurations | extends | N/A |
| image | Specify the image to start the container from | image | BuildConfig/image |
| labels | Add metedata to container using docker labels | labels | ObjectMeta/annotations |
| ports | Exposed ports | ports | Service |
| volumes | Mount paths or named volumes | volumes | DeploymentConfig/volume |
| volumes_from | Mount all of the volumes from another service or container | volumes_from | volumes_from N/A, use volumes instead of colocation + volume mounts of all the defined mounts on the first image | 
| entrypoint | Override the default entrypoint | entrypoint | Pod.spec.container.command |
| env_file | Add env variables from a file. Can be single value or list | env_file | N/A |
| links | Links to container in another service. Either specify a service name or a link alias (SERVICE:ALIAS), or just a service name | links | Pod.spec.container
