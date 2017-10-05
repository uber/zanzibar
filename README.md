## zanzibar [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][go-report-img]][go-report]

A build system & runtime component to generate configuration driven gateways.

## Installation

```
mkdir -p $GOPATH/src/github.com/uber
git clone git@github.com:uber/zanzibar $GOPATH/src/github.com/uber/zanzibar
cd $GOPATH/src/github.com/uber/zanzibar
make install
```

## Running the tests

```
make test
```

## Running the benchmarks

```
for i in `seq 5`; do make bench; done
```

## Running the end-to-end benchmarks

First fetch `wrk`

```
git clone https://github.com/wg/wrk ~/wrk
cd ~/wrk
make
sudo ln -s $HOME/wrk/wrk /usr/local/bin/wrk
```

Then you can run the benchmark comparison script

```
# Assume you are on feature branch ABC
./benchmarks/compare_to.sh master
```

## Running the server

First create log dir...

```
sudo mkdir -p /var/log/my-gateway
sudo chown $USER /var/log/my-gateway
chmod 755 /var/log/my-gateway

sudo mkdir -p /var/log/example-gateway
sudo chown $USER /var/log/example-gateway
chmod 755 /var/log/example-gateway
```

```
make run
# Logs are in /var/log/example-gateway/example-gateway.log
```

## Adding new dependencies

We use glide @ 0.12.3 to add dependencies.

Download [glide @ 0.12.3](https://github.com/Masterminds/glide/releases)
and make sure it's available in your path

If we want to add a dependency:

 - Add a new section to the glide.yaml with your package and version
 - run `glide up --quick`
 - check in the `glide.yaml` and `glide.lock`

If you want to update a dependency:

 - Change the `version` field in the `glide.yaml`
 - run `glide up --quick`
 - check in the `glide.yaml` and `glide.lock`

[doc-img]: https://godoc.org/github.com/uber/zanzibar?status.svg
[doc]: https://godoc.org/github.com/uber/zanzibar
[ci-img]: https://travis-ci.org/uber/zanzibar.svg?branch=master
[ci]: https://travis-ci.org/uber/zanzibar
[cov-img]: https://coveralls.io/repos/github/uber/zanzibar/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber/zanzibar?branch=master
[go-report-img]: https://goreportcard.com/badge/github.com/uber/zanzibar
[go-report]: https://goreportcard.com/report/github.com/uber/zanzibar

## Update golden files

Run the test that compares golden files with `-update` flag, e.g.,
```
go test ./backend/repository -update
```
