#!/usr/bin/env bash

# AVOID INVOKING THIS SCRIPT DIRECTLY -- USE `drake run publish`

set -euox pipefail

source scripts/versioning.sh

go get -u github.com/tcnksm/ghr

set +x

echo "Publishing CLI binaries for commit $full_git_version"

ghr -t $GITHUB_TOKEN -u lovethedrake -r mallard -c $full_git_version -n $rel_version $rel_version /shared/bin/
