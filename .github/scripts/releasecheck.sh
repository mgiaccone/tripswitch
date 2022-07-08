#!/usr/bin/env bash

echo "::echo::on"

git --no-pager log --decorate=short --pretty=oneline -n5
git diff --name-only ${INPUT_BASE_SHA} ${INPUT_SHA}

echo "::debug::Checking diff with previous commit for releasable file changes"
release=$(git diff --name-only ${INPUT_BASE_SHA} ${INPUT_SHA} | grep -cE '[^\.].*\.(go|mod|sum|md)')

if [ ${release} -gt 0 ]; then
  echo "::set-output name=changed::true"
else
  echo "::set-output name=changed::false"
fi
