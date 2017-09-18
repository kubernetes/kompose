# Functional tests for Kompose on OpenShift

## Introduction

The functional tests for Kompose on OpenShift cluster leverages  `oc cluster up` to bring a single-cluster OpenShift instance. The test scripts
are hosted under script/test_in_openshift.

The directory structure is as below:

```
        script/test_in_openshift/
        ├── compose-files
        │   └── docker-compose-command.yml
        ├── lib.sh
        ├── run.sh
        └── tests
                ├── buildconfig.sh
                ├── entrypoint-command.sh
                ├── etherpad.sh
                └── redis-replica-2.sh
                └── ..
```

- [run.sh](/script/test_in_openshift/run.sh) is the master script which executes all the tests.
  
- [lib.sh](/script/test_in_openshift/lib.sh) consists of helper functions for `kompose up` and `kompose down` checks.
  
- [tests/](/script/test_in_openshift/tests) directory contains the test scripts.

- [compose-files/](/script/test_in_openshift/compose-files/) directory contains the docker compose file used by the test scripts.

- The scripts use [`oc cluster up`](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) for setting up a single-machine OpenShift cluster. It exits if oc binaries are not installed.

- Docker Compose examples are available under [examples/](/examples] or [/script/test/fixtures](/script/test/fixtures).

## Running OpenShift tests

### Deploy `oc cluster up`

The scripts use [`oc cluster up`](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) for setting up a single-machine OpenShift cluster. Installing oc binary is a pre-requisite for running these tests.

Use `make test-openshift` to run the OpenShift tests.

## Adding OpenShift tests

* You can add the OpenShift tests by adding your script under [script/test_in_openshift/tests](/script/test_in_openshift/tests). 
