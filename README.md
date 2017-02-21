## zanzibar

A build system & runtime component to generate configuration driven gateways. Edit

## Installation

```
mkdir -p $GOPATH/src/github.com/uber
git clone gitolite@code.uber.internal:uber/github/zanzibar $GOPATH/src/github.com/uber/zanzibar
cd $GOPATH/src/github.com/uber/zanzibar
make install
```

## Running the tests

```
make test
```

## Running the server

First create log dir...

```
sudo mkdir -p /var/log/my-gateway
sudo chown $USER /var/log/my-gateway
chmod 755 /var/log/my-gateway
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
