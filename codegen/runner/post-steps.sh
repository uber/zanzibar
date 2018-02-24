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
# The assumption here is that either mockgen package resides in Zanzibar's vendor dir
# or Zanzibar itself is in a vendor dir, in which mockgen also resides.
# For the second case, it's also assumed that the vendor directory is flattened as Glide does.
if [ -d "$DIRNAME/../../vendor" ]; then
	MOCKGEN_RAW_DIR="$DIRNAME/../../vendor/github.com/golang/mock"
else
	MOCKGEN_RAW_DIR="$DIRNAME/../../../../../github.com/golang/mock"
fi
MOCKGEN_DIR="$(cd "$MOCKGEN_RAW_DIR";pwd)/mockgen"
MOCKGEN_BINARY="$MOCKGEN_DIR/mockgen"

if ! [ -x "$(command -v $MOCKGEN_BINARY)" ]; then
	go build -o "$MOCKGEN_BINARY" "$MOCKGEN_DIR"/*.go
	end=$(date +%s)
	runtime=$((end-start))
	echo "Compiled mockgen: +$runtime"
fi

GEN_ERR="/tmp/zanzibar-mockgen.err"
genmock() {
	for d in "$1/clients/"*/; do
		abs_path="$(cd "$d"; pwd)"
		import_path=${abs_path#$GOPATH/src/}
		dest_dir="$abs_path/mock-client"
		mkdir -p "$dest_dir"
		# TODO: need a better way to deal with errors, silencing for now
		"$MOCKGEN_BINARY" -destination="$dest_dir/mock_client.go" -package="clientmock" "$import_path" "Client" 2>>"$GEN_ERR"
		if [ $? -ne 0 ]; then
		    rm -r "$dest_dir"
		else
			end=$(date +%s)
			runtime=$((end-start))
			printf "Generated mock for %-25s: +"$runtime"\n" "$(basename "$import_path")"
		fi
	done
}

# TODO: reflect mode is quite slow, may need do some caching here
set +e
rm -f "$GEN_ERR"
genmock "$BUILD_DIR"
genmock "$CONFIG_DIR"
