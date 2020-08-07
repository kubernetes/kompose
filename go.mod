module github.com/kubernetes/kompose

go 1.13

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0

replace github.com/docker/libcompose => github.com/docker/libcompose v0.4.1-0.20171025083809-57bd716502dc

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20180529093712-df6e38b81a94

replace github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v0.0.0-20160323030313-93e72a773fad

replace github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20180220021536-8e435b8279f2+incompatible

require (
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/docker/cli v0.0.0-00010101000000-000000000000
	github.com/docker/libcompose v0.4.0
	github.com/fatih/structs v1.1.0
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/mattn/go-shellwords v1.0.10 // indirect
	github.com/novln/docker-parser v1.0.0
	github.com/opencontainers/selinux v1.6.0 // indirect
	github.com/openshift/api v0.0.0-20200803131051-87466835fcc0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.0-rc.2
	k8s.io/apimachinery v0.19.0-rc.2
)
