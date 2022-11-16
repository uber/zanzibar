#!/usr/bin/env bash
set -e

rm -f test.out
rm -f fail.out
touch test.out
rm -f coverage.tmp
mkdir -p ./coverage
rm -f ./coverage/*.out

start=`date +%s`

if [ $# -eq 0 ]
then
	FILES=$(go list -e ./... | grep -v "vendor" | \
		grep -v "workspace" | \
		grep "test\|examples\|runtime\|codegen\|repository")
elif [ $# -eq 1 ]
then
	FILES=$(go list -e ./... | grep -v "vendor" | \
		grep -v "workspace" | grep "$1")
else
	FILES=$@
fi

rm -f ./test/.cached_binary_test_info.json

REAL_TEST_FILES=$(git grep -l 'func Test.' | \
	xargs -I{} dirname github.com/uber/zanzibar/{} | sort | uniq)

FILES_ARR=($FILES)

echo "Starting coverage tests:"

for file in "${FILES_ARR[@]}"; do

	if ! grep -q $file <<<$REAL_TEST_FILES; then
		continue
	fi

	RAND=$(hexdump -n 8 -v -e '/1 "%02X"' /dev/urandom)
	COVERNAME="./coverage/cover-unit-$RAND.out"

	relativeName=$(echo $file | sed s#github.com/uber/zanzibar#.#)

    # TODO: need better solution for coverage from different package
    if [[ "$relativeName" == *"test/clients"* ]] || [[ "$relativeName" == *"test/endpoints"* ]]; then
        coverpkg=$(echo $file | sed s#zanzibar/test/#zanzibar/examples/example-gateway/build/#)
        COVER_ON=1 ZANZIBAR_CACHE=1 go test -tags mock \
            -cover -coverprofile coverage.tmp -coverpkg $coverpkg $file 2>&1 | \
		    tee test.tmp.out >>test.out && \
		    mv coverage.tmp "$COVERNAME" 2>/dev/null || true
    else
	    COVER_ON=1 ZANZIBAR_CACHE=1 go test -tags mock \
		    -cover -coverprofile coverage.tmp $file 2>&1 | \
		    tee test.tmp.out >>test.out && \
		    mv coverage.tmp "$COVERNAME" 2>/dev/null || true
    fi

	# cat test.tmp.out | grep -E '[0-9]s' || true
	rm test.tmp.out

	end=`date +%s`
	runtime=$((end-start))
	printf "Finished coverage test  :  %-60s  :  +%3d \n" $relativeName $runtime
done

echo ""
echo "      --------------------        "
echo ""

rm -f ./test/.cached_binary_test_info.json

cat test.out | grep -v "warning: no packages" | grep -v "\[no test files\]" || true
rm -f coverage.tmp
grep "FAIL" test.out | tee -a fail.out
[ ! -s fail.out ]

go get github.com/wadey/gocovmerge
rm -f ./coverage/cover-temp.out
gocovmerge ./coverage/cover-*.out > ./coverage/cover-temp.out

cat ./coverage/cover-temp.out | \
	grep -v "_easyjson.go" | \
	grep -v "gen-code" | \
	grep -v "codegen/runner" | \
	sed "s/github.com\/uber\/zanzibar/./" > \
	./coverage/cover.out

rm ./coverage/cover-temp.out

echo ""
echo "      --------------------        "
echo ""

end=`date +%s`
runtime=$((end-start))
echo "Finished concating coverage : +$runtime"

make generate-istanbul-json

end=`date +%s`
runtime=$((end-start))
echo "Finished generating istanbul json : +$runtime"

ls ./node_modules/.bin/istanbul 2>/dev/null || npm i istanbul
./node_modules/.bin/istanbul report --root ./coverage \
	--include "**/istanbul.json" text
./node_modules/.bin/istanbul report --root ./coverage \
	--include "**/istanbul.json" html
./node_modules/.bin/istanbul report --root ./coverage \
	--include "**/istanbul.json" lcovonly

end=`date +%s`
runtime=$((end-start))
echo "Finished building istanbul reports : +$runtime"


# TODO: (lu) remove runtime/tchannel exclusion
cat ./coverage/istanbul.json | jq '[
	. |
	to_entries |
	.[] |
	select(.key | contains("runtime")) |
	select(.key | contains("runtime/tchannel") | not) |
	select(.key | contains("runtime/middlewares/logger") | not) |
	select(.key | contains("runtime/gateway") | not)
] | from_entries' > ./coverage/istanbul-runtime.json

echo "Checking code coverage for runtime folder"
./node_modules/.bin/istanbul check-coverage --statements 99.51 \
	./coverage/istanbul-runtime.json

cat ./coverage/istanbul.json | jq '[
	. |
	to_entries |
	.[] |
	select(.key | contains("codegen/type_converter"))
] | from_entries' > ./coverage/istanbul-codegen.json

echo "Checking code coverage for codegen folder"
./node_modules/.bin/istanbul check-coverage --statements 100 \
	./coverage/istanbul-codegen.json
