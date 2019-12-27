
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


GITCOMMIT := $(shell git rev-parse --short HEAD)
BUILD_FLAGS := -ldflags="-w -X github.com/kubernetes/kompose/pkg/version.GITCOMMIT=$(GITCOMMIT)"
PKGS = $(shell glide novendor)
TEST_IMAGE := kompose/tests:latest

default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	CGO_ENABLED=0 go build ${BUILD_FLAGS} -o kompose main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# kompile kompose for multiple platforms
.PHONY: cross
cross:
	CGO_ENABLED=0 gox -osarch="darwin/amd64 linux/amd64 linux/arm windows/amd64" -output="bin/kompose-{{.OS}}-{{.Arch}}" $(BUILD_FLAGS)

.PHONY: clean
clean:
	rm -f kompose
	rm -r -f bundles

.PHONY: test-unit
test-unit:
	go test -short $(BUILD_FLAGS) -race -cover -v $(PKGS)

# Run unit tests and collect coverage
.PHONY: test-unit-cover
test-unit-cover:
	# First install packages that are dependencies of the test.
	go test -short -i -race -cover $(PKGS)
	# go test doesn't support colleting coverage across multiple packages,
	# generate go test commands using go list and run go test for every package separately
	go list -f '"go test -short -race -cover -v -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"' github.com/kubernetes/kompose/...  | grep -v "vendor" | xargs -L 1 -P4 sh -c


# run openshift up/down tests
.PHONY: test-openshift
test-openshift:
	./script/test_in_openshift/run.sh

# run commandline tests
.PHONY: test-cmd
test-cmd:
	./script/test/cmd/tests.sh

# generate commandline tests
.PHONY: gen-cmd
gen-cmd:
	./script/test/cmd/make-test.sh

# run all validation tests
.PHONY: validate
validate: gofmt vet lint

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: lint
lint:
	golint $(PKGS)

.PHONY: gofmt
gofmt:
	./script/check-gofmt.sh

# Checks if there are nested vendor dirs inside Kompose vendor and if vendor was cleaned by glide-vc
.PHONY: check-vendor
check-vendor:
	./script/check-vendor.sh

# Run all tests
.PHONY: test
test: bin test-dep check-vendor validate test-unit-cover install test-cmd

# Install all the required test-dependencies before executing tests (only valid when running `make test`)
.PHONY: test-dep
test-dep:
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	go get github.com/Masterminds/glide
	go get github.com/sgotti/glide-vc
	go get golang.org/x/lint/golint
	go get github.com/mitchellh/gox


# build docker image that is used for running all test locally
.PHONY: test-image
test-image:
	docker build -t $(TEST_IMAGE) -f script/test_in_container/Dockerfile script/test_in_container/

# run all test locally in docker image (image can be build by by build-test-image target)
.PHONY: test-container
test-container:
	docker run -v `pwd`:/opt/tmp/kompose:ro -it $(TEST_IMAGE)

# Update vendoring
# Vendoring is a bit messy right now
.PHONY: vendor-update
vendor-update:
	glide update --strip-vendor
	glide-vc --only-code --no-tests
	find ./vendor/github.com/docker/distribution -type f -exec sed -i 's/Sirupsen/sirupsen/g' {} \;
	rm -rf vendor/github.com/Sirupsen

	# a field does not has json tag defined.
	cp script/vendor-sync/types.go.txt vendor/k8s.io/kubernetes/pkg/api/types.go


.PHONY: test-k8s
test-k8s:
	./script/test_k8s/test.sh
