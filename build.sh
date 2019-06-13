#!/bin/sh
#usage: ./build.sh <options for `go build`>

# find the directory containing this script
script_dir="$(dirname "$(command -v "$0")")"

# ensure that git commands are run from within this repository
cd "$script_dir" || exit 1

commit="$(git rev-parse HEAD)"
suffix=""

# check if git is clean. If not, notify user and taint the build
if ! git diff --exit-code > /dev/null 2>&1 ||\
   ! git diff --cached --exit-code > /dev/null 2>&1 ; then
    suffix="-modified"
    echo "Warning: your current working directory contains unstaged or uncommitted changes.
Building a \"modified\" binary
Run 'git diff && git diff --cached' to see your unmodified changes"
fi

go build -ldflags "-X main.Version=$commit$suffix" $@
