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

# Here are tests for pushing image on authentication and custom registry.
# These tests only work on local for authentication inconvenient.
# Prerequisites:
# * `docker.io` account and docker login successfully
# * custom registry which login as well. Or a local hosted registry.
# * `jq` installed

# Variables
TEMP_DIR="/tmp/kompose"
mkdir -p $TEMP_DIR
mkdir -p "$TEMP_DIR/build"

DOCKER_LOGIN_USER="lexcao" # TODO change this to your account for pushing to docker.io
COMPOSE_FILE="$TEMP_DIR/docker-compose-push.yml"
BUILD_FILE="$TEMP_DIR/build/Dockerfile"
CUSTOM_REGISTRY="localhost:5000" # TODO change this to your local registry

# Custom compose file based on parameter
build_file_content="FROM busybox
                    RUN touch /test"
echo "$build_file_content" >> "$BUILD_FILE"

compose_file_content="version: \"2\"

services:
    foo:
        build: \"./build\"
        image: docker.io/$DOCKER_LOGIN_USER/foobar"

echo "$compose_file_content" >> "$COMPOSE_FILE"

# Some helper functions
function get_docker_hub_tag() {
  local image=$1
  local tag=$2
  curl "https://hub.docker.com/v2/repositories/$image/tags/$tag/" | jq
}

function get_custom_registry_tag() {
  local image=$1
  local tag=$2
  curl "http://$CUSTOM_REGISTRY/v2/$image/manifests/$tag" -I
}

###################################################################################
cmd="kompose convert -f $COMPOSE_FILE -o $TEMP_DIR/output_file --build=local --push-image=true"

printf "Push image without custom registry default to docker.io\n"
echo "executing cmd '$cmd'"
$cmd
printf "\nVerify push success...\n"
get_docker_hub_tag "$DOCKER_LOGIN_USER/foobar" "latest"

#######
printf "\nPush image with custom registry\n"
cmd="$cmd --push-image-registry=$CUSTOM_REGISTRY"
echo "executing cmd '$cmd'"
$cmd
#kompose convert -f "$COMPOSE_FILE" -o "$TEMP_DIR/output_file" --build=local --push-image=true
printf "\nVerify push success...\n"
get_custom_registry_tag "$DOCKER_LOGIN_USER/foobar" "latest"

# Clean resource
rm -rf $TEMP_DIR
