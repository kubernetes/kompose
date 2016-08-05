# Change Log

## [v0.0.1-beta.2](https://github.com/skippbox/kompose/tree/v0.0.1-beta.2) (2016-08-04)
[Full Changelog](https://github.com/skippbox/kompose/compare/v0.0.1-beta.1...v0.0.1-beta.2)

**Fixed bugs:**

- Kompose help needs improvment [\#76](https://github.com/skippbox/kompose/issues/76)

**Closed issues:**

- The example .dsb file doesn't work  [\#85](https://github.com/skippbox/kompose/issues/85)
- docker-compose labels should be converted to k8s annotations instead of labels  [\#81](https://github.com/skippbox/kompose/issues/81)
- Bug: incorrect version  [\#64](https://github.com/skippbox/kompose/issues/64)
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