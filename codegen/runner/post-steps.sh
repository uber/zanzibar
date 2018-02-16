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
# The assumption here is that either Mockery package resides in Zanzibar's vendor dir
# or Zanzibar itself is in a vendor dir, in which Mockery also resides.
# For the second case, it's also assumed that the vendor directory is flattened as Glide does.
if [ -d "$DIRNAME/../../vendor" ]; then
	MOCKERY_RAW_DIR="$DIRNAME/../../vendor/github.com/vektra/mockery"
else
	MOCKERY_RAW_DIR="$DIRNAME/../../../../../github.com/vektra/mockery"
fi
MOCKERY_DIR="$(cd "$MOCKERY_RAW_DIR";pwd)"
MOCKERY_MAIN_FILE="$MOCKERY_DIR/cmd/mockery/mockery.go"
MOCKERY_BINARY="$MOCKERY_DIR/cmd/mockery/mockery"

go build -o "$MOCKERY_BINARY" "$MOCKERY_MAIN_FILE"
end=$(date +%s)
runtime=$((end-start))
echo "Compiled Mockery: +$runtime"

"$MOCKERY_BINARY" -name="^Client$" -dir="$BUILD_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null
"$MOCKERY_BINARY" -name="^Client$" -dir="$CONFIG_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null
end=$(date +%s)
runtime=$((end-start))
echo "Generated mock clients: +$runtime"
