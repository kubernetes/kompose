#!/usr/bin/env bash

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
# See the License for the specific language governing permissions and
# limitations under the License.

# Reference Link: https://github.com/theochem/horton/blob/master/tools/qa/fix_detached_head.sh

# Only do something if the HEAD is detached.
if [ $(git rev-parse --abbrev-ref HEAD) == 'HEAD' ]; then
    # Give the  some random name that no serious person would use.
    git checkout -b build_branch
fi
