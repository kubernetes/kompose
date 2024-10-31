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



# ## TEST V2
DIR="v2"
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-os.yaml"

convert::expect_success "$k8s_cmd" "$k8s_output"
convert::expect_success "$os_cmd" "$os_output"



## TEST V3
DIR="v3.0"
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/$DIR/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/$DIR/output-os.yaml"

convert::expect_success_and_warning "$k8s_cmd" "$k8s_output"
convert::expect_success_and_warning "$os_cmd" "$os_output"

######
# Test the output file behavior of kompose convert
# Default behavior without -o
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -j" "redis-deployment.json" "redis-service.json" "web-deployment.json" "web-service.json"
# Behavior with -o <filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -o output_file -j" "output_file"
# Behavior with -o <dirname>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -o $TEMP_DIR -j" "$TEMP_DIR/redis-deployment.json" "$TEMP_DIR/redis-service.json" "$TEMP_DIR/web-deployment.json" "$TEMP_DIR/web-service.json"
# Behavior with -o <dirname>/<filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -o $TEMP_DIR/output_file -j" "$TEMP_DIR/output_file"
# Behavior with -o <non-existent-dirname>/
dst=$TEMP_DIR/output_dir/
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -o $dst -j" "${dst}redis-deployment.json" "${dst}redis-service.json" "${dst}web-deployment.json" "${dst}web-service.json"
# Behavior with -o <non-existent-dirname>/<filename>
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/compose.yaml convert -o $TEMP_DIR/output_dir2/output_file -j" "$TEMP_DIR/output_dir2/output_file"

#TEST the pvc-request-size command parameter
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/pvc-request-size/compose.yaml convert -o $TEMP_DIR/output_dir2/output-k8s.json -j --pvc-request-size=300Mi" "$TEMP_DIR/output_dir2/output-k8s.json"
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/pvc-request-size/compose.yaml convert -o $TEMP_DIR/output_dir2/output-os.json -j --pvc-request-size=300Mi" "$TEMP_DIR/output_dir2/output-os.json"

######
# Test the path of build image
# Test build v2 absolute compose file
convert::check_artifacts_generated "kompose --build local -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/compose.yaml convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v2 relative compose file
relative_path=$(realpath --relative-to="$PWD" "$KOMPOSE_ROOT/script/test/fixtures/buildconfig/compose.yaml")
convert::check_artifacts_generated "kompose --build local -f $relative_path convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v3 absolute compose file with context
convert::check_artifacts_generated "kompose --build local -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/compose-v3.yaml convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"
# Test build v3 relative compose file with context
relative_path=$(realpath --relative-to="$PWD" "$KOMPOSE_ROOT/script/test/fixtures/buildconfig/compose-v3.yaml")
convert::check_artifacts_generated "kompose --build local -f $relative_path convert -o $TEMP_DIR/output_file" "$TEMP_DIR/output_file"

# #####
# Test the build config with push image
# see tests_push_image.sh for local push test
# Should warn when push image disabled
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/compose-build-no-image.yaml -o $TEMP_DIR/output_file convert --build=local --push-image-registry=whatever"
convert::expect_warning "$cmd" "Push image registry 'whatever' is specified but push image is disabled, skipping pushing to repository"

#TEST the kompose.volume.storage-class-name label
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/storage-class-name/compose.yaml convert -o $TEMP_DIR/output-k8s.yaml" "$TEMP_DIR/output-k8s.yaml"
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/storage-class-name/compose.yaml convert -o $TEMP_DIR/output-os.yaml -j" "$TEMP_DIR/output-os.yaml"

# TEST the windows volume
# windows host path to windows container
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success_and_warning "$os_cmd" "$os_output" || exit 1

# # TEST the placement
# should convert placement to affinity
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/compose-placement.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/compose-placement.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success_and_warning "$os_cmd" "$os_output" || exit 1


# test configmap volume
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/compose.yaml convert --stdout --with-kompose-annotation=false --volumes=configMap"
os_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/compose.yaml convert --stdout --with-kompose-annotation=false --volumes=configMap"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# test configmap volume using service label
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/compose-withlabel.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/compose-withlabel.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s-withlabel.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os-withlabel.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# test configmap pod generation
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-pod/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-pod/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-pod/output-os.yaml"
os_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-pod/compose.yaml convert --stdout --with-kompose-annotation=false"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test that emptyDir works
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/compose.yaml convert --with-kompose-annotation=false --stdout --volumes emptyDir"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-empty-vols-template.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/compose.yaml convert --with-kompose-annotation=false --stdout --volumes emptyDir"
os_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-empty-vols-template.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test that emptyvols works
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/compose.yaml convert --with-kompose-annotation=false --stdout --emptyvols"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/compose.yaml convert --with-kompose-annotation=false --stdout --emptyvols"
os_output="$KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success_and_warning "$os_cmd" "$os_output" || exit 1

# test service expose
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout --with-kompose-annotation=false"
ocp_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/expose/output-k8s.yaml"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/expose/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$ocp_cmd" "$ocp_output" || exit 1


# test service group by volume, not support openshift for now
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-group/compose.yaml convert --stdout --with-kompose-annotation=false --service-group-mode=volume --service-group-name=librenms-dispatcher"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/service-group/output-k8s.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1

# test merge multiple compose files
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout --with-kompose-annotation=false"
ocp_cmd="kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-k8s.yaml"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$ocp_cmd" "$ocp_output" || exit 1

