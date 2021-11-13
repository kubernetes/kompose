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

#TEST the kompose.volume.storage-class-name label
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/storage-class-name/docker-compose.yml convert -o $TEMP_DIR/output-k8s.json -j" "$TEMP_DIR/output-k8s.json"
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/storage-class-name/docker-compose.yml convert -o $TEMP_DIR/output-os.json -j" "$TEMP_DIR/output-os.json"

# TEST the windows volume
# windows host path to windows container
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-os.json"
convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"

# TEST the placement
# should convert placement to affinity
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/docker-compose-placement.yaml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/docker-compose-placement.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-os.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success_and_warning "$os_cmd" "$os_output"


# test configmap volume
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml convert --stdout -j --with-kompose-annotation=false --volumes=configMap"
os_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml convert --stdout -j --with-kompose-annotation=false --volumes=configMap"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"

# test configmap volume using service label
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose-withlabel.yml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose-withlabel.yml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s-withlabel.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os-withlabel.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"

# Test that emptyDir works
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout -j --volumes emptyDir"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-empty-vols-template.json"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout -j --volumes emptyDir"
os_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-empty-vols-template.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success_and_warning "$os_cmd" "$os_output"

# Test that emptyvols works
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout -j --emptyvols"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-empty-vols-template.json"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout -j --emptyvols"
os_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-empty-vols-template.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success_and_warning "$os_cmd" "$os_output"

# test service expose
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout -j --with-kompose-annotation=false"
ocp_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/expose/output-k8s.json"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/expose/output-ocp.json"
convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$ocp_cmd" "$ocp_output"


# test service group by volume, not support openshift for now
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-group/compose.yaml convert --stdout -j --with-kompose-annotation=false --service-group-mode=volume"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/service-group/output-k8s.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"

# test merge multiple compose files
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout -j --with-kompose-annotation=false"
ocp_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-k8s.json"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-ocp.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success "$ocp_cmd" "$ocp_output"

# test health check
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-healthcheck.yaml convert --stdout -j --service-group-mode=label --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-healthcheck.yaml convert --stdout -j --service-group-mode=label --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-os.json"
convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"

# test statefulset
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false --controller statefulset"
ocp_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false --controller statefulset"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/statefulset/output-k8s.json"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/statefulset/output-os.json"
convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$ocp_cmd" "$ocp_output"

# test specifying volume type using service label
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/docker-compose.yaml convert --stdout -j --with-kompose-annotation=false"
os_cmd="kompose  --provider=openshift -f  $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/docker-compose.yaml  convert --stdout -j --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-k8s.json"
os_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-os.json"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"