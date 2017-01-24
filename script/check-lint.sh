#!/bin/bash
set -e

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


for pkg in $(glide novendor); do
	lintOutput=$(golint "$pkg")
	# if lineOutput is not empty, save it to errros array
	if [ "$lintOutput" ]; then
		errors+=("$lintOutput")
	fi
done 


if [ ${#errors[@]} -eq 0 ]; then
	echo "golint OK"
else
	echo "golint ERRORS:"
	for err in "${errors[@]}"; do
		echo "$err"
	done
	exit 1
fi