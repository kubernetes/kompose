#!/bin/bash

# These tests will bring up a single-node Kubernetes cluster and test against examples
# WARNING: This will actively create Docker containers on your machine as well as remove them
# do not run this on a production cluster / machine.


# Check requirements!
if ! hash go 2>/dev/null; then
  echo "ERROR: go required"
  exit 1
fi

if ! hash docker 2>/dev/null; then
  echo "ERROR: docker required"
  exit 1
fi

if ! hash kubectl 2>/dev/null; then
  echo "ERROR: kubectl required"
  exit 1
fi

# First off, we have to compile the latest binary
# We *assume* that the binary has already been built
# make bin

#####################
# KUBERNETES TESTS ##
#####################

# Now we can start our Kubernetes cluster!
./script/test_k8s/kubernetes.sh start

# And we're off! Let's test those example files
./script/test_k8s/kubernetes.sh test

# Stop our Kubernetes cluster
./script/test_k8s/kubernetes.sh stop
