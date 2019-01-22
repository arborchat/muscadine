#!/bin/bash

set -euo pipefail

# only build on master
if [ "$TRAVIS_PULL_REQUEST" != "false" ] || [ "$TRAVIS_BRANCH" != "master" ]; then
    echo "Not building release, this is a pull request"
    exit 0
fi

# get hub if we don't have it
if ! command -v hub > /dev/null 2>&1 ; then
    go get github.com/github/hub
fi

release_flags=""
head_commit="$(git rev-parse HEAD 2> /dev/null)"

# build pre-release if not on a tag, otherwise build release
if ! git describe --tags --exact-match HEAD > /dev/null 2>&1 ; then
    echo "Building pre-release, not tagged commit"
    readonly tag="release-$(echo "$head_commit" | head -c 7)"
    release_flags="--prerelease"
else 
    echo "Building release, tagged commit"
    readonly tag="$(git describe --tags --exact-match HEAD 2> /dev/null)"
fi

readonly project="muscadine"
readonly bin_name="$project"
readonly message="Automatic build"

declare -A arch_for_os
arch_for_os["darwin"]="amd64"
arch_for_os["windows"]="amd64"
arch_for_os["linux"]="amd64 arm64 ppc64 ppc64le mips64 mips64le s390x"
arch_for_os["openbsd"]="amd64"

# create the release and upload artifacts
hub release create $release_flags --message="$message" --commitish="$head_commit" "$tag"
for os in darwin linux windows openbsd; do
    for arch in ${arch_for_os["$os"]}; do
        archive_name="$project-$os-$arch.tar.gz"
        echo "Building $project for $os on $arch"
        env GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 ./build.sh -o "$bin_name" &&\
        tar czf "$archive_name" "$bin_name" &&\
        rm "$bin_name"
        echo "Adding $archive_name to release $tag"
        hub release edit --message="$message" --attach="$archive_name#muscadine_${os}_${arch}" "$tag"
    done
done
