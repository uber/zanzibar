#!/usr/bin/env bash
set -e

rm -f ./coverage/cover-temp.out
gocovmerge ./coverage/cover-*.out > ./coverage/cover-temp.out

cat ./coverage/cover-temp.out | \
    grep -v "_easyjson.go" | \
    grep -v "gen-code" | \
    sed "s/github.com\/uber\/zanzibar/./" > \
    ./coverage/cover.out

rm ./coverage/cover-temp.out