# test health check
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/compose-healthcheck.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/compose-healthcheck.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# test statefulset
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/compose.yaml convert --stdout --with-kompose-annotation=false --controller statefulset"
ocp_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/compose.yaml convert --stdout --with-kompose-annotation=false --controller statefulset"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/statefulset/output-k8s.yaml"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/statefulset/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$ocp_cmd" "$ocp_output" || exit 1

# test cronjob
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/cronjob/compose.yaml convert --stdout --with-kompose-annotation=false"
ocp_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/cronjob/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/cronjob/output-k8s.yaml"
ocp_output="$KOMPOSE_ROOT/script/test/fixtures/cronjob/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$ocp_cmd" "$ocp_output" || exit 1

# test specifying volume type using service label
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose  --provider=openshift -f  $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/compose.yaml  convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test environment variables interpolation
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose  --provider=openshift -f  $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/compose.yaml  convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success_and_warning "$os_cmd" "$os_output" || exit 1

# Test single file output feature
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/single-file-output/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/single-file-output/output-k8s.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1

# Test host port and protocol feature
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/compose.yaml convert --stdout --with-kompose-annotation=false"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/output-k8s.yaml"
os_output="$KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test external traffic policy feature with valid configuration, warning is coming from the network policy.
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/compose-v1.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-k8s-v1.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/compose-v1.yaml convert --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-os-v1.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test external traffic policy feature with warning, because we have nodeport with external traffic policy
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/compose-v2.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-k8s-v2.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/compose-v2.yaml convert --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-os-v2.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success_and_warning "$os_cmd" "$os_output" || exit 1

# Test Pod security context fs group
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/fsgroup/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/fsgroup/output-k8s.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/fsgroup/compose.yaml convert --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/fsgroup/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test support for compose.yaml file
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/compose-file-support/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/compose-file-support/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1

# Test support for compose env interpolation
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/compose-env-interpolation/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/compose-env-interpolation/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test support for subpath volume
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/vols-subpath/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/vols-subpath/output-k8s.yaml"
os_cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/vols-subpath/compose.yaml convert --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/vols-subpath/output-os.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test support for network policies generation
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/network-policies/compose.yaml convert --generate-network-policies --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/network-policies/output-k8s.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1

# Test support for namespace generation
k8s_cmd="kompose -f ./script/test/fixtures/namespace/compose.yaml convert --stdout --with-kompose-annotation=false -n web"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/namespace/output-k8s.yaml"
os_cmd="kompose -f ./script/test/fixtures/namespace/compose.yaml convert --stdout --with-kompose-annotation=false -n web --provider openshift"
os_output="$KOMPOSE_ROOT/script/test/fixtures/namespace/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test support for read only root fs
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/read-only/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/read-only/output-k8s.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/read-only/compose.yaml convert --stdout --with-kompose-annotation=false --provider openshift"
os_output="$KOMPOSE_ROOT/script/test/fixtures/read-only/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test env_file support
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/env/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/env/output-k8s.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/env/compose.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/env/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# disabled until FormatEnvName can take into account conflicting/duplicate file names
# Test env_file support for multiple env_file with the same name from different dirs
# k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/env-multiple/compose.yaml convert --stdout --with-kompose-annotation=false"
# k8s_output="$KOMPOSE_ROOT/script/test/fixtures/env-multiple/output-k8s.yaml"
# os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/env-multiple/compose.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
# os_output="$KOMPOSE_ROOT/script/test/fixtures/env-multiple/output-os.yaml"
# convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
# convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test that if we have no profile a warning should raised
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/no-profile-warning/compose.yaml convert"
convert::expect_warning "$cmd" "No service selected. The profile specified in services of your compose yaml may not exist." || exit 1

# Test COMPOSE_FILE env variable is honored
export COMPOSE_FILE="$KOMPOSE_ROOT/script/test/fixtures/compose-file-env-variable/compose.yaml $KOMPOSE_ROOT/script/test/fixtures/compose-file-env-variable/alternative-compose.yaml"
k8s_cmd="kompose convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/compose-file-env-variable/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1

# Test resources names lowercase
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/resources-lowercase/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/resources-lowercase/output-k8s.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/resources-lowercase/compose.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/resources-lowercase/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test resources to generate initcontainer
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/initcontainer/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/initcontainer/output-k8s.yaml"
convert::expect_success_and_warning "$k8s_cmd" "$k8s_output" || exit 1

# Test HPA 
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/hpa/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/hpa/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1

#Test auto configmaps from files/dir
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-1.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-k8s-1.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-1.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-os-1.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

#Test auto configmaps from files/dir
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-2.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-k8s-2.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-2.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-os-2.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

#Test auto configmaps from files/dir
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-3.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-k8s-3.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/compose-3.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/configmap-file-configs/output-os-3.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test network_mode: service:
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/network-mode-service/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/network-mode-service/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1

# Test env var with status
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-with-status/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/envvars-with-status/output-k8s.yaml"
os_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-with-status/compose.yaml convert --provider openshift --stdout --with-kompose-annotation=false"
os_output="$KOMPOSE_ROOT/script/test/fixtures/envvars-with-status/output-os.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1
convert::expect_success "$os_cmd" "$os_output" || exit 1

# Test label in compose.yaml appears in the output annotation
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/label/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/label/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1

# Test deploy.labels in compose.yaml appears in the output
k8s_cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/deploy/labels/compose.yaml convert --stdout --with-kompose-annotation=false"
k8s_output="$KOMPOSE_ROOT/script/test/fixtures/deploy/labels/output-k8s.yaml"
convert::expect_success "$k8s_cmd" "$k8s_output" || exit 1