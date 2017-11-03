PKGS = $(shell glide novendor | grep -v "workspace/...")
PKG_FILES = benchmarks codegen examples runtime test backend

COVER_PKGS = $(shell glide novendor | grep -v "test/..." | \
	grep -v "main/..." | grep -v "benchmarks/..." | \
	grep -v "workspace/..." | awk -vORS=, '{ print $1 }' | sed 's/,$$/\n/')

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' -o -path './workspace' ')' -prune -o -name '*.go' -print | cut -b3-)

FILTER_LINT := grep -v -e "vendor/" -e "third_party/" -e "gen-code/" -e "config/" -e "codegen/templates/" -e "codegen/template_bundle/"

# list all executables
PROGS = benchmarks/benchserver/benchserver \
	benchmarks/runner/runner

EXAMPLE_BASE_DIR = examples/example-gateway/
EXAMPLE_SERVICES_DIR = $(EXAMPLE_BASE_DIR)build/services/
EXAMPLE_SERVICES = $(sort $(dir $(wildcard $(EXAMPLE_SERVICES_DIR)*/)))

.PHONY: install
install:
	@echo "Mounting git pre-push hook"
	cp .git-pre-push-hook .git/hooks/pre-push
	@echo "Installing Glide and locked dependencies..."
	glide --version || go get -u -f github.com/Masterminds/glide
	glide install

.PHONY: check-licence
check-licence:
	@echo "Checking uber-licence..."
	@ls ./node_modules/.bin/uber-licence >/dev/null 2>&1 || npm i uber-licence
	@./node_modules/.bin/uber-licence --dry --file '*.go' --dir '!workspace' --dir '!vendor' --dir '!examples' --dir '!.tmp_gen' --dir '!template_bundle'

.PHONY: fix-licence
fix-licence:
	@ls ./node_modules/.bin/uber-licence >/dev/null 2>&1 || npm i uber-licence
	./node_modules/.bin/uber-licence --file '*.go' --dir '!vendor' --dir '!workspace' --dir '!examples' --dir '!.tmp_gen' --dir '!template_bundle'

.PHONY: eclint-check
eclint-check:
	@echo "Checking eclint..."
	@ls ./node_modules/.bin/eclint >/dev/null 2>&1 || npm i eclint@v1.1.5
	@./node_modules/.bin/eclint check "./{codegen,examples,backend}/**/*.{json,tmpl}"

.PHONY: eclint-fix
eclint-fix:
	@ls ./node_modules/.bin/eclint >/dev/null 2>&1 || npm i eclint@v1.1.5
	./node_modules/.bin/eclint fix "./{codegen,examples,backend}/**/*.{json,tmpl}"

.PHONY: spell-check
spell-check:
	@go get github.com/client9/misspell/cmd/misspell
	@misspell $(PKG_FILES)

.PHONY: spell-fix
spell-fix:
	@go get github.com/client9/misspell/cmd/misspell
	@misspell -w $(PKG_FILES)

.PHONY: cyclo-check
cyclo-check:
	@go get github.com/fzipp/gocyclo
	@gocyclo -over 15 $(filter-out examples ,$(PKG_FILES))

