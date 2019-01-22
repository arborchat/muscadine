#!/bin/sh
#usage: ./build.sh <options for `go build`>

# find the directory containing this script
script_dir="$(dirname "$(command -v "$0")")"

# ensure that git commands are run from within this repository
cd "$script_dir" || exit 1

commit="$(git rev-parse HEAD)"
suffix=""
version_file="$(pwd)/version.go"

# check if git is clean. If not, notify user and taint the build
if ! git diff --stat --exit-code ||\
   ! git diff --stat --cached --exit-code ; then
    suffix="-modified"
    echo "Warning: your current working directory contains unstaged or uncommitted changes.
Building a \"modified\" binary
Run 'git diff && git diff --cached' to see your unmodified changes"
fi

# write the file and format it
echo "package main; const Version = \"$commit$suffix\"" > "$version_file"
gofmt -s -w "$version_file"

# show the user
echo "Wrote file $version_file"

# ensure dependencies are clean
dep ensure

# actually perform the build
go build "$@"

# reset the version.go file so that the git repo doesn't appear modified
git checkout HEAD -- "$version_file"
