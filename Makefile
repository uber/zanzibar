PKGS = $(shell glide novendor | grep -v "workspace/...")
PKG_FILES = benchmarks codegen examples runtime test
PWD = $(shell pwd)

COVER_PKGS = $(shell glide novendor | grep -v "test/..." | \
	grep -v "main/..." | grep -v "benchmarks/..." | \
	grep -v "workspace/..." | awk -vORS=, '{ print $1 }' | sed 's/,$$/\n/')

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' -o -path './workspace' ')' -prune -o -name '*.go' -print | cut -b3-)

FILTER_LINT := grep -v -e "vendor/" -e "third_party/" -e "gen-code/" -e "config/" -e "codegen/templates/" -e "codegen/template_bundle/"
ENV_CONFIG := $(shell find ./config -name 'production.json' -o -name 'production.yaml')
TEST_ENV_CONFIG := $(shell find ./config -name 'test.json' -o -name 'test.yaml')

# list all executables
PROGS = benchmarks/benchserver/benchserver \
	benchmarks/runner/runner

EXAMPLE_BASE_DIR = examples/example-gateway/
EXAMPLE_SERVICES_DIR = $(EXAMPLE_BASE_DIR)build/services/
EXAMPLE_SERVICES = $(sort $(dir $(wildcard $(EXAMPLE_SERVICES_DIR)*/)))
GOIMPORTS = "$(PWD)/vendor/golang.org/x/tools/cmd/goimports"
GOBINDATA = "$(PWD)/vendor/github.com/jteeuwen/go-bindata/go-bindata"
GOMOCK = "$(PWD)/vendor/github.com/golang/mock/mockgen"
GOGOSLICK = "$(PWD)/vendor/github.com/gogo/protobuf/protoc-gen-gogoslick"
YARPCGO = "$(PWD)/vendor/go.uber.org/yarpc/encoding/protobuf/protoc-gen-yarpc-go"

.PHONY: install
install:
	@echo "Mounting git pre-push hook"
	cp .git-pre-push-hook .git/hooks/pre-push
	@echo "Installing Glide and locked dependencies..."
	pip install --user yq
	glide --version || go get -u -f github.com/Masterminds/glide
	glide install
	go build -o $(GOIMPORTS)/goimports ./vendor/golang.org/x/tools/cmd/goimports/
	go build -o $(GOBINDATA)/go-bindata ./vendor/github.com/jteeuwen/go-bindata/go-bindata/
	go build -o $(GOMOCK)/mockgen ./vendor/github.com/golang/mock/mockgen/
	go build -o $(GOGOSLICK)/protoc-gen-gogoslick ./vendor/github.com/gogo/protobuf/protoc-gen-gogoslick/
	go build -o $(YARPCGO)/protoc-gen-yarpc-go ./vendor/go.uber.org/yarpc/encoding/protobuf/protoc-gen-yarpc-go/

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
	@./node_modules/.bin/eclint check "./{codegen,examples}/**/*.{json,tmpl}"

