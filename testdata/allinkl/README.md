# How to run tests

* Set your credentials in the `allinl-credentials.yaml`
* Setup test-environment
* Execute Tests

## Setup

'''bash
ASSETS="$(setup-envtest use 1.34.x --print path)"                                
export TEST_ASSET_KUBE_APISERVER="$ASSETS/kube-apiserver"
export TEST_ASSET_ETCD="$ASSETS/etcd"
export TEST_ASSET_KUBECTL="$ASSETS/kubectl"
unset USE_EXISTING_CLUSTER
test -x "$TEST_ASSET_KUBE_APISERVER" && echo ok-apiserver
test -x "$TEST_ASSET_ETCD" && echo ok-etcd
"$TEST_ASSET_KUBECTL" version --client
'''

## Execute
'''bash
TEST_ZONE_NAME=johanneskueber.com. go test -v -run TestRunsSuite ./main_test.go   
'''