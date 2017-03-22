#!/usr/bin/env bash
set -e

rm -f test.out
rm -f fail.out
touch test.out
rm -f coverage.tmp
mkdir -p ./coverage
rm -f ./coverage/*.out

start=`date +%s`
COVER_PKGS=$(glide novendor | grep -v "test/..." | \
	grep -v "main/..." | grep -v "benchmarks/..." | \
	awk -v ORS=, '{ print $1 }' | sed $'s/,$/\\\n/')

if [ $# -eq 0 ]
then
    FILES=$(go list ./... | grep -v "vendor" | grep "test\|examples")
else
    FILES="$@"
fi

FILES_ARR=($FILES)

for file in "${FILES_ARR[@]}"; do
	RAND=$(hexdump -n 8 -v -e '/1 "%02X"' /dev/urandom)
	echo "Running coverage test : $file"
	COVER_ON=1 go test -cover -coverpkg $COVER_PKGS \
		-coverprofile coverage.tmp $file >>test.out 2>&1 && \
		mv coverage.tmp "./coverage/cover-unit-$RAND.out" 2>/dev/null || true
	end=`date +%s`
	runtime=$((end-start))
	echo "Finished coverage test : $file : +$runtime"
done

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
    sed "s/github.com\/uber\/zanzibar/./" > \
    ./coverage/cover.out

rm ./coverage/cover-temp.out


end=`date +%s`
runtime=$((end-start))
echo "Finished concatting coverage : +$runtime"

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
