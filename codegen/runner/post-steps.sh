#!/usr/bin/env bash
set -e
set -o pipefail

if [ -z "$1" ]; then
	echo "prefix argument (\$1) is missing"
	exit 1
fi

if [ -z "$2" ]; then
	echo "config dir argument (\$2) is missing"
	exit 1
fi

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)

BUILD_DIR="$1"
CONFIG_DIR="$2"

DIRNAME="$(dirname "$0")"
if [ -d "$DIRNAME/../../vendor" ]; then
	MOCKERY_RAW_DIR="$DIRNAME/../../vendor/github.com/vektra/mockery"
else
	MOCKERY_RAW_DIR="$DIRNAME/../../../../../github.com/vektra/mockery"
fi
MOCKERY_DIR="$(cd "$MOCKERY_RAW_DIR";pwd)"
MOCKERY_MAIN_FILE="$MOCKERY_DIR/cmd/mockery/mockery.go"
MOCKERY_BINARY="$MOCKERY_DIR/cmd/mockery/mockery"

go build -o "$MOCKERY_BINARY" "$MOCKERY_MAIN_FILE"

"$MOCKERY_BINARY" -name="^Client$" -dir="$BUILD_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null
"$MOCKERY_BINARY" -name="^Client$" -dir="$CONFIG_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null
end=$(date +%s)
runtime=$((end-start))
echo "Generated mock clients: +$runtime"
