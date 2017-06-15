#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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


KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/script/test/cmd/lib.sh
source $KOMPOSE_ROOT/script/test_in_openshift/lib.sh

convert::print_msg "Running tests for docker compose v3"

docker_compose_file="${KOMPOSE_ROOT}/script/test_in_openshift/compose-files/v3-docker-compose.yaml"

# Run kompose up
convert::kompose_up $docker_compose_file

# Check if redis and web pods are up. Replica count: 2
convert::kompose_up_check -p "redis web"

# Run Kompose down
convert::kompose_down $docker_compose_file

convert::kompose_down_check 2
