#!/usr/bin/env bash


## README:
## This script is ran by running:
## cd script
## ./manual-docs-sync.sh
##
## This will take all documentation from the /docs folder of the master dir, add jekyll-related-metadata
## and push to the gh-pages branch.

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

# Remove README.md from docs folder as it isn't relevant
rm docs/README.md

# Use introduction.md instead as the main index page
mv docs/introduction.md index.md

# Check that index.md has the appropriate Jekyll format
index="index.md"
if cat $index | head -n 1 | grep "\-\-\-";
then
echo "index.md already contains Jekyll format"
else
# Remove ".md" from the name
name=${index%".md"}
echo "Adding Jekyll file format to $index"
jekyll="---
layout: default
---
"
echo -e "$jekyll\n$(cat $index)" > $index
fi

# clean-up the docs and convert to jekyll-friendly docs
cd docs
for filename in *.md; do
    if cat $filename | head -n 1 | grep "\-\-\-";
    then
    echo "$filename already contains Jekyll format"
    else
    # Remove ".md" from the name
    name=${filename%".md"}
    echo "Adding Jekyll file format to $filename"
    jekyll="---
layout: default
permalink: /$name/
redirect_from: 
  - /docs/$name.md/
---
"
    echo -e "$jekyll\n$(cat $filename)" > $filename
    fi
done
cd ..

git add --all

# Check if anything changed, and if it's the case, push to origin/master.
if git commit -m 'Update docs' -m "Synchronize documentation against website" ; then
  git push
fi

# cd back to the original root folder
cd ..
