#!/usr/bin/env bash

# AVOID INVOKING THIS SCRIPT DIRECTLY -- USE `drake run publish`

set -euox pipefail

source scripts/versioning.sh

go get -u github.com/tcnksm/ghr

set +x

echo "Publishing CLI binaries for commit $git_version"

ghr -t $GITHUB_TOKEN -u lovethedrake -r devdrake -c $git_version -delete $rel_version ./bin/drake/
