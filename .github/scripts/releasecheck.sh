#!/usr/bin/env bash

echo "::group::releasecheck"

echo "::debug::Checking for releasable file changes"
release=$(git diff --name-only ${INPUT_BASE_SHA} ${INPUT_SHA} | grep -cE '[^\.].*\.(go|mod|sum|md)')

if [ ${release} -gt 0 ]; then
  echo "::debug::Changes detected"
  echo "::set-output name=changed::true"
else
  echo "::debug::No changes detected"
  echo "::set-output name=changed::false"
fi
