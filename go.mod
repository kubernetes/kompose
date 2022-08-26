module github.com/kubernetes/kompose

go 1.13

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.1

replace github.com/docker/libcompose => github.com/docker/libcompose v0.4.1-0.20190808084053-143e0f3f1ab9

replace github.com/docker/cli => github.com/docker/cli v20.10.16+incompatible

replace github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v1.2.1-0.20201027075954-b076d39a02e5

replace github.com/docker/docker => github.com/docker/docker v20.10.0-beta1.0.20201030232932-c2cc352355d4+incompatible

replace github.com/containerd/containerd => github.com/containerd/containerd v1.6.4

replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8

replace k8s.io/kubernetes => k8s.io/kubernetes v1.24.1

replace github.com/miekg/dns => github.com/miekg/dns v1.1.49

replace github.com/docker/distribution => github.com/docker/distribution v2.8.1+incompatible

replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.2+incompatible

require (
	github.com/containerd/containerd v1.6.4 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/go-connections v0.4.0
	github.com/docker/libcompose v0.4.0
	github.com/fatih/structs v1.1.0
	github.com/fsouza/go-dockerclient v1.7.11
	github.com/google/go-cmp v0.5.8
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/joho/godotenv v1.3.0
	github.com/novln/docker-parser v1.0.0
	github.com/openshift/api v0.0.0-20200803131051-87466835fcc0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	golang.org/x/tools v0.1.6-0.20210726203631-07bc1bf47fb2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
)
