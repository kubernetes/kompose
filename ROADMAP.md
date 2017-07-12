# Kompose roadmap

The goal of Kompose is to have the best possible experience converting and using Docker Compose files with Kubernetes (and OpenShift!). For both development and production environments.

Here we outline our roadmap which encompasses our goals in reaching _1.0.0_. 

__Release schedule:__ Kompose is on time-based release cycle. On a 3-4 week cycle, usually ending on the last week of each month.

### [Kompose 0.6.0](https://github.com/kubernetes/kompose/releases/tag/v0.6.0)

 - [X] Adding more support for more currently unsupported docker-compose keys (cap_add and cap_drop added) [#580](https://github.com/kubernetes/kompose/pull/580)
 - [X] Improved edge-cases for deployment (namespace + insecure-repository parameters added) [#547](https://github.com/kubernetes/kompose/pull/547)

### [Kompose 0.7.0](https://github.com/kubernetes/kompose/releases/tag/v0.7.0)

 - [X] Args support added to `build` key [#424](https://github.com/kubernetes/kompose/pull/424)
 - [X] Conversion bug fixes [#613](https://github.com/kubernetes/kompose/pull/613) [#606](https://github.com/kubernetes/kompose/pull/606) [#578](https://github.com/kubernetes/kompose/pull/578)

### Kompose 0.8.0 (end of June)

 - [X] Basic support for docker-compose v3 format [#580](https://github.com/kubernetes/kompose/issues/412) [#600](https://github.com/kubernetes/kompose/pull/600)
 - [X] Initial support for building and pushing containers locally [#97](https://github.com/kubernetes/kompose/issues/97) [#521](https://github.com/kubernetes/kompose/pull/521)

### Kompose 0.9.0 (end of July)

 - [ ] Improved support for docker-compose v3 format
 - [ ] Kompose consumable as Go library [#464](https://github.com/kubernetes/kompose/issues/464)

### Kompose 1.0.0 (???)

 - [ ] Kompose supports every possible key (that is mappable to Kubernetes)
 - [ ] Backwards-compatibility on all future releases

### x.x.x (end of October)

 - **[Deadline for exiting Kubernetes Incubator](https://github.com/kubernetes/community/blob/master/incubator.md#exiting-incubation)**
