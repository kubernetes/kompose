# Change Log

## [v0.1.1](https://github.com/skippbox/kompose/tree/v0.1.1) (2016-10-06)
[Full Changelog](https://github.com/skippbox/kompose/compare/v0.1.0...v0.1.1)

**Implemented enhancements:**

- Persistent Volumes [\#150](https://github.com/skippbox/kompose/issues/150)
- Add flags for sliencing warning and for treating warnings as error [\#100](https://github.com/skippbox/kompose/issues/100)

**Fixed bugs:**

- kompose up always deploys to default namespace [\#162](https://github.com/skippbox/kompose/issues/162)
- godep save ./... : cannot find package "k8s.io/kubernetes/pkg/apis/authentication.k8s.io" [\#117](https://github.com/skippbox/kompose/issues/117)

**Closed issues:**

- come up with a release schedule [\#187](https://github.com/skippbox/kompose/issues/187)
- go 1.5 not building [\#181](https://github.com/skippbox/kompose/issues/181)
- `--provider` flag for kompose [\#179](https://github.com/skippbox/kompose/issues/179)
- kompose --version - print out dev tag  [\#170](https://github.com/skippbox/kompose/issues/170)
- suggestion: let `-` denote stdout for -o option [\#169](https://github.com/skippbox/kompose/issues/169)
- Proposal: make --dab/--bundle global flag [\#161](https://github.com/skippbox/kompose/issues/161)
- Support for "9995:9995/tcp" [\#158](https://github.com/skippbox/kompose/issues/158)
- `kompose up`  for OpenShift [\#152](https://github.com/skippbox/kompose/issues/152)
- Generate ImageStream for every image in DeploymentConfig [\#145](https://github.com/skippbox/kompose/issues/145)

**Merged pull requests:**

- Create PVC object for docker-compose volumes [\#186](https://github.com/skippbox/kompose/pull/186) ([surajssd](https://github.com/surajssd))
- Update .dsb references to .dab [\#184](https://github.com/skippbox/kompose/pull/184) ([cdrage](https://github.com/cdrage))
- Update README + Docker Compose Bundle references [\#183](https://github.com/skippbox/kompose/pull/183) ([cdrage](https://github.com/cdrage))
- --provider global flag for kompose [\#182](https://github.com/skippbox/kompose/pull/182) ([surajssd](https://github.com/surajssd))
- Changed version tag to reflect the tip of the branch [\#180](https://github.com/skippbox/kompose/pull/180) ([cab105](https://github.com/cab105))
- Add .gitignore for Go files + compiled Kompose file [\#178](https://github.com/skippbox/kompose/pull/178) ([cdrage](https://github.com/cdrage))
- support -o - to stdout [\#172](https://github.com/skippbox/kompose/pull/172) ([ngtuna](https://github.com/ngtuna))
- remove executable perms from docs [\#171](https://github.com/skippbox/kompose/pull/171) ([dustymabe](https://github.com/dustymabe))
- Make --dab/--bundle global flag [\#168](https://github.com/skippbox/kompose/pull/168) ([kadel](https://github.com/kadel))
- Prepare up/down for other providers [\#166](https://github.com/skippbox/kompose/pull/166) ([kadel](https://github.com/kadel))
- kompose up - Get namespace from kubeconfig [\#164](https://github.com/skippbox/kompose/pull/164) ([kadel](https://github.com/kadel))
- OpenShift - generate DeploymentConfig with ImageStream [\#160](https://github.com/skippbox/kompose/pull/160) ([kadel](https://github.com/kadel))
- Add port protocol handing for docker-compose. [\#159](https://github.com/skippbox/kompose/pull/159) ([kadel](https://github.com/kadel))
- Added flag `--suppress-warnings`, `--verbose`, `--error-on-warning` global flags [\#111](https://github.com/skippbox/kompose/pull/111) ([surajssd](https://github.com/surajssd))

## [v0.1.0](https://github.com/skippbox/kompose/tree/v0.1.0) (2016-09-09)
[Full Changelog](https://github.com/skippbox/kompose/compare/v0.0.1-beta.2...v0.1.0)

**Implemented enhancements:**

- hostPath volumes? [\#109](https://github.com/skippbox/kompose/issues/109)

**Fixed bugs:**

- Wrong output when port is missing [\#121](https://github.com/skippbox/kompose/issues/121)
- kompose convert panic on v1 compose file [\#102](https://github.com/skippbox/kompose/issues/102)
- Problems of converting volumes [\#75](https://github.com/skippbox/kompose/issues/75)
- Print warning for unsupported fields in docker-compose format   [\#71](https://github.com/skippbox/kompose/issues/71)

**Closed issues:**

- \[PROPOSAL\] Use -f as a global flag [\#138](https://github.com/skippbox/kompose/issues/138)
- Should we use libcompose project.Context{} instead of docker.Context{}? [\#134](https://github.com/skippbox/kompose/issues/134)
- services should be first in List  [\#130](https://github.com/skippbox/kompose/issues/130)
- cmd tests are not working properly [\#125](https://github.com/skippbox/kompose/issues/125)
- OpenShift conversoin - invalid DeploymentConfig [\#124](https://github.com/skippbox/kompose/issues/124)
- Create a pod of containers sharing volume [\#116](https://github.com/skippbox/kompose/issues/116)
- Release: kompose binary should be statically linked  [\#98](https://github.com/skippbox/kompose/issues/98)
- Update libcompose to v0.3.0 [\#95](https://github.com/skippbox/kompose/issues/95)
- Wrong warning about networks [\#88](https://github.com/skippbox/kompose/issues/88)
- `--stdout` output as `List` kind [\#73](https://github.com/skippbox/kompose/issues/73)
- Bug: incorrect version  [\#64](https://github.com/skippbox/kompose/issues/64)
- panic: runtime error: invalid memory address or nil pointer dereference [\#59](https://github.com/skippbox/kompose/issues/59)
- Breaking code in app.go to multiple packags [\#55](https://github.com/skippbox/kompose/issues/55)
- Write an architecture document for kompose [\#45](https://github.com/skippbox/kompose/issues/45)
- new behavior of `kompose delete` [\#41](https://github.com/skippbox/kompose/issues/41)
- Add OpenShift support [\#36](https://github.com/skippbox/kompose/issues/36)
- We don't have any tests [\#34](https://github.com/skippbox/kompose/issues/34)

**Merged pull requests:**

- Update README.md [\#143](https://github.com/skippbox/kompose/pull/143) ([luebken](https://github.com/luebken))
- Use libcompose project.Context{} instead of docker.Context{} [\#142](https://github.com/skippbox/kompose/pull/142) ([ngtuna](https://github.com/ngtuna))
- update user guide: add `kompose up`, `kompose down` [\#141](https://github.com/skippbox/kompose/pull/141) ([ngtuna](https://github.com/ngtuna))
- make --file as global flag [\#139](https://github.com/skippbox/kompose/pull/139) ([ngtuna](https://github.com/ngtuna))
- improve messages of kompose up [\#136](https://github.com/skippbox/kompose/pull/136) ([sebgoa](https://github.com/sebgoa))
- New guestbook example [\#135](https://github.com/skippbox/kompose/pull/135) ([sebgoa](https://github.com/sebgoa))
- Moves examples to docs/user-guide and adds basic roadmap to main readme [\#132](https://github.com/skippbox/kompose/pull/132) ([sebgoa](https://github.com/sebgoa))
- Add more owners [\#128](https://github.com/skippbox/kompose/pull/128) ([janetkuo](https://github.com/janetkuo))
- docker-compose - Entrypoint support [\#127](https://github.com/skippbox/kompose/pull/127) ([kadel](https://github.com/kadel))
- Fix conversion to OpenShift \(invalid DeploymentConfig\) [\#126](https://github.com/skippbox/kompose/pull/126) ([kadel](https://github.com/kadel))
- clean code [\#123](https://github.com/skippbox/kompose/pull/123) ([ngtuna](https://github.com/ngtuna))
- fix \#121: update all objects, even when port is missing [\#122](https://github.com/skippbox/kompose/pull/122) ([ngtuna](https://github.com/ngtuna))
- Update architecture doc format [\#120](https://github.com/skippbox/kompose/pull/120) ([janetkuo](https://github.com/janetkuo))
- Improve error message for invalid port [\#119](https://github.com/skippbox/kompose/pull/119) ([janetkuo](https://github.com/janetkuo))
- Remove hostPath and print warnings [\#118](https://github.com/skippbox/kompose/pull/118) ([janetkuo](https://github.com/janetkuo))
- Reuse creation of controller object code [\#115](https://github.com/skippbox/kompose/pull/115) ([surajssd](https://github.com/surajssd))
- Removed unwanted svcnames list [\#114](https://github.com/skippbox/kompose/pull/114) ([surajssd](https://github.com/surajssd))
- support kompose down subcommand [\#113](https://github.com/skippbox/kompose/pull/113) ([ngtuna](https://github.com/ngtuna))
- update Libcompose to v0.3.0 [\#112](https://github.com/skippbox/kompose/pull/112) ([kadel](https://github.com/kadel))
- Fix output comparison for cmd tests [\#110](https://github.com/skippbox/kompose/pull/110) ([surajssd](https://github.com/surajssd))
- Create service function in kubernetes utils [\#108](https://github.com/skippbox/kompose/pull/108) ([surajssd](https://github.com/surajssd))
- Abstracted port checking function [\#107](https://github.com/skippbox/kompose/pull/107) ([surajssd](https://github.com/surajssd))
- Add more unit tests for Transform [\#106](https://github.com/skippbox/kompose/pull/106) ([janetkuo](https://github.com/janetkuo))
- Support container name and args in kompose convert [\#105](https://github.com/skippbox/kompose/pull/105) ([janetkuo](https://github.com/janetkuo))
- Add unit test for komposeConvert [\#104](https://github.com/skippbox/kompose/pull/104) ([janetkuo](https://github.com/janetkuo))
- Update tests output files [\#101](https://github.com/skippbox/kompose/pull/101) ([surajssd](https://github.com/surajssd))
- Build statically linked binaries in makefile; remove make clean [\#99](https://github.com/skippbox/kompose/pull/99) ([janetkuo](https://github.com/janetkuo))
- Output List kind object when using stdout [\#94](https://github.com/skippbox/kompose/pull/94) ([surajssd](https://github.com/surajssd))
- Run tests on travis-ci [\#93](https://github.com/skippbox/kompose/pull/93) ([kadel](https://github.com/kadel))
- loader-transformer [\#91](https://github.com/skippbox/kompose/pull/91) ([ngtuna](https://github.com/ngtuna))
- enhance warning: networks, network config, volume config. Fixes \#88, \#71 [\#90](https://github.com/skippbox/kompose/pull/90) ([ngtuna](https://github.com/ngtuna))
- Functional Testing for kompose cmdline [\#89](https://github.com/skippbox/kompose/pull/89) ([surajssd](https://github.com/surajssd))
- New behavior of kompose up [\#86](https://github.com/skippbox/kompose/pull/86) ([ngtuna](https://github.com/ngtuna))
- Modularize convert into loader & transformer [\#72](https://github.com/skippbox/kompose/pull/72) ([ngtuna](https://github.com/ngtuna))

## [v0.0.1-beta.2](https://github.com/skippbox/kompose/tree/v0.0.1-beta.2) (2016-08-04)
[Full Changelog](https://github.com/skippbox/kompose/compare/v0.0.1-beta.1...v0.0.1-beta.2)

**Fixed bugs:**

- Kompose help needs improvment [\#76](https://github.com/skippbox/kompose/issues/76)

**Closed issues:**

- The example .dsb file doesn't work  [\#85](https://github.com/skippbox/kompose/issues/85)
- docker-compose labels should be converted to k8s annotations instead of labels  [\#81](https://github.com/skippbox/kompose/issues/81)
- Should we support converting to Replica Sets? [\#63](https://github.com/skippbox/kompose/issues/63)
- `targetPort` is 0 in a converted service definition [\#60](https://github.com/skippbox/kompose/issues/60)
- docker-compose service with no ports is mapped to k8s svc with no ports [\#58](https://github.com/skippbox/kompose/issues/58)
- `depends\_on` is not supported [\#57](https://github.com/skippbox/kompose/issues/57)
- Environment Variable substitution not working [\#56](https://github.com/skippbox/kompose/issues/56)
- update README for bundles, compose v2 [\#54](https://github.com/skippbox/kompose/issues/54)
- Consider changing `--from-bundles` \(bool\) to `--bundle-file` \(string\) [\#53](https://github.com/skippbox/kompose/issues/53)
- Consider changing `--rc` flag to bool and adding `--replicas` [\#52](https://github.com/skippbox/kompose/issues/52)
- Unable to go build  [\#49](https://github.com/skippbox/kompose/issues/49)
- convert file fail [\#47](https://github.com/skippbox/kompose/issues/47)
- \[Discuss\] Optimize convert function [\#44](https://github.com/skippbox/kompose/issues/44)
- Default objects of `kompose convert` [\#38](https://github.com/skippbox/kompose/issues/38)
- Idea: kompose up, ps, delete, scale redirect to kubectl  [\#27](https://github.com/skippbox/kompose/issues/27)
- Print out warning for undefined fields [\#3](https://github.com/skippbox/kompose/issues/3)

**Merged pull requests:**

- Converting compose labels to k8s annotations [\#84](https://github.com/skippbox/kompose/pull/84) ([janetkuo](https://github.com/janetkuo))
- Clean up kompose help, remove support for unimplemented commands [\#83](https://github.com/skippbox/kompose/pull/83) ([janetkuo](https://github.com/janetkuo))
- Enable warnings in stdout [\#79](https://github.com/skippbox/kompose/pull/79) ([janetkuo](https://github.com/janetkuo))
- Convert volumes in \[name:\]\[host:\]container\[:access\_mode\] format [\#78](https://github.com/skippbox/kompose/pull/78) ([janetkuo](https://github.com/janetkuo))
- Volumes default not read-only [\#77](https://github.com/skippbox/kompose/pull/77) ([janetkuo](https://github.com/janetkuo))
- Correctly log error [\#74](https://github.com/skippbox/kompose/pull/74) ([janetkuo](https://github.com/janetkuo))
- Remove the support for converting to Replica Sets [\#69](https://github.com/skippbox/kompose/pull/69) ([janetkuo](https://github.com/janetkuo))
- Warning on missing port information and no service created [\#68](https://github.com/skippbox/kompose/pull/68) ([surajssd](https://github.com/surajssd))
- Support for environment variables substitution [\#67](https://github.com/skippbox/kompose/pull/67) ([surajssd](https://github.com/surajssd))
- Development Guide: use script/godep-restore.sh [\#66](https://github.com/skippbox/kompose/pull/66) ([kadel](https://github.com/kadel))
- Allow --chart and --out to be specified together [\#65](https://github.com/skippbox/kompose/pull/65) ([janetkuo](https://github.com/janetkuo))
- Add --replicas flag and changed --rc from string to bool [\#62](https://github.com/skippbox/kompose/pull/62) ([janetkuo](https://github.com/janetkuo))
- Add --bundle,-dab flag for specifying dab file [\#61](https://github.com/skippbox/kompose/pull/61) ([janetkuo](https://github.com/janetkuo))

## [v0.0.1-beta.1](https://github.com/skippbox/kompose/tree/v0.0.1-beta.1) (2016-07-22)
[Full Changelog](https://github.com/skippbox/kompose/compare/v0.0.1-alpha...v0.0.1-beta.1)

**Closed issues:**

- Default controller object is always generated. [\#33](https://github.com/skippbox/kompose/issues/33)
- Generating both ReplicationControllers and Deployments [\#31](https://github.com/skippbox/kompose/issues/31)
- Generating both ReplicationControllers and Deployments [\#30](https://github.com/skippbox/kompose/issues/30)
- update OpenShift dependency  [\#29](https://github.com/skippbox/kompose/issues/29)
- Bug: chart only expect .json files  [\#25](https://github.com/skippbox/kompose/issues/25)
- Services only get created when there is a links key present [\#23](https://github.com/skippbox/kompose/issues/23)
- Services should be created first [\#21](https://github.com/skippbox/kompose/issues/21)
- Sometimes redundant services are printed/converted in `kompose convert` [\#20](https://github.com/skippbox/kompose/issues/20)
- Redundant file creation message [\#18](https://github.com/skippbox/kompose/issues/18)
- specify replica count [\#15](https://github.com/skippbox/kompose/issues/15)
- Output for what happened after command execution [\#13](https://github.com/skippbox/kompose/issues/13)
- Support k8s 1.3 [\#12](https://github.com/skippbox/kompose/issues/12)
- Support compose v2..v3? versions [\#11](https://github.com/skippbox/kompose/issues/11)
- Change template dir for Helm charts [\#10](https://github.com/skippbox/kompose/issues/10)
- Document unsupported fileds [\#9](https://github.com/skippbox/kompose/issues/9)
- if random docker-compose file is not present --file option does not work [\#8](https://github.com/skippbox/kompose/issues/8)
- Decide status of skippbox/kompose [\#7](https://github.com/skippbox/kompose/issues/7)
- travis build failed because "speter.net/go/exp/math/dec/inf" has been removed [\#6](https://github.com/skippbox/kompose/issues/6)
- Support docker bundles format as input [\#4](https://github.com/skippbox/kompose/issues/4)
- Support output to stdout to pipe to kubectl [\#2](https://github.com/skippbox/kompose/issues/2)
- Support output in a single file [\#1](https://github.com/skippbox/kompose/issues/1)

**Merged pull requests:**

- Fix some nits in README [\#51](https://github.com/skippbox/kompose/pull/51) ([janetkuo](https://github.com/janetkuo))
- Add a bundle example file [\#50](https://github.com/skippbox/kompose/pull/50) ([janetkuo](https://github.com/janetkuo))
- Fix failing windows build [\#48](https://github.com/skippbox/kompose/pull/48) ([kadel](https://github.com/kadel))
- Inital support for Openshift. [\#46](https://github.com/skippbox/kompose/pull/46) ([kadel](https://github.com/kadel))
- Refactor how we update controllers [\#42](https://github.com/skippbox/kompose/pull/42) ([janetkuo](https://github.com/janetkuo))
- Generate only controllers set by flag [\#35](https://github.com/skippbox/kompose/pull/35) ([kadel](https://github.com/kadel))
- Make deployment the default controller, create -rc for rc, and enable copying all types of controller to chart templates [\#32](https://github.com/skippbox/kompose/pull/32) ([janetkuo](https://github.com/janetkuo))
- Validate flags when generating charts, and prints message for file created [\#28](https://github.com/skippbox/kompose/pull/28) ([janetkuo](https://github.com/janetkuo))
- Support creating Charts when --yaml set [\#26](https://github.com/skippbox/kompose/pull/26) ([janetkuo](https://github.com/janetkuo))
- Fix the 'failed to write to file' error when --out is set [\#24](https://github.com/skippbox/kompose/pull/24) ([janetkuo](https://github.com/janetkuo))
- Allow multiple types of controllers be generated unless --out or --stdout is set [\#22](https://github.com/skippbox/kompose/pull/22) ([janetkuo](https://github.com/janetkuo))
- Remove redundant file creation message, and always overwirte files when converting [\#19](https://github.com/skippbox/kompose/pull/19) ([janetkuo](https://github.com/janetkuo))
- Support printing to stdout [\#5](https://github.com/skippbox/kompose/pull/5) ([janetkuo](https://github.com/janetkuo))

## [v0.0.1-alpha](https://github.com/skippbox/kompose/tree/v0.0.1-alpha) (2016-06-30)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*
