#!/usr/bin/env bash
set -e

if [ "$1" == "" ]; then
    echo "Must pass in another git sha"
    echo "e.g: ./benchmarks/compare_to.sh master contacts_1KB.lua"
    exit 1
fi

if [ "$2" == "" ]; then
    echo "Must pass in a target lua script to benchmark against"
    echo "e.g: ./benchmarks/compare_to.sh master contacts_1KB.lua"
    exit 1
fi

OTHER_SHA=$1
CURRENT_SHA=`git rev-parse --abbrev-ref HEAD`

function banner() {
    echo ""
    echo "#################################################"
    echo "#"
    echo "#   $1"
    echo "#"
    echo "#################################################"
    echo "#################################################"
    echo ""
}

function run() {
    local sha=$1
    local lua_script=$2

    banner "switching branch to $sha"
    git checkout $sha

    banner "compiling $sha"
    make bins

    banner "benchmarking $sha"
    $PWD/benchmarks/runner/runner -loadtest -script=$lua_script
}

# Go to root dir
while [ "$(basename $PWD)" != "zanzibar" ]; do
    cd ..
done

run $CURRENT_SHA $2
run $OTHER_SHA $2

git checkout $CURRENT_SHA