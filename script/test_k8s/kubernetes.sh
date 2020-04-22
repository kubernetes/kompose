#!/bin/bash

RED='\033[0;31m'
NOCOLOR='\033[0m'

start_k8s() {
  # Note: takes some time for the http server to pop up :)
  # MINIMUM 15 seconds
  echo "
  ##########
  STARTING KUBERNETES
  ##########
  "
  if [ ! -f /usr/bin/kubectl ] && [ ! -f /usr/local/bin/kubectl ]; then
    echo "No kubectl bin exists? Please install."
    return 1
  fi

  if [ ! -f /usr/bin/kind ] && [ ! -f /usr/local/bin/kind ]; then
    echo "No kind bin exists? Please install."
    return 1
  fi

  kind create cluster --name kompose-test

  kubectl cluster-info --context kind-kompose-test

  # Set the appropriate .kube/config configuration
  kubectl config set-cluster kind-kompose-test
  kubectl config use-context kind-kompose-test

  kubectl proxy --port=6443 &

  # Debug info:
  # cat ~/.kube/config

  # Delay due to CI being a bit too slow when first starting k8s
  sleep 5
}


stop_k8s() {
  echo "
  ##########
  STOPPING KUBERNETES
  ##########
  "

  kind delete cluster --name kompose-test
}

wait_k8s() {
  echo "Waiting for k8s po/svc/rc to finish terminating..."
  kubectl get po,svc,rc
  sleep 3 # give kubectl chance to catch up to api call
  while [ 1 ]
  do
    k8s=`kubectl get po,svc,rc | grep Terminating`
    if [[ $k8s == "" ]]
    then
      echo "k8s po/svc/rc terminated!"
      break
    else
      echo "..."
    fi
    sleep 1
  done
}

test_k8s() {
  for f in examples/*.yaml
  do
    echo -e "\n${RED}kompose up --server http://127.0.0.1:6443 -f $f ${NC}\n"
    ./kompose up --server http://127.0.0.1:6443 -f $f
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down --server http://127.0.0.1:6443 -f $f ${NC}\n"
    ./kompose down --server http://127.0.0.1:6443 -f $f
    echo -e "\nTesting controller=daemonset key\n"
    echo -e "\n${RED}kompose up --server http://127.0.0.1:6443 -f $f --controller=daemonset ${NC}\n"
    ./kompose up --server http://127.0.0.1:6443 -f $f --controller=daemonset
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down --server http://127.0.0.1:6443 -f $f --controller=daemonset ${NC}\n"
    ./kompose down --server http://127.0.0.1:6443 -f $f --controller=daemonset
    echo -e "\nTesting controller=replicationcontroller key\n"
    echo -e "\n${RED}kompose up --server http://127.0.0.1:6443 -f $f --controller=replicationcontroller ${NC}\n"
    ./kompose up --server http://127.0.0.1:6443 -f $f --controller=replicationcontroller
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down --server http://127.0.0.1:6443 -f $f --controller=replicationcontroller ${NC}\n"
    ./kompose down --server http://127.0.0.1:6443 -f $f --controller=replicationcontroller

  done

  echo -e "\nTesting stdin to kompose\n"
  echo -e "\n${RED}cat examples/docker-compose.yaml | ./kompose up --server http://127.0.0.1:6443 -f -${NC}\n"
  cat examples/docker-compose.yaml | ./kompose up --server http://127.0.0.1:6443 -f -
  sleep 2 # Sleep for k8s to catch up to deployment
  echo -e "\n${RED}cat examples/docker-compose.yaml | ./kompose down --server http://127.0.0.1:6443 -f - ${NC}\n"
  cat examples/docker-compose.yaml | ./kompose down --server http://127.0.0.1:6443 -f -
}

if [[ $1 == "start" ]]; then
  start_k8s
elif [[ $1 == "stop" ]]; then
  stop_k8s
elif [[ $1 == "wait" ]]; then
  wait_k8s
elif [[ $1 == "test" ]]; then
  test_k8s
else
  echo $"Usage: kubernetes.sh {answers|start|stop|wait}"
fi
