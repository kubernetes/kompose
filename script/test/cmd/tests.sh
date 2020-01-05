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
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Unsupported depends_on key - ignoring"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/etherpad/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "Unsupported depends_on key - ignoring"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/etherpad/envs | cut -d'=' -f1)


######
# Tests related to docker-compose file in /script/test/fixtures/gitlab
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
export $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs)
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/gitlab/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/gitlab/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs | cut -d'=' -f1)


######
# Tests related to docker-compose file in /script/test/fixtures/nginx-node-redis
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
# Replacing variables with current branch and uri
# Test BuildConfig
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose.yml convert --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"

# Replacing variables with current branch and uri
# Test BuildConfig v3/OpenShift Test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose-v3.yml convert --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template-v3.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"

#Kubernetes Test for v3
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/docker-compose-v3.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-k8s-template-v3.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

######
# Tests related to docker-compose file in /script/test/fixtures/entrypoint-command
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/mem-limit
# kubernetes test

# Test the "memory limit" conversion
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/mem-limit/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# Test the "memory limit" conversion with "Mb" tagged on
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/mem-limit/docker-compose-mb.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/mem-limit/output-mb-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

######
# Tests merge multiple docker-compose files
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/dev.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-base-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-port-base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-port-prod.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-compose-port-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-port-base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-new-service-prob.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-compose-new-service-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose --provider=openshift convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-port-base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/compose-new-service-prob.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-openshift-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/service-merge-concat-base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/service-merge-concat-override.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-service-merge-concat-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Volume mount on the host \"/override/dir\" isn't supported - ignoring path on the host"

# Test other top level keys
# In merge
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/other-toplevel-dev.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/other-toplevel-ext.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-other-toplevel-merge-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# In Override
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/other-toplevel-base.yml -f $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/other-toplevel-override.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/merge-multiple-compose/output-other-toplevel-override-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

######
# Tests related to docker-compose file in /script/test/fixtures/ports-with-proto
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-proto/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/same-port-different-proto
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/same-port-different-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/same-port-different-proto/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/same-port-different-proto/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/same-port-different-proto/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/simple-vol-mounts
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/simple-vol-mounts/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/named-volume
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/docker-compose-v3.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/output-k8s-v3.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/named-volume/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/hostpath
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/docker-compose.yml convert --stdout -j --volumes hostPath"
hostpath=$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/data_sux
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%HOSTPATH%;$hostpath;g" $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/docker-compose.yml convert --stdout -j --volumes hostPath"
hostpath=$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/data_sux
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  -e "s;%HOSTPATH%;$hostpath;g" $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/docker-compose-v3.yml convert --stdout -j --volumes hostPath"
hostpath=$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/data_sux
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%HOSTPATH%;$hostpath;g" $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/docker-compose-v3.yml convert --stdout -j --volumes hostPath"
hostpath=$KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/data_sux
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  -e "s;%HOSTPATH%;$hostpath;g" $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/hostpath/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"




######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/volumes-from
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "ignoring path on the host"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/volumes-from/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "ignoring path on the host"

######
# Tests related to docker-compose file in /script/test/fixtures/volume-mounts/tmpfs
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/tmpfs/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/tmpfs/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/tmpfs/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/tmpfs/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

######
# Tests related to docker-compose file in /script/test/fixtures/envvars-separators
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-separators/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/envvars-separators/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Unsupported volume_driver key - ignoring"


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
# Kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Unsupported depends_on key - ignoring"

# OpenShift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-k8s.yml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/docker-os.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/multiple-compose-files/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "Unsupported depends_on key - ignoring"


######
# Test related to restart options in docker-compose
# kubernetes test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-onfail.json"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-unless-stopped.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-k8s-restart-unless-stopped.json" > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" /tmp/output-k8s.json

# openshift test
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-no.yml --provider openshift convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-no.json"
convert::expect_success "kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-onfail.yml --provider openshift convert --stdout -j" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-onfail.json"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/restart-options/docker-compose-restart-unless-stopped.yml --provider openshift convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/restart-options/output-os-restart-unless-stopped.json" > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" /tmp/output-os.json



#####
# Test related to hostname/domainname in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/domain/docker-compose.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/domain/output-k8s.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" /tmp/output-k8s.json

cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/domain/docker-compose-v3.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/domain/output-k8s.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" /tmp/output-k8s.json


# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/domain/docker-compose.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/domain/output-os.json" > /tmp/output-os.json
convert::expect_success "$cmd" /tmp/output-os.json

cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/domain/docker-compose-v3.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/domain/output-os.json" > /tmp/output-os.json
convert::expect_success "$cmd" /tmp/output-os.json


######
# Test key-only environment variable
export $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs)
cmd="kompose --file $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/env.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

unset $(cat $KOMPOSE_ROOT/script/test/fixtures/keyonly-envs/envs | cut -d'=' -f1)

#####
# Test related to host:port:container in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/ports-with-ip/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/ports-with-ip/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


######
# Test related to "stdin_open: true" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/stdin-true/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/stdin-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/stdin-true/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"



######
# Test related to "tty: true" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/tty-true/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/tty-true/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/tty-true/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


# Test related to "group_add" in docker-compose
# kubernetes test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/group-add/output-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# openshift test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/group-add/output-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# Test related to Failing "group_add" in docker-compose
# kubernetes test
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose-fail.yml convert --stdout -j"
# openshift test
convert::expect_failure "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/group-add/docker-compose-fail.yml convert --stdout -j"

# Test related to kompose.service.expose label in docker compose file to ensure that services are exposed properly
#kubernetes tests
# when kompose.service.expose="True"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-true.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname>"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-hostname.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname>" and kompose.service.expose.tls-secret="<secret>"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-tls.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-hostname-tls.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="True" and multiple ports in docker compose file (first port should be selected)
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-true-multiple-ports.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname>" and multiple ports in docker compose file (first port should be selected)
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-hostname-multiple-ports.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-hostname-multiple-ports.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname_1>;<hostname_2>;...<hostname_N>"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-multiple-hostname.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-multiple-hostname.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.service.expose="<hostname_1>;<hostname_2>;...<hostname_N>" and kompose.service.expose.tls-secret="<secret>"
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-multiple-hostname-tls.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/kubernetes-expose-multiple-hostname-tls.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

#openshift tests
# when kompose.service.expose="True"
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/expose-service/provider-files/openshift-expose-true.json > /tmp/output-os.json
convert::expect_success "kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose-service/compose-files/docker-compose-expose-true.yml convert --stdout -j" "/tmp/output-os.json"
# when kompose.service.expose="<hostname>"
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

# Test related to kompose.image-pull-secret label in docker compose file to ensure imagePullSecrets are populated properly.
# Kubernetes Test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/image-pull-secret/compose-files/docker-compose-image-pull-secret.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/image-pull-secret/provider-files/kubernetes-image-pull-secret.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# Test related to kompose.image-pull-secrets label in docker compose file.
# Kubernetes Tests
# when kompose.image-pull-secrets is invalid
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/image-pull-policy/compose-files/v12-fail-image-pull-policy.yml convert --stdout"
# when kompose.image-pull-secrets in v1 and 2
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/image-pull-policy/compose-files/v12-image-pull-policy.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/image-pull-policy/provider-files/kubernetes-v12-image-pull-policy.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"
# when kompose.image-pull-secrets in v3
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/image-pull-policy/compose-files/v3-image-pull-policy.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/image-pull-policy/provider-files/kubernetes-v3-image-pull-policy.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test the change in the service name
# Kubernetes Test
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-name-change/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Unsupported root level volumes key - ignoring"


# Openshift Test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/service-name-change/docker-compose.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-name-change/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "Unsupported root level volumes key - ignoring"

#####
# Test secrets
# Kubernetes Test

cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/secrets/docker-compose-secrets-long.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/secrets/output-long-k8s.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "External secrets my_other_secret is not currently supported - ignoring"

cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/secrets/docker-compose-secrets-short.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/secrets/output-short-k8s.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "External secrets my_other_secret is not currently supported - ignoring"



# Openshift Test
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/secrets/docker-compose-secrets-long.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/secrets/output-long-os.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "External secrets my_other_secret is not currently supported - ignoring"

cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/secrets/docker-compose-secrets-short.yml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/secrets/output-short-os.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "External secrets my_other_secret is not currently supported - ignoring"


#####
# Test regarding validating dockerfilepath
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/dockerfilepath/docker-compose.yml convert --stdout"

# Test regarding while label (nodeport or loadbalancer ) is provided but not ports
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/label-port/docker-compose.yml convert --stdout"


