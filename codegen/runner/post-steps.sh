#!/usr/bin/env bash
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

#DIRNAME="$(dirname "$0")"
# The assumption here is that either Mockery package resides in Zanzibar's vendor dir
# or Zanzibar itself is in a vendor dir, in which Mockery also resides.
# For the second case, it's also assumed that the vendor directory is flattened as Glide does.
#if [ -d "$DIRNAME/../../vendor" ]; then
#	MOCKERY_RAW_DIR="$DIRNAME/../../vendor/github.com/vektra/mockery"
#else
#	MOCKERY_RAW_DIR="$DIRNAME/../../../../../github.com/vektra/mockery"
#fi
#MOCKERY_DIR="$(cd "$MOCKERY_RAW_DIR";pwd)"
#MOCKERY_MAIN_FILE="$MOCKERY_DIR/cmd/mockery/mockery.go"
#MOCKERY_BINARY="$MOCKERY_DIR/cmd/mockery/mockery"
#
#go build -o "$MOCKERY_BINARY" "$MOCKERY_MAIN_FILE"
#end=$(date +%s)
#runtime=$((end-start))
#echo "Compiled Mockery: +$runtime"
#
#"$MOCKERY_BINARY" -name="^Client$" -dir="$BUILD_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null
#"$MOCKERY_BINARY" -name="^Client$" -dir="$CONFIG_DIR/clients" -inpkg -case=underscore -recursive -note="+build mock" > /dev/null

if ! [ -x "$(command -v mockgen)" ]; then
	go get -u github.com/golang/mock/gomock
	go get -u github.com/golang/mock/mockgen
	end=$(date +%s)
	runtime=$((end-start))
	echo "Installed mockgen: +"$runtime""
fi

genmock() {
	for d in "$1/clients/"*/; do
		abs_path="$(cd "$d"; pwd)"
		import_path=${abs_path#$GOPATH/src/}
		dest_dir="$abs_path/mock_client"
		mkdir -p "$dest_dir"
		# TODO: need a better way to deal with errors, silencing for now
		mockgen -destination="$dest_dir/mock_client.go" -package="clientmock" "$import_path" "Client" 2>/dev/null
		if [ $? -ne 0 ]; then
		    rm -r "$dest_dir"
		else
			end=$(date +%s)
			runtime=$((end-start))
			printf "Generated mock for %-15s: +"$runtime"\n" "$(basename "$import_path")"
		fi
	done
}

# TODO: reflect mode is quite slow, may need do some caching here
genmock "$BUILD_DIR"
genmock "$CONFIG_DIR"
