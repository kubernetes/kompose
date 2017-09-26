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

# Test case for checking replicas option with kompose

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/script/test/cmd/lib.sh
source $KOMPOSE_ROOT/script/test_in_openshift/lib.sh

convert::print_msg "Running tests for replica option"

# Run kompose up
./kompose up --provider=openshift --volumes emptyDir --replicas 2 -f ${KOMPOSE_ROOT}/examples/docker-compose-counter.yaml; exit_status=$?

if [ $exit_status -ne 0 ]; then
    convert::print_fail "kompose up has failed"
    exit 1
fi


# Check if redis and web pods are up. Replica count: 2
convert::kompose_up_check -p "redis web" -r 2

# Run Kompose down
convert::kompose_down ${KOMPOSE_ROOT}/examples/docker-compose-counter.yaml

convert::kompose_down_check 4
