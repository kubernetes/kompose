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

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
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

######
# Tests related to docker-compose file in /script/test/fixtures/etherpad
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout"

convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose-no-image.yml convert --stdout"

export $(cat $KOMPOSE_ROOT/script/test/fixtures/etherpad/envs)
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/etherpad/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json" "Unsupported depends_on key - ignoring"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/etherpad/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout -j" "/tmp/output-os.json" "Unsupported depends_on key - ignoring"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/etherpad/envs | cut -d'=' -f1)


######
# Tests related to docker-compose file in /script/test/fixtures/gitlab
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
export $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs)
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/gitlab/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/gitlab/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs | cut -d'=' -f1)


######
# Tests related to docker-compose file in /script/test/fixtures/nginx-node-redis
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"


# openshift test

# Replacing variables with current branch and uri
# Test BuildConfig
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j --build build-config" "/tmp/output-os.json" "$warning"


######
# Tests related to docker-compose file in /script/test/fixtures/entrypoint-command
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/mem-limit
# kubernetes test

# Test the "memory limit" conversion
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/mem-limit/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# Test the "memory limit" conversion with "Mb" tagged on
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose-mb.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/mem-limit/output-mb-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose-mb.yml convert --stdout -j" "/tmp/output-k8s.json"



######
# Tests related to docker-compose file in /script/test/fixtures/ports-with-proto
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/simple-vol-mounts
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/volumes-from
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json" "ignoring path on the host"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j" "/tmp/output-os.json" "ignoring path on the host"



######
# Tests related to docker-compose file in /script/test/fixtures/envvars-separators
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-separators/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/envvars-separators/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-separators/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json" "Unsupported volume_driver key - ignoring"


######
# Tests related to unknown arguments with cli commands
convert::expect_failure "kompose up $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml -j" "Unknown Argument docker-gitlab.yml"
convert::expect_failure "kompose down $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml -j" "Unknown Argument docker-gitlab.yml"
convert::expect_failure "kompose convert $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml -j" "Unknown Argument docker-gitlab.yml"

# Tests related to kompose --bundle convert usage and that setting the compose file results in a failure
# We no longer test --bundle, however, this command may come back in the future thus we are keeping the test commented
# convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/bundles/foo.yml --bundle $KOMPOSE_ROOT/script/test/fixtures/bundles/dab/docker-compose-bundle.dab convert"
# convert::expect_success "kompose --bundle $KOMPOSE_ROOT/script/test/fixtures/bundles/dab/docker-compose-bundle.dab convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/bundles/dab/output-k8s.json"
# convert::expect_success_and_warning "kompose --bundle $KOMPOSE_ROOT/script/test/fixtures/bundles/dsb/docker-voting-bundle.dsb convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/bundles/dsb/output-k8s.json"  "Unsupported Networks key - ignoring"

# Test related to multiple-compose files
# Kubernets test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j" "/tmp/output-k8s.json" "Unsupported depends_on key - ignoring"

# OpenShift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j" "/tmp/output-os.json" "Unsupported depends_on key - ignoring"


######
# Test related to restart options in docker-compose
# kubernetes test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-onfail.json"
# openshift test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml --provider openshift convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml --provider openshift convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-onfail.json"


######
# Test key-only envrionment variable
export $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs)
cmd="kompose --file $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/env.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose --file $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/env.yml convert --stdout -j" "/tmp/output-k8s.json"

unset $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs | cut -d'=' -f1)

#####
# Test related to host:port:container in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-ip/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-ip/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-ip/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"


######
# Test related to "stdin_open: true" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/stdin-true/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/stdin-true/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"



######
# Test related to "tty: true" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/tty-true/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"


# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/tty-true/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"


# Test related to "group_add" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/group-add/output-k8s.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json"
# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/group-add/output-os.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j" "/tmp/output-os.json"

# Test related to Failing "group_add" in docker-compose
# kubernetes test
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose-fail.yml convert --stdout -j"
# openshift test
convert::expect_failure "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose-fail.yml convert --stdout -j"

# Test related to kompose.expose.service label in docker compose file to ensure that services are exposed properly
#kubernetes tests
# when kompose.service.expose="True"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-true.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j" "/tmp/output-k8s.json"
# when kompose.expose.service="<hostname>"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-hostname.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml convert --stdout -j" "/tmp/output-k8s.json"
# when kompose.service.expose="True" and multiple ports in docker compose file (first port should be selected)
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-true-multiple-ports.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml convert --stdout -j" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname>" and multiple ports in docker compose file (first port should be selected)
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-hostname-multiple-ports.json > /tmp/output-k8s.json
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml convert --stdout -j" "/tmp/output-k8s.json"

