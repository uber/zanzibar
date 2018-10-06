#!/usr/bin/env bash
set -e
set -o pipefail

if [ -z "$1" ]; then
	echo "build dir argument (\$1) is missing"
	exit 1
fi

if [ -z "$2" ]; then
	echo "config dir argument (\$2) is missing"
	exit 1
fi
if [ -z "$3" ]; then
	echo "annotation prefix argument (\$3) is missing"
	exit 1
fi

BUILD_DIR="$1"
CONFIG_DIR="$2"
ANNOPREFIX="$3"
IDL_DIR="${4:-$CONFIG_DIR/idl}"

if [ -z "$4" ]; then
	THRIFTRW_SRCS="$(find "$IDL_DIR" -name '*.thrift')"
else
	THRIFTRW_SRCS="$4"
fi
THRIFTRW_SRCS="$(echo "$THRIFTRW_SRCS" | xargs -n1 | sort | uniq)"

DIRNAME="$(dirname "$0")"
EASY_JSON_RAW_DIR="$DIRNAME/../../scripts/easy_json"
EASY_JSON_DIR="$(cd "$EASY_JSON_RAW_DIR";pwd)"
EASY_JSON_FILE="$EASY_JSON_DIR/easy_json.go"
EASY_JSON_BINARY="$EASY_JSON_DIR/easy_json"
RESOLVE_THRIFT_FILE="$DIRNAME/../../scripts/resolve_thrift/main.go"
RESOLVE_THRIFT_BINARY="$DIRNAME/../../scripts/resolve_thrift/resolve_thrift"
RESOLVE_I64_FILE="$DIRNAME/../../scripts/resolve_i64/main.go"
RESOLVE_I64_BINARY="$DIRNAME/../../scripts/resolve_i64/resolve_i64"

if [ -d "$DIRNAME/../../vendor" ]; then
	THRIFTRW_RAW_DIR="$DIRNAME/../../vendor/go.uber.org/thriftrw"
	THRIFTRW_DIR="$(cd "$THRIFTRW_RAW_DIR";pwd)"
	THRIFTRW_MAIN_FILE="$THRIFTRW_DIR/main.go"
	THRIFTRW_BINARY="$THRIFTRW_DIR/thriftrw"
else
	THRIFTRW_RAW_DIR="$DIRNAME/../../../../../go.uber.org/thriftrw"
	THRIFTRW_DIR="$(cd "$THRIFTRW_RAW_DIR";pwd)"
	THRIFTRW_MAIN_FILE="$THRIFTRW_DIR/main.go"
	THRIFTRW_BINARY="$THRIFTRW_DIR/thriftrw"
fi

start=$(date +%s)
echo "$start" > .TMP_ZANZIBAR_TIMESTAMP_FILE.txt

go build -o "$THRIFTRW_BINARY" "$THRIFTRW_MAIN_FILE"
end=$(date +%s)
runtime=$((end-start))
echo "Compiled thriftrw : +$runtime"

echo "Generating Go code from Thrift files"
mkdir -p "$BUILD_DIR/gen-code"
for tfile in ${THRIFTRW_SRCS}; do
	"$THRIFTRW_BINARY" --out="$BUILD_DIR/gen-code" \
		--no-embed-idl \
		--thrift-root="$IDL_DIR" "$tfile"
done
gofmt -w "$BUILD_DIR/gen-code/"

end=$(date +%s)
runtime=$((end-start))
echo "Generated structs : +$runtime"

go build -o "$EASY_JSON_BINARY" "$EASY_JSON_FILE"
end=$(date +%s)
runtime=$((end-start))
echo "Compiled easyjson : +$runtime"

go build -o "$RESOLVE_THRIFT_BINARY" "$RESOLVE_THRIFT_FILE"
go build -o "$RESOLVE_I64_BINARY" "$RESOLVE_I64_FILE"

# find the modules that actually need JSON (un)marshallers
ABS_IDL_DIR="$(cd "$IDL_DIR" && pwd)"
ABS_GENCODE_DIR="$(cd "$BUILD_DIR" && pwd)/$(basename "$BUILD_DIR/gen-code")"
target_dirs=""
found_thrifts=""
config_files=$(find "$CONFIG_DIR" -name "*-config.json" -o -name "*-config.yaml" | sort)
for config_file in ${config_files}; do
	if [[ ${config_file} == "./vendor"* ]]; then
		continue
	fi

	processor="yq"
	if [[ $config_file == *.json ]]; then
		processor="jq"
	fi

	module_type=$($processor -r .type "$config_file")
	[[ ${module_type} != "http" ]] && continue
	dir=$(dirname "$config_file")
	yaml_files=$(find "$dir" -name "*.json" -o -name "*.yaml")
	for yaml_file in ${yaml_files}; do
		processor="yq"
		if [[ $yaml_file == *.json ]]; then
			processor="jq"
		fi

		thrift_file=$($processor -r '.. | .thriftFile? | select(strings | endswith(".thrift"))' "$yaml_file")
		[[ -z ${thrift_file} ]] && continue
		[[ ${found_thrifts} == *${thrift_file}* ]] && continue
		found_thrifts+=" $thrift_file"

		thrift_file="$IDL_DIR/$thrift_file"
		gen_code_dir=$(
		"$RESOLVE_THRIFT_BINARY" "$thrift_file" "$ANNOPREFIX" | \
			sed "s|$ABS_IDL_DIR\/\(.*\)\/.*.thrift|$ABS_GENCODE_DIR/\1|" | \
			sort | uniq | xargs
		)
		"$RESOLVE_I64_BINARY" "$thrift_file"
		target_dirs+=" $gen_code_dir"
	done
done
target_dirs=($(echo "$target_dirs" | tr ' ' '\n' | sort | uniq))

echo "Generating JSON Marshal/Unmarshal"
thriftrw_gofiles=(
$(find "${target_dirs[@]}" -name "*.go" | \
	grep -v "versioncheck.go" | \
	grep -v "easyjson.go" | sort)
)
"$EASY_JSON_BINARY" -all -- "${thriftrw_gofiles[@]}"

end=$(date +%s)
runtime=$((end-start))
echo "Generated structs : +$runtime"
