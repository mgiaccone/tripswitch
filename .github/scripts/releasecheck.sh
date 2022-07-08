#!/usr/bin/env bash

echo "::echo::on"

current_hash=$(git rev-list --no-merges -n 1 HEAD)
previous_hash=$(git rev-list --no-merges -n 1 HEAD^)

echo "::debug::Current commit: ${current_hash}"
echo "::debug::Previous commit: ${previous_hash}"

echo "::debug::Commit log"
git --no-pager log --decorate=short --pretty=oneline -n2

echo "::debug::Commit diff"
git diff --name-only ${previous_hash} ${current_hash}

echo "::debug::Checking diff with previous commit for releasable file changes"
release=$(git diff --name-only ${previous_hash} ${current_hash} | grep -cE '[^\.].*\.(go|mod|sum|md)')

if [ ${release} -gt 0 ]; then
  echo "::set-output name=changed::true"
else
  echo "::set-output name=changed::false"
fi