####
# test node port with port set
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-label/docker-compose.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-label/output-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/service-label/docker-compose.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/service-label/output-oc.json > /tmp/output-oc.json
convert::expect_success "$cmd" "/tmp/output-oc.json"




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

######
# Test charts generate with custom dir
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/redis-example/docker-compose.yml convert -o $TEMP_DIR -j -c" "$TEMP_DIR/Chart.yaml" "$TEMP_DIR/README.md" "$TEMP_DIR/templates/redis-deployment.json" "$TEMP_DIR/templates/redis-service.json" "$TEMP_DIR/templates/web-deployment.json" "$TEMP_DIR/templates/web-service.json"

####
# Test regarding build context (running kompose from various directories)
# Replacing variables with current branch and uri
# Test BuildConfig
CURRENT_DIR=$(pwd)
cd "$KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/"
cmd="kompose convert --provider openshift --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"
cd "$KOMPOSE_ROOT/script/test/fixtures/"
cmd="kompose convert --provider openshift --stdout -j -f nginx-node-redis/docker-compose.yml --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"
cd "$KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/node"
cmd="kompose convert --provider openshift --stdout -j -f ../docker-compose.yml --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g" $KOMPOSE_ROOT/script/test/fixtures/nginx-node-redis/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"
cd $CURRENT_DIR



#### Test docker build feature
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose-dockerfile.yml --stdout -j --build=local"
convert::expect_cmd_success "$cmd"

cmd="kompose convert  --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/buildconfig/docker-compose-dockerfile.yml --stdout -j --build=local"
convert::expect_cmd_success "$cmd"


# Test the presence of build args in buildconfig
# Replacing variables with current branch and uri
# Test BuildConfig
cmd="kompose --provider openshift -f $KOMPOSE_ROOT/script/test/fixtures/buildargs/docker-compose.yml convert --stdout -j --build build-config"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" -e "s;%URI%;$uri;g" -e "s;%REF%;$branch;g"   $KOMPOSE_ROOT/script/test/fixtures/buildargs/output-os-template.json > /tmp/output-os.json
export $(cat $KOMPOSE_ROOT/script/test/fixtures/buildargs/envs)
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "$warning"
#rm /tmp/output-buildarg-os.json

####
# Test related to change in pvc name if volume used is in the current directory
# kubernetes test
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Volume mount on the host "\"."\" isn't supported - ignoring path on the host"


# openshift test
cmd="kompose convert --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-template.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json" "Volume mount on the host "\"."\" isn't supported - ignoring path on the host"

# Test that empty-vols works
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --stdout -j --volumes emptyDir"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-empty-vols-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json" "Volume mount on the host "\"."\" isn't supported - ignoring path on the host"

#Failing test for `--volumes` option
convert::expect_failure "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml --volumes foobar"

# Test related to support docker-compose.yml beside docker-compose.yml
# Store the original path
CURRENT_DIR=$(pwd)
# Kubernetes test
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

# Test HealthCheck
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-k8s-template.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-os-template.json" > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-only-command.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-only-command-k8s-template.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-only-command.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-only-command-os-template.json" > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


# Test ConfigMap generation
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/configmap/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/configmap/output-k8s-template.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test configmap as volume
cmd="kompose convert --stdout -j --volumes=configMap -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s.json" >  /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout --provider=openshift -j --volumes=configMap -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" "$KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-oc.json" >  /tmp/output-oc.json
convert::expect_success "$cmd" "/tmp/output-oc.json"


# Test V3 Support of Docker Compose

# Test deploy mode: global
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-deploy.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-deploy-k8s.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-deploy.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-deploy-os.json" > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# Test support for cpu and memory limits + reservations
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-memcpu.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-memcpu-k8s.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-memcpu-partial.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-memcpu-partial-k8s.json" > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# Test volumes are passed correctly
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-volumes.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-volumes-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test environment variables are passed correctly
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-env-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test unset environment variables are passed correctly
export V3_HOST_ENV_TEST_SET_TO_BAR=BAR
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-unset-env.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-unset-env-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test environment variables substitution
unset foo
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-env-subs.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-env-subs.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test key/value env with env_files
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/env/docker-compose.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" $KOMPOSE_ROOT/script/test/fixtures/env/output-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/env/docker-compose.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g" $KOMPOSE_ROOT/script/test/fixtures/env/output-os.json > /tmp/output-os.json
 convert::expect_success "$cmd" "/tmp/output-os.json"