.PHONY: eclint-fix
eclint-fix:
	@ls ./node_modules/.bin/eclint >/dev/null 2>&1 || npm i eclint@v1.1.5
	./node_modules/.bin/eclint fix "./{codegen,examples}/**/*.{json,tmpl}"

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
	@$(GOIMPORTS)/goimports -d $(PKG_FILES) 2>&1 | $(FILTER_LINT) | tee -a lint.log
	@echo "Installing test dependencies for vet..."
	@go test -i $(PKGS)
	@echo "Checking printf statements..."
	@git grep -E 'Fprintf\(os.Std(err|out)' | $(FILTER_LINT) | tee -a lint.log
	@echo "Checking vet..."
	@$(foreach dir,$(PKG_FILES),go vet $(VET_RULES) ./$(dir)/... 2>&1 | $(FILTER_LINT) | tee -a lint.log;)
	@echo "Checking lint..."
	@go get golang.org/x/lint/golint
	@$(foreach dir,$(PKGS),golint $(dir) 2>&1 | $(FILTER_LINT) | tee -a lint.log;)
	@echo "Checking errcheck..."
	@go run vendor/github.com/kisielk/errcheck/main.go $(PKGS) 2>&1 | $(FILTER_LINT) | tee -a lint.log
	@echo "Checking staticcheck..."
	@go build -o vendor/honnef.co/go/tools/cmd/staticcheck/staticcheck vendor/honnef.co/go/tools/cmd/staticcheck/staticcheck.go
	@./vendor/honnef.co/go/tools/cmd/staticcheck/staticcheck $(PKGS) 2>&1 | $(FILTER_LINT) | tee -a lint.log
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
	@ls ./node_modules/.bin/uber-licence >/dev/null 2>&1 || npm i uber-licence
	@chmod 644 ./codegen/templates/*.tmpl
	@chmod 644 $(ENV_CONFIG)
	@$(GOBINDATA)/go-bindata -pkg config -nocompress -modtime 1 -prefix config -o config/production.gen.go $(ENV_CONFIG)
	@./node_modules/.bin/uber-licence --file "production.gen.go" --dir "config" > /dev/null
	@$(GOBINDATA)/go-bindata -pkg templates -nocompress -modtime 1 -prefix codegen/templates -o codegen/template_bundle/template_files.go codegen/templates/...
	@$(GOIMPORTS)/goimports -w -e "codegen/template_bundle/template_files.go"
	@PATH=$(GOGOSLICK):$(YARPCGO):$(GOIMPORTS):$(GOMOCK):$(PATH) bash ./scripts/generate.sh

.PHONY: check-generate
check-generate:
	@rm -f git-status.log
	rm -rf ./examples/example-gateway/build
	rm -rf ./examples/selective-gateway/build
	make generate
	# TODO: @rpatali enable after migrating to Github Actions
	#       coden-gen currently generates slightly different on Github vs. Mac
	#       mock-client imports have `zanzibar` alias on local where as `runtime` alias on GH
	#       so the git-status is not clean.
	#git status --porcelain > git-status.log
	#@[ ! -s git-status.log ] || ( cat git-status.log ; git --no-pager diff ; [ ! -s git-status.log ] );

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

.PHONY: test-only
test-only:
	@rm -f ./test/.cached_binary_test_info.json
	@echo "Running all tests..."
	@ZANZIBAR_CACHE=1 go test ./test/health_test.go # preload the binary cache
	@PATH=$(PATH):$(GOIMPORTS) ZANZIBAR_CACHE=1 go test -race ./codegen/... ./runtime/... | grep -v '\[no test files\]'
	@PATH=$(PATH):$(GOIMPORTS) ZANZIBAR_CACHE=1 go test -race $$(go list ./examples/example-gateway/... | grep -v build) | \
	 grep -v '\[no test files\]'
	@PATH=$(PATH):$(GOIMPORTS) ZANZIBAR_CACHE=1 go test -race $$(go list ./examples/selective-gateway/... | grep -v build) | \
	 grep -v '\[no test files\]'
	@PATH=$(PATH):$(GOIMPORTS) ZANZIBAR_CACHE=1 go test ./test/... ./examples/example-gateway/build/... ./examples/selective-gateway/build/... | \
	 grep -v '\[no test files\]'
	@rm -f ./test/.cached_binary_test_info.json
	@echo "<coverage />" > ./coverage/cobertura-coverage.xml

.PHONY: fast-bench
fast-bench:
	@rm -f bench.log bench-fail.log
	time -p sh -c "go test -run _NONE_ -bench . -benchmem -benchtime 1s -cpu 1 ./test/... | grep -v '^ok ' | grep -v '\[no test files\]' | grep -v '^PASS'  | tee -a bench.log"
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
		./bin/$* --config=$(TEST_ENV_CONFIG)

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
	@PATH=$(PATH):$(GOIMPORTS) bash ./scripts/cover.sh

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
	@rm -rf ./vendor/
	make install

.PHONY: jenkins-test
jenkins-test:
	make check-generate
	make lint
	make test-only

.PHONY: jenkins
jenkins:
	$(MAKE) jenkins-install
	$(MAKE) jenkins-test
