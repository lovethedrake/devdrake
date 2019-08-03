#!/usr/bin/env bash

# AVOID INVOKING THIS SCRIPT DIRECTLY -- USE `drake run env`

set -euo pipefail

echo "---------- Dumping \`env | grep DRAKE\` for debug purposes ----------"

env | grep DRAKE || true
