# How to run tests

* Copy `allinkl-credentials.sample.yaml` to `allinkl-credentials.yaml` and set your KAS credentials (the file is gitignored)
* Setup test-environment
* Execute Tests

## Setup

Install `setup-envtest` once (it lands in `$(go env GOPATH)/bin`, usually `~/go/bin`):

```bash
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
```

Then download the test binaries and point the suite at them:

```bash
ASSETS="$(setup-envtest use 1.36.x --print path)"
export TEST_ASSET_KUBE_APISERVER="$ASSETS/kube-apiserver"
export TEST_ASSET_ETCD="$ASSETS/etcd"
export TEST_ASSET_KUBECTL="$ASSETS/kubectl"
unset USE_EXISTING_CLUSTER
test -x "$TEST_ASSET_KUBE_APISERVER" && echo ok-apiserver
test -x "$TEST_ASSET_ETCD" && echo ok-etcd
"$TEST_ASSET_KUBECTL" version --client
```

## Execute

The suite creates and deletes real TXT records in the zone, so use a zone the
KAS account controls:

```bash
TEST_ZONE_NAME=johanneskueber.com. go test -v -run TestRunsSuite -timeout 25m ./main_test.go
```
