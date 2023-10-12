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

# Constants. Enter relevant repo information here.
UPSTREAM_REPO="kubernetes"
CLI="kompose"

usage() {
  echo "This will prepare $CLI for release!"
  echo ""
  echo "Requirements:"
  echo " git"
  echo " gh"
  echo " "
  echo "Not only that, but you must have permission for:"
  echo " Tagging releases within Github"
  echo ""
}

requirements() {

  if ! hash git 2>/dev/null; then
    echo "ERROR: No git."
    exit 0
  fi

}

# Make sure that upstream had been added to the repo 
init_sync() {
  CURRENT_ORIGIN=`git config --get remote.origin.url`
  CURRENT_UPSTREAM=`git config --get remote.upstream.url`
  ORIGIN="git@github.com:$ORIGIN_REPO/$CLI.git"
  UPSTREAM="git@github.com:$UPSTREAM_REPO/$CLI.git"

  if [ $CURRENT_ORIGIN != $ORIGIN ]; then
    echo "Origin repo must be set to $ORIGIN"
    exit 0
  fi

  if [ $CURRENT_UPSTREAM != $UPSTREAM ]; then
    echo "Upstream repo must be set to $UPSTREAM"
    exit 0
  fi

  git checkout main
  git fetch upstream
  git merge upstream/main
  git checkout -b release-$1
}

replaceversion() {
  echo "Replaced version in pkg/version/version.go"
  sed -i "s/$1/$2/g" pkg/version/version.go || gsed -i "s/$1/$2/g" pkg/version/version.go

  echo "Replaced version in README.md"
  sed -i "s/$1/$2/g" README.md || gsed -i "s/$1/$2/g" README.md

  echo "Replaced version in docs/installation.md"
  sed -i "s/$1/$2/g" docs/installation.md || gsed -i "s/$1/$2/g" docs/installation.md
  sed -i "s/$1/$2/g" site/docs/installation.md || gsed -i "s/$1/$2/g" site/docs/installation.md

  echo "Replaced version in build/VERSION"
  sed -i "s/$1/$2/g" build/VERSION || gsed -i "s/$1/$2/g" build/VERSION


  echo "Ignore above errors if you're using macOS"
}

build_binaries() {
  make cross
}

create_tarballs() {
  # cd into the bin directory so we don't have '/bin' inside the tarball
  cd bin
  for f in *
  do
    tar cvzf $f.tar.gz $f
  done
  cd ..
}

git_commit() {
  BRANCH=`git symbolic-ref --short HEAD`
  if [ -z "$BRANCH" ]; then
    echo "Unable to get branch name, is this even a git repo?"
    return 1
  fi
  echo "Branch: " $BRANCH

  git add .
  git commit -m "$1 Release"
  git push origin $BRANCH
  gh pr create

  echo ""
  echo "PR opened against main to update version"
  echo "MERGE THIS BEFORE CONTINUING"
  echo ""
}

git_pull() {
  git pull
}


git_sync() {
  git fetch upstream main
  git rebase upstream/main
}

generate_install_guide() {
  echo "
# Installation

__Linux and macOS:__

\`\`\`sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v$1/kompose-linux-amd64 -o kompose

# Linux ARM64
curl -L https://github.com/kubernetes/kompose/releases/download/v$1/kompose-linux-arm64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v$1/kompose-darwin-amd64 -o kompose

# macOS ARM64
curl -L https://github.com/kubernetes/kompose/releases/download/v$1/kompose-darwin-arm64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
\`\`\`

__Windows:__

Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v$1/kompose-windows-amd64.exe) and add the binary to your PATH.

__Checksums:__

| Filename        | SHA256 Hash |
| ------------- |:-------------:|" > install_guide.txt

  touch bin/SHA256_SUM

  for f in bin/*
  do
    HASH=`sha256sum $f | head -n1 | awk '{print $1;}'`
    NAME=`echo $f | sed "s,bin/,,g"`
    echo "[$NAME](https://github.com/kubernetes/kompose/releases/download/v$1/$NAME) | $HASH" >> install_guide.txt
    echo "$HASH $NAME" >> bin/SHA256_SUM
  done

}

push() {

  echo "!!PLEASE READ!!
  1. Say YES to GitHub generating the release notes
  2. Append install_guide.txt to the TOP of the release notes when prompted
  3. Double check that the binaries have been UPLOADED!
  "

  gh release create v$1 bin/*

  echo "DONE"
  echo "DOUBLE CHECK IT:"
  echo "!!!"
  echo "https://github.com/$UPSTREAM_REPO/$CLI/releases/edit/$1"
  echo "!!!"
}

clean() {
  rm install_guide.txt
  rm -r bin/*
}

main() {
  local cmd=$1
  usage

  requirements

  echo "What is your Github username? (location of your $CLI fork)"
  read ORIGIN_REPO 
  echo "You entered: $ORIGIN_REPO"
  echo ""
  
  echo ""
  echo "First, please enter the version of the NEW release: "
  read VERSION
  echo "You entered: $VERSION"
  echo ""

  echo ""
  echo "Second, please enter the version of the LAST release: "
  read PREV_VERSION
  echo "You entered: $PREV_VERSION"
  echo ""

  clear

  echo "Now! It's time to go through each step of releasing $CLI!"
  echo "If one of these steps fails / does not work, simply re-run ./release.sh"
  echo "Re-enter the information at the beginning and continue on the failed step"
  echo ""

  PS3='Please enter your choice: '
  options=(
  "Initial sync with upstream"
  "Replace version number"
  "Create PR"
  "Sync with upstream"
  "Build binaries"
  "Create tarballs"
  "Generate install guide"
  "Upload the binaries and push to GitHub release page"
  "Update the website"
  "Clean"
  "Quit")
  select opt in "${options[@]}"
  do
      echo ""
      case $opt in
          "Initial sync with upstream")
              init_sync $VERSION
              ;;
          "Replace version number")
              replaceversion $PREV_VERSION $VERSION
              ;;
          "Create PR")
              git_commit $VERSION
              ;;
          "Sync with upstream")
              git_sync
              ;;
          "Build binaries")
              build_binaries
              ;;
          "Create tarballs")
              create_tarballs
              ;;
          "Generate install guide")
              generate_install_guide $VERSION
              ;;
          "Upload the binaries and push to GitHub release page")
              push $VERSION
              ;;
          "Update the website")
              echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
              echo "Run ./script/manual-docs-sync.sh on the main branch"
              echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
              ;;
          "Clean")
              clean $VERSION
              ;;
          "Quit")
              clear
              break
              ;;
          *) echo invalid option;;
      esac
      echo ""
  done
}

main "$@"
