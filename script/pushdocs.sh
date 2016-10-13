#!/usr/bin/env bash
# Copyright 2016 Skippbox, Ltd All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
#     Unless required by applicable law or agreed to in writing, software
#     distributed under the License is distributed on an "AS IS" BASIS,
#     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#     See the License for the specific language governing permissions and
#     limitations under the License.

if [ "$TRAVIS_BRANCH" != "master" ] || [ "$BUILD_DOCS" != "yes" ] || [ "$TRAVIS_SECURE_ENV_VARS" == "false" ] || [ "$TRAVIS_PULL_REQUEST" != "false" ] ; then
    exit 0
fi


DOCS_GIT_REPO="github.com/skippbox/kompose.git"

# Install dev version of mkdocs
pip install -U git+https://github.com/mkdocs/mkdocs.git

# Change to the "docs" folder
cd docs

# Push to the "gh-pages" branch
mkdocs gh-deploy --clean --force --remote-name https://$GITHUB_API_KEY@$DOCS_GIT_REPO
