#!/usr/bin/env bash

echo "::echo::on"

echo "::debug::INPUT_SHA ${INPUT_SHA}"
echo "::debug::INPUT_BASE_SHA ${INPUT_BASE_SHA}"

echo "::debug::Commit log"
git --no-pager log --decorate=short --pretty=oneline -n2

echo "::debug::Commit diff"
git diff --name-only ${INPUT_BASE_SHA} ${INPUT_SHA}

echo "::debug::Checking diff with previous commit for releasable file changes"
release=$(git diff --name-only ${INPUT_BASE_SHA} ${INPUT_SHA} | grep -cE '[^\.].*\.(go|mod|sum|md)')

if [ ${release} -gt 0 ]; then
  echo "::set-output name=changed::true"
else
  echo "::set-output name=changed::false"
fi