.PHONY: lint
lint: check-licence eclint-check
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
	@go get -u github.com/jteeuwen/go-bindata/...
	@ls ./node_modules/.bin/uber-licence >/dev/null 2>&1 || npm i uber-licence
	@chmod 644 ./codegen/templates/*.tmpl
	@chmod 644 ./config/production.json
	@go-bindata -pkg config -nocompress -modtime 1 -prefix config -o config/production.json.go config/production.json
	@./node_modules/.bin/uber-licence --file "production.json.go" --dir "config" > /dev/null
	@go-bindata -pkg templates -nocompress -modtime 1 -prefix codegen/templates -o codegen/template_bundle/template_files.go codegen/templates/...
	@gofmt -w -e -s "codegen/template_bundle/template_files.go"
	@goimports -h 2>/dev/null || go get golang.org/x/tools/cmd/goimports
	@bash ./scripts/generate.sh

.PHONY: check-generate
check-generate:
	@rm -f git-status.log
	rm -rf ./examples/example-gateway/build
	make generate
	git status --porcelain > git-status.log
	@[ ! -s git-status.log ] || ( cat git-status.log ; git --no-pager diff ; [ ! -s git-status.log ] );

.PHONY: test-all
test-all: 
	$(MAKE) jenkins
	$(MAKE) install
	$(MAKE) cover
	$(MAKE) fast-bench 
	$(MAKE) bins 
	$(MAKE) install-wrk
	$(MAKE) test-benchmark-runner

.PHONY: test-benchmark-runner
test-benchmark-runner:
	PATH=$$PATH:$$PWD/vendor/wrk ./benchmarks/runner/runner -loadtest

.PHONY: install-wrk
install-wrk:
	ls ./vendor/wrk/wrk 2>/dev/null || git clone \
		https://github.com/wg/wrk.git ./vendor/wrk
	cd ./vendor/wrk ; (ls ./wrk 2>/dev/null || make >install_wrk.log)

.PHONY: test
test: generate lint
	@make test-only

.PHONY: test-update
test-update:
	go test ./codegen/ ./backend/... -update

.PHONY: test-only
test-only:
	@rm -f ./test/.cached_binary_test_info.json
	@echo "Running all tests..."
	@ZANZIBAR_CACHE=1 go test ./test/health_test.go # preload the binary cache
	@ZANZIBAR_CACHE=1 go test \
		./examples/example-gateway/... \
		./codegen/... \
		./backend/... \
		./runtime/... \
		./test/... | \
		grep -v '\[no test files\]'
	@rm -f ./test/.cached_binary_test_info.json
	@echo "<coverage />" > ./coverage/cobertura-coverage.xml

.PHONY: travis-coveralls
travis-coveralls:
	ls ./node_modules/coveralls/bin/coveralls.js 2>/dev/null || \
		npm i coveralls
	cat ./coverage/lcov.info | ./node_modules/coveralls/bin/coveralls.js

.PHONY: fast-bench
fast-bench:
	@rm -f bench.log bench-fail.log
	time -p sh -c "go test -run _NONE_ -bench . -benchmem -benchtime 1s -cpu 2 ./test/... | grep -v '^ok ' | grep -v '\[no test files\]' | grep -v '^PASS'  | tee -a bench.log"
	@cat bench.log | grep "FAIL" | tee -a bench-fail.log
	@[ ! -s bench-fail.log ]

.PHONY: bench
bench:
	time -p sh -c "go test -run _NONE_ -bench . -benchmem -benchtime 7s -cpu 2 ./test/... | grep -v '^ok ' | grep -v '\[no test files\]' | grep -v '^PASS'"

$(PROGS): $(GO_FILES)
	@echo Building $@
	go build -o $@ $(dir ./$@)

# These dirs are generated by `make generate`, if `make generate` is ran after any $(GO_FILES) is modified,
# these dirs will have a later timestamp than the modified $(GO_FILES), therefore make believes they are up to date.
# These needs to be phony despite they are actual files, as the target is not generating them but the binaries instead.
.PHONY: $(EXAMPLE_SERVICES)
$(EXAMPLE_SERVICES): $(GO_FILES)
	@echo Building $@
	go build -o "$@../../../bin/$(shell basename $@)" ./$@main

run-%: $(EXAMPLE_SERVICES_DIR)%/
	cd "$(EXAMPLE_BASE_DIR)"; \
		UBER_ENVIRONMENT=production \
		CONFIG_DIR=./config \
		./bin/$* --config="config/production.json;"


.PHONY: bins
bins: generate $(PROGS) $(EXAMPLE_SERVICES)

.PHONY: run
run: run-example-gateway

.PHONY: go-docs
go-docs:
	godoc -http=:6060

.PHONY: clean-easyjson
clean-easyjson:
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
	@goimports -h 2>/dev/null || go get golang.org/x/tools/cmd/goimports
	@bash ./scripts/cover.sh

.PHONY: generate-istanbul-json
generate-istanbul-json:
	@go get github.com/axw/gocov/gocov
	@gocov convert ./coverage/cover.out > coverage/gocov.json
	@node ./scripts/gocov-to-istanbul-coverage.js ./coverage/gocov.json \
		> coverage/istanbul.json

.PHONY: view-istanbul
view-istanbul: generate-istanbul-json
	./node_modules/.bin/istanbul report --root ./coverage \
		--include "**/istanbul.json" html
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

.PHONY: jenkins-install
jenkins-install:
	PWD=$(pwd)
	@rm -rf ./vendor/
	@rm -rf ./workspace/
	@mkdir -p ./workspace/src/github.com/uber/
	@ln -s $(PWD) workspace/src/github.com/uber/zanzibar
	cd workspace/src/github.com/uber/zanzibar && \
		GOPATH=$(PWD)/workspace \
		PATH=$(PWD)/workspace/bin:$(PATH) \
		make install

.PHONY: jenkins-test
jenkins-test:
	PWD=$(pwd)
	cd workspace/src/github.com/uber/zanzibar && \
		GOPATH=$(PWD)/workspace \
		PATH=$(PWD)/workspace/bin:$(PATH) \
		make check-generate
	cd workspace/src/github.com/uber/zanzibar && \
		GOPATH=$(PWD)/workspace \
		PATH=$(PWD)/workspace/bin:$(PATH) \
		make lint
	cd workspace/src/github.com/uber/zanzibar && \
		GOPATH=$(PWD)/workspace \
		PATH=$(PWD)/workspace/bin:$(PATH) \
		make test-only

.PHONY: jenkins
jenkins:
	$(MAKE) jenkins-install
	$(MAKE) jenkins-test
