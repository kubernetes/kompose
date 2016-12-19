#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
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
EXIT_STATUS=0

$KOMPOSE_ROOT/script/test/cmd/tests-cmd-conversion.sh
if [ $? -ne 0 ]; then EXIT_STATUS=1; fi
$KOMPOSE_ROOT/script/test/cmd/tests-cmd-preferencefile.sh
if [ $? -ne 0 ]; then EXIT_STATUS=1; fi

######
# Test related to restart options in docker-compose
# kubernetes test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml convert --stdout" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml convert --stdout" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-onfail.json"
# openshift test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml --provider openshift convert --stdout" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml --provider openshift convert --stdout" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-onfail.json"

######
# Test key-only envrionment variable
export $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs)
convert::expect_success "kompose --file $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/env.yml convert --stdout" "$KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/output-k8s.json"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs | cut -d'=' -f1)

exit $EXIT_STATUS
