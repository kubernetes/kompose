#!/bin/bash

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/script/test/cmd/lib.sh


#########################################
# normal tests without preference file

# kompose convert
# should generate deployments and services
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deployment.json" "mariadb-deployment.json"

# kompose convert --ds --d
# should generate daemonsets and deployments
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml convert --ds --d" "etherpad-service.json" "mariadb-service.json" "etherpad-deployment.json" "etherpad-daemonset.json" "mariadb-deployment.json" "mariadb-daemonset.json"

# kompose --provider=openshift convert
# should generate service, deploymentconfig, imagestreams
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deploymentconfig.json" "etherpad-imagestream.json" "mariadb-deploymentconfig.json" "mariadb-imagestream.json"


#########################################
# with preference file

# valid preference file1
# kompose convert; reads from pref-file
# should generate services, deployments, replicationcontrollers
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid1.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deployment.json" "etherpad-replicationcontroller.json" "mariadb-deployment.json" "mariadb-replicationcontroller.json"

# valid preference file1
# kompose --provider openshift convert; overrides pref-file and creates openshift artifacts
# should generate service, deploymentconfig, imagestreams
convert::check_artifacts_generated "kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid1.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deploymentconfig.json" "etherpad-imagestream.json" "mariadb-deploymentconfig.json" "mariadb-imagestream.json"

# valid preference file1
# kompose convert --ds --d; augments to what has been provided in pref-file from cmd line
# should generate services, deployments, replicationcontrollers, daemonsets
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid1.yml convert --ds --d" "etherpad-service.json" "mariadb-service.json" "etherpad-deployment.json" "etherpad-replicationcontroller.json" "mariadb-deployment.json" "mariadb-replicationcontroller.json" "etherpad-daemonset.json" "mariadb-daemonset.json"

# valid preference file2
# kompose convert; reads from pref-file
# should generate service, deploymentconfig, imagestreams
convert::check_artifacts_generated "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid2.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deploymentconfig.json" "etherpad-imagestream.json" "mariadb-deploymentconfig.json" "mariadb-imagestream.json"

# valid preference file2
# kompose --provider kubernetes convert; overrides pref-file and creates kubernetes artifacts
# should generate service, deployments
convert::check_artifacts_generated "kompose --provider=kubernetes -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid2.yml convert" "etherpad-service.json" "mariadb-service.json" "etherpad-deployment.json" "mariadb-deployment.json"

# valid preference file2
# kompose --provider kubernetes convert --rc; over-rides pref-file and create openshift artifacts
# should generate service and replicationcontrollers
convert::check_artifacts_generated "kompose --provider=kubernetes -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-valid2.yml convert --rc" "etherpad-service.json" "mariadb-service.json" "etherpad-replicationcontroller.json" "mariadb-replicationcontroller.json"

#########################################

# with invalid pref-file

# invalid profile in preference file
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-invalid-profile.yml convert"

# invalid provider in preference file, k8s instead of kubernetes
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-invalid-provider.yml convert"

# invalid object provided, rcs instead of replicationcontroller
convert::expect_failure "kompose -f $KOMPOSE_ROOT/script/test/fixtures/preference-file/docker-compose.yml --pref-file $KOMPOSE_ROOT/script/test/fixtures/preference-file/kompose-invalid-objects.yml convert"


exit $EXIT_STATUS
