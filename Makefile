
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
BUILD_FLAGS := -ldflags="-w -X github.com/kubernetes/kompose/cmd.GITCOMMIT=$(GITCOMMIT)"
PKGS = $(shell glide novendor)
TEST_IMAGE := kompose/tests:latest

default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} -o kompose main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# kompile kompose for multiple platforms
.PHONY: cross
cross:
	gox -osarch="darwin/amd64 linux/amd64 linux/arm windows/amd64" -output="bin/kompose-{{.OS}}-{{.Arch}}" $(BUILD_FLAGS)

.PHONY: clean
clean:
	rm -f kompose
	rm -r -f bundles

.PHONY: test-unit
test-unit:
	go test $(BUILD_FLAGS) -race -cover -v $(PKGS)

# Run unit tests and collect coverage
.PHONY: test-unit-cover
test-unit-cover:
	# First install packages that are dependencies of the test. 
	go test -i -race -cover $(PKGS)
	# go test doesn't support colleting coverage across multiple packages,
	# generate go test commands using go list and run go test for every package separately 
	go list -f '"go test -race -cover -v -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"' github.com/kubernetes/kompose/...  | grep -v "vendor" | xargs -L 1 -P4 sh -c 


# run openshift up/down tests
.PHONY: test-openshift
test-openshift:
	./script/test_in_openshift/run.sh

# run commandline tests
.PHONY: test-cmd
test-cmd:
	./script/test/cmd/tests.sh

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
test: test-dep check-vendor validate test-unit-cover install test-cmd

# Install all the required test-dependencies before executing tests (only valid when running `make test`)
.PHONY: test-dep
test-dep:
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	go get github.com/Masterminds/glide
	go get github.com/sgotti/glide-vc
	go get github.com/golang/lint/golint

# build docker image that is used for running all test localy
.PHONY: test-image
test-image:
	docker build -t $(TEST_IMAGE) -f script/test_in_container/Dockerfile script/test_in_container/

# run all test localy in docker image (image can be build by by build-test-image target)
.PHONY: test-container
test-container:
	docker run -v `pwd`:/opt/tmp/kompose:ro -it $(TEST_IMAGE) 

.PHONE: test-k8s
test-k8s:
	./script/test_k8s/test.sh
