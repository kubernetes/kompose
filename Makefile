
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
BUILD_FLAGS := -ldflags="-w -s -X github.com/kubernetes/kompose/pkg/version.GITCOMMIT=$(GITCOMMIT)"
TEST_IMAGE := kompose/tests:latest

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -o kompose main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# kompile kompose for multiple platforms
.PHONY: cross
cross:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-linux-amd64" main.go
	GOOS=linux GOARCH=arm CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-linux-arm" main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-linux-arm64" main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-windows-amd64.exe" main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-darwin-amd64" main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GO111MODULE=on go build  ${BUILD_FLAGS} -installsuffix cgo  -o "bin/kompose-darwin-arm64" main.go

.PHONY: clean
clean:
	rm -f kompose
	rm -r -f bundles

.PHONY: test-unit
test-unit:
	go test -short $(BUILD_FLAGS) -race -cover -v ./...

# Run unit tests and collect coverage
.PHONY: test-unit-cover
test-unit-cover:
	# First install packages that are dependencies of the test.
	go test -short -i -race -cover ./...
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
	./script/test/cmd/tests_new.sh

# generate commandline tests
.PHONY: gen-cmd
gen-cmd:
	./script/test/cmd/make-test.sh

# run all validation tests
.PHONY: validate
validate: gofmt vet lint

.PHONY: vet
vet:
	go vet ./pkg/...

.PHONY: lint
lint:
	golint ./pkg/...

.PHONY: gofmt
gofmt:
	./script/check-gofmt.sh

# Run all tests
.PHONY: test
test: bin test-dep  validate test-unit-cover install test-cmd

# Install all the required test-dependencies before executing tests (only valid when running `make test`)
.PHONY: test-dep
test-dep:
	go install github.com/mattn/goveralls@latest
	go install github.com/modocache/gover@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/mitchellh/gox@latest


# build docker image that is used for running all test locally
.PHONY: test-image
test-image:
	docker build -t $(TEST_IMAGE) -f script/test_in_container/Dockerfile script/test_in_container/

# run all test locally in docker image (image can be build by by build-test-image target)
.PHONY: test-container
test-container:
	docker run -v `pwd`:/opt/tmp/kompose:ro -it $(TEST_IMAGE)


.PHONY: test-k8s
test-k8s:
	./script/test_k8s/test.sh

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
.PHONY: install-golangci-lint
install-golangci-lint:
# golangci-lint version must consistent with github CI
# ref: ./.github/workflows/golangci-lint.yml
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.32.2)

.PHONY: golangci-lint
golangci-lint: install-golangci-lint
	$(GOLANGCI_LINT) run -c .golangci.yml --timeout 5m
