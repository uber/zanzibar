#!/usr/bin/env bash
set -e
set -o pipefail

usage() {
	echo "Usage: $0 [config-dir] [build-dir] "
}

if [ -z "$1" ]; then
	echo "config dir argument (\$1) is missing"
	usage
	exit 1
fi

if [ -z "$2" ]; then
	echo "build dir argument (\$2) is missing"
	usage
	exit 1
fi

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)

CONFIG_DIR="$1"
BUILD_DIR="$2"

DIRNAME="$(dirname "$0")"
