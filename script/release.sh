#!/usr/bin/env bash

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

# Constants. Enter relevant repo information here.
UPSTREAM_REPO="kubernetes-incubator"
CLI="kompose"

usage() {
  echo "This will prepare $CLI for release!"
  echo ""
  echo "Requirements:"
  echo " git"
  echo " hub"
  echo " github-release"
  echo " github_changelog_generator"
  echo " GITHUB_TOKEN in your env variable"
  echo " "
  echo "Not only that, but you must have permission for:"
  echo " Tagging releases within Github"
  echo ""
}

requirements() {
  if [ ! -f /usr/bin/git ] && [ ! -f /usr/local/bin/git ]; then
    echo "No git."
    return 1
  fi

  if [ ! -f $GOPATH/bin/github-release ]; then
    echo "No $GOPATH/bin/github-release. Please run 'go get -v github.com/aktau/github-release'"
    return 1
  fi

  if [ ! -f /usr/local/bin/github_changelog_generator ]; then
    echo "github_changelog_generator required to generate the change log. Please run 'gem install github_changelog_generator"
    return 1
  fi

  if [ ! -f /usr/bin/hub ]; then
    echo "Hub needed in order to create the relevant PR's. Please install hub @ https://github.com/github/hub"
    return 1
  fi

  if [[ -z "$GITHUB_TOKEN" ]]; then
    echo "export GITHUB_TOKEN=yourtoken needed for using github-release"
  fi
}

# Clone and then change to user's upstream repo for pushing to master / opening PR's :)
clone() {
  git clone ssh://git@github.com/$UPSTREAM_REPO/$CLI.git
  if [ $? -eq 0 ]; then
        echo OK
  else
        echo FAIL
        exit
  fi
  cd $CLI
  git remote remove origin
  git remote add origin git@github.com:$ORIGIN_REPO/$CLI.git
  git checkout -b release-$1
  cd ..
}

replaceversion() {
  cd $CLI

  echo "1. Replaced version in version.go"
  sed -i "s/$1/$2/g" cmd/version.go

  echo "2. Replaced README.md versioning"
  sed -i "s/$1/$2/g" README.md
  
  cd ..
}

changelog() {
  cd $CLI
  echo "Generating changelog using github-changelog-generator"
  github_changelog_generator $UPSTREAM_REPO/$CLI -t $GITHUB_TOKEN --future-release v$1
  cd ..
}

changelog_github() {
  touch changes.txt
  echo "Write your GitHub changelog here" >> changes.txt
  $EDITOR changes.txt
}

build_binaries() {
  cd $CLI
  make cross
  cd ..
}

git_commit() {
  cd $CLI 

  BRANCH=`git symbolic-ref --short HEAD`
  if [ -z "$BRANCH" ]; then
    echo "Unable to get branch name, is this even a git repo?"
    return 1
  fi
  echo "Branch: " $BRANCH

  git add .
  git commit -m "$1 Release"
  git push origin $BRANCH
  hub pull-request -b $UPSTREAM_REPO/$CLI:master -h $ORIGIN_REPO/$CLI:$BRANCH

  cd ..
  echo ""
  echo "PR opened against master to update version"
  echo "MERGE THIS BEFORE CONTINUING"
  echo ""
}

git_pull() {
  cd $CLI
  git pull
  cd ..
}

push() {
  CHANGES=$(cat changes.txt)
  # Release it!

  echo "Creating GitHub tag"
  github-release release \
      --user $UPSTREAM_REPO \
      --repo $CLI \
      --tag v$1 \
      --name "v$1" \
      --description "$CHANGES"
  if [ $? -eq 0 ]; then
        echo UPLOAD OK 
  else 
        echo UPLOAD FAIL
        exit
  fi

  # Upload all the binaries generated in bin/
  for f in $CLI/bin/*
  do
    echo "Uploading $f binary"
    NAME=`echo $f | sed "s,$CLI/bin/,,g"`
    github-release upload \
        --user $UPSTREAM_REPO \
        --repo $CLI \
        --tag v$1 \
        --file $f \
        --name $NAME
    if [ $? -eq 0 ]; then
          echo UPLOAD OK 
    else 
          echo UPLOAD FAIL
          exit
    fi
  done

  echo "DONE"
  echo "DOUBLE CHECK IT:"
  echo "!!!"
  echo "https://github.com/$UPSTREAM_REPO/$CLI/releases/edit/$1"
  echo "!!!"
}

clean() {
  rm -rf $CLI $CLI-$1 $CLI-$1.tar.gz changes.txt
}

main() {
  local cmd=$1
  usage

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
  "Git clone master"
  "Replace version number"
  "Generate changelog"
  "Generate GitHub changelog"
  "Create PR"
  "Build binaries"
  "Upload the binaries and push to GitHub release page"
  "Clean"
  "Quit")
  select opt in "${options[@]}"
  do
      echo ""
      case $opt in
          "Git clone master")
              clone $VERSION
              ;;
          "Replace version number")
              replaceversion $PREV_VERSION $VERSION
              ;;
          "Generate changelog")
              changelog $VERSION
              ;;
          "Generate GitHub changelog")
              changelog_github $VERSION
              ;;
          "Create PR")
              git_commit $VERSION
              ;;
          "Build binaries")
              build_binaries
              ;;
          "Upload the binaries and push to GitHub release page")
              push $VERSION
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
