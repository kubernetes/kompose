#!/usr/bin/env bash
set -e

export KOMPOSE_PKG='github.com/skippbox/kompose'

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
	validate-gofmt
	validate-git-marks
	validate-lint
	validate-vet
	binary

	# test-unit
	# test-integration
	# test-acceptance
)
bundle() {
    local bundle="$1"; shift
    echo "---> Making bundle: $(basename "$bundle") (in $DEST)"
    source "script/$bundle" "$@"
}

if [ $# -lt 1 ]; then
    bundles=(${DEFAULT_BUNDLES[@]})
else
    bundles=($@)
fi
for bundle in ${bundles[@]}; do
    export DEST=.
    ABS_DEST="$(cd "$DEST" && pwd -P)"
    bundle "$bundle"
    echo
done

