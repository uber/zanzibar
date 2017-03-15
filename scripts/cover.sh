#!/usr/bin/env bash
set -e

rm -f test.out
rm -f fail.out
touch test.out
rm -f coverage.tmp
mkdir -p ./coverage
rm -f ./coverage/*.out

COVER_PKGS=$(glide novendor | grep -v "test/..." | \
	grep -v "main/..." | grep -v "benchmarks/..." | \
	awk -v ORS=, '{ print $1 }' | sed 's/,$/\n/')

FILES=$(go list ./... | grep -v "vendor" | grep "test")
FILES_ARR=($FILES)

for file in "${FILES_ARR[@]}"; do
	RAND=$(hexdump -n 8 -v -e '/1 "%02X"' /dev/urandom)
	COVER_ON=1 go test -cover -coverpkg $COVER_PKGS \
		-coverprofile coverage.tmp $file >>test.out 2>&1 && \
		mv coverage.tmp "./coverage/cover-unit-$RAND.out" 2>/dev/null || true
done

cat test.out | grep -v "warning: no packages" | grep -v "\[no test files\]" || true
rm -f coverage.tmp
grep "FAIL" test.out | tee -a fail.out
[ ! -s fail.out ]

go get github.com/wadey/gocovmerge
bash ./scripts/concat-coverage.sh
echo "\nOutputting coverage info... \n"
make generate-istanbul-json

ls ./node_modules/.bin/instanbul 2>/dev/null || npm i istanbul
./node_modules/.bin/istanbul report --root ./coverage \
	--include "**/istanbul.json" text
