#!/bin/bash

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/script/test/cmd/lib.sh

#######
# Tests related to docker-compose file in /script/test/fixtures/etherpad
convert::expect_failure "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml"

# commenting this test case out until image handling is fixed
#convert::expect_failure "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose-no-image.yml"
convert::expect_warning "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose-no-ports.yml" "Service cannot be created because of missing port."
export $(cat $KOMPOSE_ROOT/script/test/fixtures/etherpad/envs)
# kubernetes test
convert::expect_success_and_warning "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/etherpad/output-k8s.json" "Unsupported key depends_on - ignoring"
# openshift test
convert::expect_success_and_warning "kompose convert --stdout --dc -f $KOMPOSE_ROOT/script/test/fixtures/etherpad/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/etherpad/output-os.json" "Unsupported key depends_on - ignoring"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/etherpad/envs | cut -d'=' -f1)

######
# Tests related to docker-compose file in /script/test/fixtures/gitlab
convert::expect_failure "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml"
export $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs)
# kubernetes test
convert::expect_success "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/gitlab/output-k8s.json"
# openshift test
convert::expect_success "kompose convert --stdout --dc -f $KOMPOSE_ROOT/script/test/fixtures/gitlab/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/gitlab/output-os.json"
unset $(cat $KOMPOSE_ROOT/script/test/fixtures/gitlab/envs | cut -d'=' -f1)

######
# Tests related to docker-compose file in /script/test/fixtures/ngnix-node-redis
# kubernetes test
convert::expect_success_and_warning "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/ngnix-node-redis/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/ngnix-node-redis/output-k8s.json" "Unsupported key build - ignoring"
# openshift test
convert::expect_success_and_warning "kompose convert --stdout --dc -f $KOMPOSE_ROOT/script/test/fixtures/ngnix-node-redis/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/ngnix-node-redis/output-os.json" "Unsupported key build - ignoring"


######
# Tests related to docker-compose file in /script/test/fixtures/entrypoint-command
# kubernetes test
convert::expect_success_and_warning "kompose convert --stdout -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-k8s.json" "Service cannot be created because of missing port."
# openshift test
convert::expect_success_and_warning "kompose convert --stdout --dc -f $KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/docker-compose.yml" "$KOMPOSE_ROOT/script/test/fixtures/entrypoint-command/output-os.json" "Service cannot be created because of missing port."


exit $EXIT_STATUS

