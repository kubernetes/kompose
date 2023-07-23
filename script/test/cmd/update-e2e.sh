#!/bin/bash
# HOW TO USE: 
# In your kompose director execute: UPDATE_OS=true UPDATE_K8S=false ./update-e2e.sh


make bin
# Feel free to update this variable to match your current FS
export KOMPOSE_ROOT=$HOME/projects/kompose

if $UPDATE_K8S ; then
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/v2.0/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/v2.0/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/v3.0/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/v3.0/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/docker-compose-placement.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml convert --stdout --with-kompose-annotation=false --volumes=configMap > $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose-withlabel.yml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-k8s-withlabel.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout --volumes emptyDir > $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s-empty-vols-template.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout --emptyvols > $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/expose/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/service-group/compose.yaml convert --stdout --with-kompose-annotation=false --service-group-mode=volume >$KOMPOSE_ROOT/script/test/fixtures/service-group/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-healthcheck.yaml convert --stdout --service-group-mode=label --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/docker-compose.yaml convert --stdout --with-kompose-annotation=false --controller statefulset > $KOMPOSE_ROOT/script/test/fixtures/statefulset/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/single-file-output/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/single-file-output/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/docker-compose-v1.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-k8s-v1.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/docker-compose-v2.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-k8s-v2.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/compose-file-support/compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/compose-file-support/output-k8s.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/env/docker-compose.yml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/env/output-k8s.yaml
fi

if $UPDATE_OS ; then
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/v2.0/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/v2.0/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/volume-mounts/windows/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/v3.0/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/v3.0/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/docker-compose-placement.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/deploy/placement/output-placement-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose.yml convert --stdout --with-kompose-annotation=false --volumes=configMap > $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/docker-compose-withlabel.yml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/configmap-volume/output-os-withlabel.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout --volumes emptyDir > $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os-empty-vols-template.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/docker-compose.yml convert --with-kompose-annotation=false --stdout --emptyvols > $KOMPOSE_ROOT/script/test/fixtures/change-in-volume/output-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/expose/compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/expose/output-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/first.yaml -f $KOMPOSE_ROOT/script/test/fixtures/multiple-files/second.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/multiple-files/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/healthcheck/docker-compose-healthcheck.yaml convert --stdout --service-group-mode=label --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/healthcheck/output-healthcheck-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/statefulset/docker-compose.yaml convert --stdout --with-kompose-annotation=false --controller statefulset > $KOMPOSE_ROOT/script/test/fixtures/statefulset/output-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f  $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/docker-compose.yaml  convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/multiple-type-volumes/output-os.yaml
$KOMPOSE_ROOT/kompose  --provider=openshift -f  $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/docker-compose.yaml  convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/envvars-interpolation/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/docker-compose.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/host-port-protocol/output-os.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/docker-compose-v1.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-os-v1.yaml
$KOMPOSE_ROOT/kompose --provider=openshift -f $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/docker-compose-v2.yaml convert --stdout --with-kompose-annotation=false > $KOMPOSE_ROOT/script/test/fixtures/external-traffic-policy/output-os-v2.yaml
$KOMPOSE_ROOT/kompose -f $KOMPOSE_ROOT/script/test/fixtures/env/docker-compose.yml convert --stdout --with-kompose-annotation=false --provider openshift > $KOMPOSE_ROOT/script/test/fixtures/env/output-os.yaml
fi
