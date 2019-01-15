#!/bin/bash

set -euo pipefail

if [ "$TRAVIS_PULL_REQUEST" != "false" ] || [ "$TRAVIS_BRANCH" != "master" ]; then
    echo "Not building release, this is a pull request"
    exit 0
fi

if ! command -v hub 2>&1 > /dev/null ; then
    go get github.com/github/hub
fi

hub_flags=""
head_commit="$(git rev-parse HEAD 2> /dev/null)"

if ! git describe --tags --exact-match HEAD > /dev/null 2>&1 ; then
    echo "Building pre-release, not tagged commit"
    readonly tag="release-$(echo ${head_commit} | head -c 7)"
    hub_flags=" --prerelease "
else 
    echo "Building release, tagged commit"
    readonly tag="$(git describe --tags --exact-match HEAD 2> /dev/null)"
fi

readonly project="muscadine"
readonly bin_name="$project"

hub release create $hub_flags  --message="Automatic build" --commitish="$head_commit" "$tag"
for os in darwin linux windows openbsd; do
    archive_name="$project-$os.tar.gz"
    echo "Building $project for $os"
    env GOOS="$os" CGO_ENABLED=0 go build -o "$bin_name" &&\
    tar czf "$archive_name" "$bin_name" &&\
    rm "$bin_name"
    echo "Adding $archive_name to release $tag"
    hub release edit  --attach="$archive_name#Muscadine for $os" "$tag"
done
