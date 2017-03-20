

PKGS = $(shell glide novendor)
PKG_FILES = benchmarks codegen examples runtime test

COVER_PKGS = $(shell glide novendor | grep -v "test/..." | \
	grep -v "main/..." | grep -v "benchmarks/..." | \
	awk -vORS=, '{ print $1 }' | sed 's/,$$/\n/')

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

FILTER_LINT := grep -v -e "vendor/" -e "third_party/" -e "gen-code/"

# list all executables
PROGS = examples/example-gateway/build/example-gateway \
	benchmarks/benchserver/benchserver \
	benchmarks/runner/runner

.PHONY: check-licence
check-licence:
	ls ./node_modules/.bin/uber-licence 2>/dev/null || npm i uber-licence
	./node_modules/.bin/uber-licence --dry --file '*.go' --dir '!vendor' --dir '!examples' --dir '!.tmp_gen'

.PHONY: fix-licence
fix-licence:
	ls ./node_modules/.bin/uber-licence 2>/dev/null || npm i uber-licence
	./node_modules/.bin/uber-licence --file '*.go' --dir '!vendor' --dir '!examples' --dir '!.tmp_gen'

.PHONY: install
install:
	@echo "Mounting git pre-push hook"
	cp .git-pre-push-hook .git/hooks/pre-push
	@echo "Installing Glide and locked dependencies..."
	glide --version || go get -u -f github.com/Masterminds/glide
	glide install

.PHONY: lint
lint: check-licence
	@rm -f lint.log
	@echo "Checking formatting..."
	@gofmt -d -s $(PKG_FILES) 2>&1 | $(FILTER_LINT) | tee -a lint.log
	@echo "Installing test dependencies for vet..."
	@go test -i $(PKGS)
	@echo "Checking printf statements..."
	@git grep -E 'Fprintf\(os.Std(err|out)' | $(FILTER_LINT) | tee -a lint.log
	@echo "Checking vet..."
	@$(foreach dir,$(PKG_FILES),go tool vet $(VET_RULES) $(dir) 2>&1 | $(FILTER_LINT) | tee -a lint.log;)
	@echo "Checking lint..."
	@go get github.com/golang/lint/golint
	@$(foreach dir,$(PKGS),golint $(dir) 2>&1 | $(FILTER_LINT) | tee -a lint.log;)
	@echo "Checking errcheck..."
	@go get github.com/kisielk/errcheck
	@errcheck $(PKGS) 2>&1 | $(FILTER_LINT) | tee -a lint.log
	@echo "Checking staticcheck..."
	@go get honnef.co/go/tools/cmd/staticcheck
	@staticcheck $(PKGS) 2>&1 | $(FILTER_LINT) | tee -a lint.log
	@echo "Checking for unresolved FIXMEs..."
	@git grep -i fixme | grep -v -e vendor -e Makefile | $(FILTER_LINT) | tee -a lint.log
	@[ ! -s lint.log ]

# .PHONY: verify_deps
# verify_deps:
# 	@rm -f deps.log
# 	@echo "Verifying dependency conflicts"
# 	@glide -q update 2>&1 | tee -a deps.log
# 	@[ ! -s deps.log ]

.PHONY: generate
generate:
	@goimports -h 2>/dev/null || go get golang.org/x/tools/cmd/goimports
	@bash ./scripts/generate.sh

.PHONY: test-all
test-all: test cover fast-bench bins
	./benchmarks/runner/runner -loadtest

.PHONY: test
test: generate lint
	go test ./examples/example-gateway/... 1>/dev/null
	go test ./test/... | grep -v '\[no test files\]'
	echo "<coverage />" > ./coverage/cobertura-coverage.xml

.PHONY: fast-bench
fast-bench:
	time -p sh -c "go test -run _NONE_ -bench . -benchmem -benchtime 1s -cpu 2 ./test/... | grep -v '^ok ' | grep -v '\[no test files\]' | grep -v '^PASS'"

.PHONY: bench
bench:
	time -p sh -c "go test -run _NONE_ -bench . -benchmem -benchtime 7s -cpu 2 ./test/... | grep -v '^ok ' | grep -v '\[no test files\]' | grep -v '^PASS'"

$(PROGS): $(GO_FILES)
	@echo Building $@
	go build -o $@ $(dir ./$@)

.PHONY: bins
bins: generate $(PROGS)

.PHONY: run
run: examples/example-gateway/build/example-gateway
	cd ./examples/example-gateway; \
		ENVIRONMENT=production \
		CONFIG_DIR=./config \
		./build/example-gateway

.PHONY: go-docs
go-docs:
	godoc -http=:6060

.PHONY: clean-easyjson
clean-easyjson:
	find . -name "*_easyjson.go" -delete
	find . -name "*.bak" -delete
	find . -name "easyjson-bootstrap*.go" -delete

.PHONY: kill-dead-benchmarks
kill-dead-benchmarks:
	ps aux | grep gocode | awk '{ print $$2 }' | xargs kill

.PHONY: clean-cover
clean-cover:
	@rm -f ./coverage/cover.out
	@rm -f ./coverage/cover-*.out
	@rm -f ./coverage/index.html

.PHONY: cover
cover: clean-cover
	bash ./scripts/cover.sh

.PHONY: generate-istanbul-json
generate-istanbul-json:
	@go get github.com/axw/gocov/gocov
	@gocov convert ./coverage/cover.out > coverage/gocov.json
	@node ./scripts/gocov-to-istanbul-coverage.js ./coverage/gocov.json \
		> coverage/istanbul.json

.PHONY: view-istanbul
view-istanbul: generate-istanbul-json
	istanbul report --root ./coverage --include "**/istanbul.json" html
	@if [ $$(which xdg-open) ]; then \
		xdg-open coverage/index.html; \
	else \
		open coverage/index.html; \
	fi

.PHONY: view-gocov
view-gocov:
	@go get github.com/axw/gocov/gocov
	@go get -u gopkg.in/matm/v1/gocov-html
	@gocov convert ./coverage/cover.out > coverage/gocov.json
	@cat coverage/gocov.json | gocov-html > ./coverage/index.html
	@if [ $$(which xdg-open) ]; then \
		xdg-open coverage/index.html; \
	else \
		open coverage/index.html; \
	fi


.PHONY: view-cover
view-cover:
	go tool cover -html=./coverage/cover.out

.PHONY: clean-vendor
clean-vendor:
	rm -rf ./vendor

.PHONY: clean
clean: clean-easyjson clean-cover clean-vendor
	go clean
	rm -f $(PROGS)

.PHONY: tag-build
tag-build:
	date +%Y-%m-%d-%H-%M-%S >VERSION
	git add VERSION
	cat VERSION | xargs -I{} git commit -m "build: {}"
	cat VERSION | xargs -I{} git tag -m "build: {}" "build-{}"