# Test that two files that are different versions fail
convert::expect_failure "kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yaml"

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-k8s-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-os-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

####
# Test `--controller`
# kubernetes test (controller=deployment)
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/controller/docker-compose.yml --stdout -j --controller=deployment"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-deployment-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# kubernetes test (controller=daemonset)
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/controller/docker-compose.yml --stdout -j --controller=daemonset"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-daemonset-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

# kubernetes test (controller=replicationcontroller)
cmd="kompose convert -f $KOMPOSE_ROOT/script/test/fixtures/controller/docker-compose.yml --stdout -j --controller=replicationcontroller"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-rc-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


# Test the "full example" from https://raw.githubusercontent.com/aanand/compose-file/master/loader/example1.env

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-full-example.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-k8s-full-example.json" > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

# Openshift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-full-example.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  "$KOMPOSE_ROOT/script/test/fixtures/v3/output-os-full-example.json" > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json"

### Test for docker-compose files present in Examples Directory

# docker compose

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-v3-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-v3-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# docker compose v3

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-v3.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-v3-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-v3.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-v3-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# counter

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-counter.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-counter-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-counter.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-counter-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


# counter v3

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-counter-v3.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-counter-v3-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-compose-counter-v3.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-counter-v3-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"


# voting

# Kubernetes
# This test also contains headless service test (not create by default)
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-voting.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-voting-k8s.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-voting.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-voting-os.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# gitlab

# Kubernetes
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/examples/docker-gitlab.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-gitlab-k8s.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"


## OpenShift
cmd="kompose convert --provider=openshift --stdout -j -f $KOMPOSE_ROOT/examples/docker-gitlab.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/examples/output-gitlab-os.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json"

## Test compose v3 global deploy
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-global.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-global-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --controller deployment --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-global.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-global-deployment-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

## Test label kompose.controller.type
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-controller-label.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-controller-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --controller deployment --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-controller-label.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-controller-template.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-controller-label-v3.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-k8s-controller-v3-template.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose --provider=openshift convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/controller/compose-controller-label-v3.yml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/controller/output-os-controller-v3-template.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-short.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-config-short.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-short-warning.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-config-short-warning.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-long.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-config-long.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-long-warning.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-config-long-warning.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

## Test compose v3.3
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-endpoint-mode-1.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-endpoint-mode-1.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-endpoint-mode-2.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-k8s-endpoint-mode-2.json > /tmp/output-k8s.json
convert::expect_success "$cmd" "/tmp/output-k8s.json"

## Test compose v3.5+
cmd="kompose convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/v3/docker-compose-3.5.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/v3/output-k8s-3.5.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"

## Test OpenShift for compose v3.3
cmd="kompose --provider openshift convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-long.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-os-config-long.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json"

cmd="kompose --provider openshift convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-config-short.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-os-config-short.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

cmd="kompose --provider openshift convert --stdout -j -f $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/compose-endpoint-mode-1.yaml"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/compose-v3.3-test/output-os-mode-1.json > /tmp/output-os.json
convert::expect_success "$cmd" "/tmp/output-os.json"

# Testing stdin feature
cmd="kompose convert --stdout -j -f -"
sed -e "s;%VERSION%;$version;g" -e "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/stdin/output-k8s.json > /tmp/output-k8s.json
cat $KOMPOSE_ROOT/script/test/fixtures/stdin/docker-compose.yaml | $cmd | diff /tmp/output-k8s.json -

# Network Translation compose
cmd="kompose -f $KOMPOSE_ROOT/script/test/fixtures/network/docker-compose-v3.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e  "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/network/output-k8s.json > /tmp/output-k8s.json
convert::expect_success_and_warning "$cmd" "/tmp/output-k8s.json"
# OpenShift test
cmd="kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/network/docker-compose-v3.yaml convert --stdout -j"
sed -e "s;%VERSION%;$version;g" -e  "s;%CMD%;$cmd;g"  $KOMPOSE_ROOT/script/test/fixtures/network/output-os.json > /tmp/output-os.json
convert::expect_success_and_warning "$cmd" "/tmp/output-os.json"

echo -e "\n"
go test -v github.com/kubernetes/kompose/script/test/cmd

rm /tmp/output-k8s.json /tmp/output-os.json
exit $EXIT_STATUS
