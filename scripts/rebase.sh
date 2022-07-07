#!/usr/bin/env sh

set -e

branch=$(git rev-parse --abbrev-ref HEAD)
if [ ${branch} == 'main' ]; then
  echo "Already on the main branch, nothing to rebase"
  exit 1
fi

git fetch origin
git rebase origin/main
