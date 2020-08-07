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

# for mac
if type "greadlink" > /dev/null; then
  KOMPOSE_ROOT=$(greadlink -f $(dirname "${BASH_SOURCE}")/../../..)
else
  KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
fi

source $KOMPOSE_ROOT/script/test/cmd/lib.sh

# Get current branch and remote url of git repository
branch=$(git branch | grep \* | cut -d ' ' -f2-)

uri=$(git config --get remote.origin.url)
if [[ $uri != *".git"* ]]; then
    uri="${uri}.git"
fi

# Get version
version=`kompose version`

# Warning Template
warning="Buildconfig using $uri::$branch as source."
# Replacing variables with current branch and uri
sed -e "s;%VERSION%;$version;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json



## TEST V2
DIR="v2"
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-os.json"

convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"



## TEST V3
DIR="v3.0"
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-os.json"

convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success_and_warning "$os_cmd" "$os_output"



