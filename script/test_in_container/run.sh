#!/bin/bash

set -e

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


# Container should be started with Kompose source mounted as read-only volume $KOMPOSE_TMP_SRC
# Copy Kompose sources to another directory after container starts.
# We can be sure that this container won't modify any files on hosts disk.
cp -r $KOMPOSE_TMP_SRC/ $(dirname $KOMPOSE_SRC)

make test
