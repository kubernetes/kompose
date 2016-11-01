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

# inspired by: https://github.com/openshift/origin/blob/master/hack/godep-restore.sh

# Sometimes godep needs 'other' remotes. So add those remotes
preload-remote() {
  local orig_org="$1"
  local orig_project="$2"
  local alt_org="$3"
  local alt_project="$4"

  # Go get stinks, which is why we have the || true...
  go get -d "${orig_org}/${orig_project}" &>/dev/null || true

  repo_dir="${GOPATH}/src/${orig_org}/${orig_project}"
  pushd "${repo_dir}" > /dev/null
    git remote add "${alt_org}-remote" "https://${alt_org}/${alt_project}.git" > /dev/null || true
    git remote update > /dev/null
  popd > /dev/null
}

echo "Preloading some dependencies"
# OpenShift requires its own Kubernets fork :-(
preload-remote "k8s.io" "kubernetes" "github.com/openshift" "kubernetes"
# OpenShift requires its own glog fork
preload-remote "github.com/golang" "glog" "github.com/openshift" "glog"

echo "Starting to download all godeps. This takes a while"
godep restore
echo "Download finished into ${GOPATH}"
