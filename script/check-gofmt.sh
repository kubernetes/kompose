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
# See the License for the specific language governing permissions and
# limitations under the License.


# This checks if all go source files in current directory are format using gofmt

GO_FILES=$(find . -path ./vendor -prune -o -name '*.go' -print )

for file in $GO_FILES; do
	gofmtOutput=$(gofmt -l "$file")
	if [ "$gofmtOutput" ]; then
		errors+=("$gofmtOutput")
	fi
done 


if [ ${#errors[@]} -eq 0 ]; then
	echo "gofmt OK"
else
	echo "gofmt ERROR - These files are not formatted by gofmt:"
	for err in "${errors[@]}"; do
		echo "$err"
	done
	exit 1
fi