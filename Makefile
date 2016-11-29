
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

.PHONY: all

KOMPOSE_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \
	-e TESTDIRS \
	-e TESTFLAGS \
	-e TESTVERBOSE

BIND_DIR := bundles

default: bin

all: validate
	CGO_ENABLED=1 ./script/make.sh

bin:
	CGO_ENABLED=1 ./script/make.sh binary

cross:
	# CGO_ENABLED=1 ./script/make.sh binary-cross
	./script/make.sh binary-cross

clean:
	./script/make.sh clean

test-unit:
	./script/make.sh test-unit
test-cmd:
	./script/make.sh test-cmd

validate: gofmt vet

vet:
	./script/make.sh validate-vet
lint:
	./script/make.sh validate-lint
gofmt:
	./script/make.sh validate-gofmt
