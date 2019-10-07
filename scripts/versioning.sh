#!/bin/sh

set -euo pipefail

export git_version=$(git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)
rel_version=$(git tag --list 'v*' --points-at HEAD | tail -n 1)

if [ "$rel_version" == "" ]; then
  git_branch=$(git rev-parse --abbrev-ref HEAD)
  if [ "$git_branch" == "master" ]; then
    rel_version=edge
  else
    rel_version=unstable
  fi
fi

export rel_version
