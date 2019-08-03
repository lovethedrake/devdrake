#!/bin/sh

set -euo pipefail

if [ "$DRAKE_TAG" == "" ]; then
  export rel_version=edge
else
  export rel_version=$DRAKE_TAG
fi

export git_version=$(git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)
