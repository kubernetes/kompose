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

# Test case for kompose up/down with etherpad

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/kompose/script/test/cmd/lib.sh
source $KOMPOSE_ROOT/script/test_in_openshift/lib.sh
openshift_exit_status=0

convert::start_test "Functional tests on OpenShift"

if [[ -n "${TRAVIS}" ]]; then
    install_oc_client
fi

if [ -z $(whereis oc | awk '{ print $2 }') ]; then
    convert::print_fail "Please install the oc binary to run tests\n"
    exit 1
fi

convert::oc_cluster_up
convert::oc_registry_login

if [ -z $1 ]; then
    for test_case in $KOMPOSE_ROOT/script/test_in_openshift/tests/*.sh; do
	$test_case; exit_status=$?
	if [ $exit_status -ne 0 ]; then
	    openshift_exit_status=1
	fi
	convert::oc_cleanup
    done
else
    # Run individual test scripts
    test_to_run=$1
    $test_to_run; exit_status=$?
    if [ $exit_status -ne 0 ]; then
	openshift_exit_status=1
    fi
fi

convert::oc_cluster_down

exit $openshift_exit_status
