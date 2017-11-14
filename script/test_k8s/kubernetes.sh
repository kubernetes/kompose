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

  # Uses https://github.com/kubernetes/minikube/tree/master/deploy/docker
  # In which we have to git clone and create the image..
  # https://github.com/kubernetes/minikube
  # Thus we are using a public docker image

  IMAGE=calpicow/localkube-image:v1.5.3
  docker run -d \
      --volume=/:/rootfs:ro \
      --volume=/sys:/sys:rw \
      --volume=/var/lib/docker:/var/lib/docker:rw \
      --volume=/var/lib/kubelet:/var/lib/kubelet:rw \
      --volume=/var/run:/var/run:rw \
      --net=host \
      --pid=host \
      --privileged \
      --name=minikube \
      $IMAGE \
      /localkube start \
      --apiserver-insecure-address=0.0.0.0 \
      --apiserver-insecure-port=8080 \
      --logtostderr=true \
      --containerized

  until curl 127.0.0.1:8080 &>/dev/null;
  do
      echo ...
      sleep 1
  done

  # Set the appropriate .kube/config configuration
  kubectl config set-cluster dev --server=http://localhost:8080
  kubectl config set-context dev --cluster=dev --user=default
  kubectl config use-context dev
  kubectl config set-credentials default --token=foobar

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
  docker rm -f minikube

  # Delete via image name k8s.gcr.io
  # Delete all containers started (names start with k8s_)
  # Run twice in-case a container is replicated during that time
  for run in {0..2}
  do
    docker ps -a | grep 'k8s_' | awk '{print $1}' | xargs --no-run-if-empty docker rm -f
    docker ps -a | grep 'k8s.gcr.io/hyperkube-amd64' | awk '{print $1}' | xargs --no-run-if-empty docker rm -f
  done
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
    echo -e "\n${RED}kompose up -f $f ${NC}\n"
    ./kompose up -f $f
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down -f $f ${NC}\n"
    ./kompose down -f $f
    echo -e "\nTesting controller=daemonset key\n"
    echo -e "\n${RED}kompose up -f $f --controller=daemonset ${NC}\n"
    ./kompose up -f $f --controller=daemonset
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down -f $f --controller=daemonset ${NC}\n"
    ./kompose down -f $f --controller=daemonset
    echo -e "\nTesting controller=replicationcontroller key\n"
    echo -e "\n${RED}kompose up -f $f --controller=replicationcontroller ${NC}\n"
    ./kompose up -f $f --controller=replicationcontroller
    sleep 2 # Sleep for k8s to catch up to deployment
    echo -e "\n${RED}kompose down -f $f --controller=replicationcontroller ${NC}\n"
    ./kompose down -f $f --controller=replicationcontroller

  done

  echo -e "\nTesting stdin to kompose\n"
  echo -e "\n${RED}cat examples/docker-compose.yaml | ./kompose up -f -${NC}\n"
  cat examples/docker-compose.yaml | ./kompose up -f -
  sleep 2 # Sleep for k8s to catch up to deployment
  echo -e "\n${RED}cat examples/docker-compose.yaml | ./kompose down -f - ${NC}\n"
  cat examples/docker-compose.yaml | ./kompose down -f -
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
