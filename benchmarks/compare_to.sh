#!/usr/bin/env bash
set -e

if [ "$1" == "" ]; then
    echo "Must pass in another git sha"
    echo "e.g: ./benchmarks/compare_to.sh master benchmarks/contacts_1KB.lua"
    exit 1
fi

if [ "$2" == "" ]; then
    echo "Must pass in a target lua script to benchmark against"
    echo "e.g: ./benchmarks/compare_to.sh master benchmarks/contacts_1KB.lua"
    exit 1
fi

OTHER_SHA=$1
CURRENT_SHA=`git rev-parse --abbrev-ref HEAD`
BENCHMARK_FILE=$2

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

    banner "switching branch to $sha"
    git checkout $sha

    rm -f stdout.txt
    rm -f stderr.txt

    banner "compiling $sha"
    make bins 1>stdout.txt 2>stderr.txt | true
    local exit_code="${PIPESTATUS[0]}"
    if [ "$exit_code" -ne "0" ]; then
        cat stdout.txt
        cat stderr.txt 1>&2
        exit $exit_code
    fi
}

function benchmark_runner() {
    local sha=$1
    local lua_script=$2

    banner "benchmarking $sha"
    banner "benchmark output $sha" | tee -a benchmark.txt 1>/dev/null
    $PWD/benchmarks/runner/runner -loadtest -script=$lua_script \
        | tee -a benchmark.txt 1>/dev/null
    local exit_code2="${PIPESTATUS[0]}"
    if [ "$exit_code2" -ne "0" ]; then
        cat benchmark.txt
        exit $exit_code2
    fi
}

function go_bench() {
    local sha=$1

    rm -f "bench_$sha.txt"
    touch "bench_$sha.txt"
    for i in `seq 5`; do
        banner "running bench program $sha iter $i"
        make bench | tee -a "bench_$sha.txt"
    done
}

# Go to root dir
while [ "$(basename $PWD)" != "zanzibar" ]; do
    cd ..
done

rm -f benchmark.txt
touch benchmark.txt

if [ "$BENCHMARK_FILE" == "bench" ]; then
    run $CURRENT_SHA
    go_bench $CURRENT_SHA
    run $OTHER_SHA
    go_bench $OTHER_SHA

    go get -u golang.org/x/tools/cmd/benchcmp
    echo "Comparing old $CURRENT_SHA to new $OTHER_SHA"
    benchcmp -best "bench_$CURRENT_SHA.txt" "bench_$OTHER_SHA.txt"

else
    run $CURRENT_SHA
    benchmark_runner $CURRENT_SHA $BENCHMARK_FILE
    run $OTHER_SHA
    benchmark_runner $OTHER_SHA $BENCHMARK_FILE

    cat benchmark.txt
    rm -f benchmark.txt
fi


git checkout $CURRENT_SHA
rm -f stdout.txt
rm -f stderr.txt
