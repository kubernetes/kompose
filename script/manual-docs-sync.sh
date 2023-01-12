#!/usr/bin/env bash

## README:
## This script is ran by running:
## cd script
## ./manual-docs-sync.sh
##
## This will take all documentation from the /docs folder of the master dir and push to the gh-pages branch (granted you have access)

DOCS_REPO_NAME="kompose"
DOCS_REPO_URL="git@github.com:kubernetes/kompose.git"
DOCS_BRANCH="gh-pages"
DOCS_FOLDER="docs"

# clone the repo
git clone "$DOCS_REPO_URL" "$DOCS_REPO_NAME"

# change to that directory (to prevent accidental pushing to master, etc.)
cd "$DOCS_REPO_NAME"

# switch to gh-pages and grab the docs folder from master
git checkout gh-pages
git checkout master docs

# Copy it all over to the current directory
cp -r docs/* .
rm -r docs

git add --all

# Check if anything changed, and if it's the case, push to origin/master.
if git commit -m 'Update docs' -m "Synchronize documentation against website" ; then
  git push
fi

# cd back to the original root folder
cd ..