#openshift tests
# when kompose.service.expose="True"
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/openshift-expose-true.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j" "/tmp/output-os.json"
# when kompose.expose.service="<hostname>"
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/openshift-expose-hostname.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml convert --stdout -j" "/tmp/output-os.json"
# when kompose.service.expose="True" and multiple ports in docker compose file (first port should be selected)
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/openshift-expose-true-multiple-ports.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml convert --stdout -j" "/tmp/output-os.json"
# when kompose.service.expose="<hostname>" and multiple ports in docker compose file (first port should be selected)
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/openshift-expose-hostname-multiple-ports.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml convert --stdout -j" "/tmp/output-os.json"


# Test the change in the service name
# Kubernetes Test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-name-change/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j" "/tmp/output-k8s.json" "Unsupported root level volumes key - ignoring"


# Openshift Test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-name-change/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j" "/tmp/output-os.json" "Unsupported root level volumes key - ignoring"


# Test regarding validating dockerfilepath
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/dockerfilepath/docker-compose.yml convert --stdout"

# Test regarding while label (nodeport or loadbalancer ) is provided but not ports
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/label-port/docker-compose.yml convert --stdout"


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

####
# Test regarding build context (running kompose from various directories)
# Replacing variables with current branch and uri
# Test BuildConfig
CURRENT_DIR=$(pwd)
cd "$KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/"
cmd="kompose convert --provider openshift --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose convert --provider openshift --stdout -j --build build-config" "/tmp/output-os.json" "$warning"
cd "$KOMPOSE_ROOT/script/test/fixtures/"
cmd="kompose convert --provider openshift --stdout -j -f nginx-node-redis/docker-compose.yml --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose convert --provider openshift --stdout -j -f nginx-node-redis/docker-compose.yml --build build-config" "/tmp/output-os.json" "$warning"
cd "$KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/node"
cmd="kompose convert --provider openshift --stdout -j -f ../docker-compose.yml --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose convert  --provider openshift --stdout -j -f ../docker-compose.yml --build build-config" "/tmp/output-os.json" "$warning"
cd $CURRENT_DIR



# Test the presence of build args in buildconfig
# Replacing variables with current branch and uri
# Test BuildConfig
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/buildargs/docker-compose.yml convert --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g"   $KOMPOSE_ROOT/script/test/fixtures/buildargs/output-os-template.json > /tmp/output-os.json
export $(cat $KOMPOSE_ROOT/script/test/fixtures/buildargs/envs)
convert::expect_success_and_warning "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/buildargs/docker-compose.yml convert --stdout -j --build build-config" "/tmp/output-os.json" "$warning"
#rm /tmp/output-buildarg-os.json

####
# Test related to change in pvc name if volume used is in the current directory
# kubernetes test
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j" "/tmp/output-k8s.json" "Volume mount on the host "\"."\" isn't supported - ignoring path on the host"


# openshift test
cmd="kompose convert --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "kompose convert --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j" "/tmp/output-os.json" "Volume mount on the host "\"."\" isn't supported - ignoring path on the host"


# Test related to support docker-compose.yaml beside docker-compose.yml
# Store the original path
CURRENT_DIR=$(pwd)
# Kubernets test
cd "$KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/"
sed -e "s;%VERSION%;$version;g"   $KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j" "/tmp/output-k8s.json"
cd "$KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/yml"
sed -e "s;%VERSION%;$version;g"   $KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/yml/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j" "/tmp/output-k8s.json"

# OpenShift test
cd "$KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/"
sed -e "s;%VERSION%;$version;g"   $KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift convert --stdout -j" "/tmp/output-os.json"
cd "$KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/yml"
sed -e "s;%VERSION%;$version;g"   $KOMPOSE_ROOT/script/test/fixtures/yaml-and-yml/yml/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose --provider=openshift convert --stdout -j" "/tmp/output-os.json"

# Return back to the original path
cd $CURRENT_DIR

# Test V3 Support of Docker Compose

# Test support for cpu and memory limits + reservations
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-memcpu.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-memcpu-k8s.json" > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-memcpu.yaml" "/tmp/output-k8s.json"

# Test volumes are passed correctly
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-volumes.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-volumes-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-volumes.yaml" "/tmp/output-k8s.json"


# Test environment variables are passed correctly
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-env-k8s.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env.yaml" "/tmp/output-k8s.json"


# Test environment variables substitution
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env-subs.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-env-subs.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env-subs.yaml" "/tmp/output-k8s.json"


# Test that two files that are different versions fail
convert::expect_failure "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml"

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-os-template.json > /tmp/output-os.json
convert::expect_success "kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml" "/tmp/output-os.json"


# Test the "full example" from https://raw.githubusercontent.com/aanand/compose-file/master/loader/example1.env

# Kubernetes
convert::expect_success_and_warning "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-full-example.yaml" "$KOMPOSE_ROOT/script/test/fixtures/v3/output-k8s-full-example.json"

# Openshift
convert::expect_success_and_warning "kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-full-example.yaml" "$KOMPOSE_ROOT/script/test/fixtures/v3/output-os-full-example.json"

rm /tmp/output-k8s.json /tmp/output-os.json
exit $EXIT_STATUS
