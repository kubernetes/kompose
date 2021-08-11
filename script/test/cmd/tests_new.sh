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

######
# Test the output file behavior of kompose convert
# Default behavior without -o
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -j" "redis-deployment.json" "redis-service.json" "web-deployment.json" "web-service.json"
# Behavior with -o <filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o output_file -j" "output_file"
# Behavior with -o <dirname>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o $TEMP_DIR -j" "$TEMP_DIR/redis-deployment.json" "$TEMP_DIR/redis-service.json" "$TEMP_DIR/web-deployment.json" "$TEMP_DIR/web-service.json"
# Behavior with -o <dirname>/<filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o $TEMP_DIR/output_file -j" "$TEMP_DIR/output_file"
# Behavior with -o <non-existent-dirname>/
dst=$TEMP_DIR/output_dir/
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o $dst -j" "${dst}redis-deployment.json" "${dst}redis-service.json" "${dst}web-deployment.json" "${dst}web-service.json"
# Behavior with -o <non-existent-dirname>/<filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o $TEMP_DIR/output_dir2/output_file -j" "$TEMP_DIR/output_dir2/output_file"

#TEST the pvc-request-size command parameter
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/pvc-request-size/docker-compose.yml convert -o $TEMP_DIR/output_dir2/output-k8s.json -j --pvc-request-size=300Mi" "$TEMP_DIR/output_dir2/output-k8s.json"
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/pvc-request-size/docker-compose.yml convert -o $TEMP_DIR/output_dir2/output-os.json -j --pvc-request-size=300Mi" "$TEMP_DIR/output_dir2/output-os.json"

######
# Test the path of build image
# Test build v2 absolute compose file
convert::check_artifacts_generated "kompose --build local -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose.yml convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v2 relative compose file
relative_path=$(realpath --relative-to="$PWD" "$KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose.yml")
convert::check_artifacts_generated "kompose --build local -f $relative_path convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v3 absolute compose file with context
convert::check_artifacts_generated "kompose --build local -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose-v3.yml convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v3 relative compose file with context
relative_path=$(realpath --relative-to="$PWD" "$KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose-v3.yml")
convert::check_artifacts_generated "kompose --build local -f $relative_path convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"

#####
# Test the build config with push image
# see tests_push_image.sh for local push test
# Should warn when push image disabled
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose-build-no-image.yml -o $TEMP_DIR/output_file convert --build=local --push-image-registry=whatever"
convert::expect_warning "$cmd" "Push image registry 'whatever' is specified but push image is disabled, skipping pushing to repository"
