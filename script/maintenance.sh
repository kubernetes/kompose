#!/usr/bin/env msu

msu_require "console"

BASE=${PWD}
LATEST_COMMIT_FILE=${BASE}/script/latest-commit
LATEST_COMMIT=$(cat "${LATEST_COMMIT_FILE}")

log "cloning original repository: orderedlist/minimal"
rm -rf .original
git clone https://github.com/orderedlist/minimal.git .original
pushd .original/ > /dev/null

NEWEST_COMMIT=$(git rev-parse HEAD)

log "checking if there are new commits"
if [ "${LATEST_COMMIT}" == "${NEWEST_COMMIT}" ]
then
    success "no new commits"
    exit
fi

log "copying stylesheets"
cp stylesheets/* "${BASE}/css/"

log "copying js files"
cp javascripts/* "${BASE}/js/"

log "viewing changes in index.html, since last commit"
git log -p "${LATEST_COMMIT}..HEAD" -- index.html

log "storing the newest commit hash"
echo "${NEWEST_COMMIT}" > "${LATEST_COMMIT_FILE}"

success "done"
