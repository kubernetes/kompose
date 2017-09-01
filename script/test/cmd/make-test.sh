#!/bin/bash

# Copyright 2017 The Kubernetes Authors All rights reserved.
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
# See the License for the specific language governing pe#rmissions and
# limitations under the License.

# This script generates unit tests for Kompose
# Usage: ./script/test/cmd/make-test.sh

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)

# Directory in which the output files have to be generated
# Eg. script/test/fixtures/group-add/
echo -n "Enter the Test Directory where the files needs to be generated: "
read TEST_DIR

# One-line description for the test case
echo -ne "\nEnter a description for the tests to be added: "
read TEST_DESCRIPTION

echo -ne "\nPlease provide the name of docker-compose file to be used (docker-compose.yaml by default): "
read COMPOSE_FILE

if [ -z $COMPOSE_FILE ]; then
   COMPOSE_FILE="docker-compose.yaml"
fi


if [ -z $TEST_DIR ] || [ -z "${TEST_DESCRIPTION}" ]; then
    echo "Please provide values for TEST_DIR and TEST_DESCRIPTION in the script"
    exit 1;
fi

COMPOSE_FILE=$TEST_DIR/$COMPOSE_FILE

generate_k8s() {
    ./kompose convert -f $COMPOSE_FILE -j -o $TEST_DIR/output-k8s.json
    sed -i -e '/.*kompose.cmd.*:/ s/: .*/: "%CMD%"/' -e '/.*kompose.version.*:/ s/: .*/: "%VERSION%"/' ${TEST_DIR}/output-k8s.json
}

generate_os() {
    ./kompose convert --provider=openshift -f $COMPOSE_FILE -j -o $TEST_DIR/output-os.json
    sed -i -e '/.*kompose.cmd.*:/ s/: .*/: "%CMD%"/' -e '/.*kompose.version.*:/ s/: .*/: "%VERSION%"/' ${TEST_DIR}/output-os.json
}

# Generate k8s files
generate_k8s

# Generate OS files
generate_os

cat >> $KOMPOSE_ROOT/script/test/cmd/tests.sh <<EOF

# ${TEST_DESCRIPTION}
cmd="kompose -f \$KOMPOSE_ROOT/${COMPOSE_FILE} convert --stdout -j"
sed -e "s;%VERSION%;\$version;g" -e  "s;%CMD%;\$cmd;g"  \$KOMPOSE_ROOT/${TEST_DIR}/output-k8s.json > /tmp/output-k8s.json
convert::expect_success "kompose -f \$KOMPOSE_ROOT/$COMPOSE_FILE convert --stdout -j"
# OpenShift test
cmd="kompose --provider=openshift -f \$KOMPOSE_ROOT/${COMPOSE_FILE} convert --stdout -j"
sed -e "s;%VERSION%;\$version;g" -e  "s;%CMD%;\$cmd;g"  \$KOMPOSE_ROOT/${TEST_DIR}/output-os.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift -f \$KOMPOSE_ROOT/$COMPOSE_FILE convert --stdout -j"

EOF
