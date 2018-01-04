#!/bin/bash

# Copyright 2017 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


function convert::print_msg () {
    echo ""
    tput setaf 4
    tput bold
    echo -e "$@"
    tput sgr0
    echo ""
}

function install_oc_client () {
    # Valid only for Travis
    convert::print_msg "Installing oc client binary ..."
    sudo sed -i 's:DOCKER_OPTS=":DOCKER_OPTS="--insecure-registry 172.30.0.0/16 :g' /etc/default/docker
    sudo mv /bin/findmnt /bin/findmnt.backup
    sudo /etc/init.d/docker restart
    # FIXME
    wget https://github.com/openshift/origin/releases/download/v1.5.0/openshift-origin-client-tools-v1.5.0-031cbe4-linux-64bit.tar.gz -O /tmp/oc.tar.gz 2> /dev/null > /dev/null
    mkdir /tmp/ocdir && cd /tmp/ocdir && tar -xvvf /tmp/oc.tar.gz > /dev/null
    sudo mv /tmp/ocdir/*/oc /usr/bin/
}


function convert::oc_cluster_up () {

    oc cluster up; exit_status=$?

    if [ $exit_status -ne 0 ]; then
	FAIL_MSGS=$FAIL_MSGS"exit status: $exit_status\n";
	convert::print_fail "oc cluster up failed.\n"
	exit $exit_status
    fi

    convert::run_cmd "oc login -u system:admin"
}

function convert::oc_registry_login () {
    # wait for the registry to become available
    local counter=0
    while ! curl --fail --silent http://172.30.1.1:5000/healthz; do
        counter=$(($counter + 1))
        if [ $counter = 48 ]; then
            echo "Registry did not become available in time"
            break
        fi
        sleep 5
    done
    oc serviceaccounts get-token builder \
        | docker login --password-stdin -u builder 172.30.1.1:5000
}

function convert::oc_cluster_down () {

    convert::run_cmd "oc cluster down"
    exit_status=$?

    if [ $exit_status -ne 0 ]; then
	FAIL_MSGS=$FAIL_MSGS"exit status: $exit_status\n"
	exit $exit_status
    fi

}

function convert::kompose_up_check () {
    # A function to check if the pods are up in 'Running' state
    # WARNING: This function has a limitation that it works only for 2 pods.
    # TODO: Make this function work for any no. of pods
    # Usage: -p for pod name, -r replica count
    local retry_up=0

    while getopts ":p:r:" opt; do
	case $opt in
	    p ) pod=$OPTARG;;
	    r ) replica=$OPTARG;;
	esac
    done

    if [ -z $replica ]; then
       replica_1=1
       replica_2=1
    else
       replica_1=$replica
       replica_2=$replica
    fi

    pod_1=$( echo $pod | awk '{ print $1 }')
    pod_2=$( echo $pod | awk '{ print $2 }')

    query_1='grep ${pod_1} | grep -v deploy'
    query_2='grep ${pod_2} | grep -v deploy'

    query_1_status='Running'
    query_2_status='Running'

    is_buildconfig=$(oc get builds --no-headers | wc -l)

    if [ $is_buildconfig -gt 0 ]; then
	query_1='grep ${pod_1} | grep -v deploy | grep -v build'
	query_2='grep build | grep -v deploy'
	query_2_status='Completed'
	replica_2=1
    fi

    convert::print_msg "Waiting for the pods to come up ..."

    # FIXME: Make this generic to cover all cases
    while [ $(oc get pods | eval ${query_1} | awk '{ print $3 }' | \
		     grep ${query_1_status} | wc -l) -ne $replica_1 ] ||
	      [ $(oc get pods | eval ${query_2} | awk '{ print $3 }' | \
			 grep ${query_2_status} | wc -l) -ne $replica_2 ]; do

	if [ $retry_up -lt 240 ]; then
	    retry_up=$(($retry_up + 1))
	    sleep 1
	else
	    convert::print_fail "kompose up has failed to bring the pods up\n"
	    oc get pods
	    exit 1
	fi

    done

    # Wait
    sleep 2

    # If pods are up, print a success message
    if [ $(oc get pods | eval ${query_1} | awk '{ print $3 }' | \
		  grep ${query_1_status} | wc -l) -eq $replica_1 ] &&
	   [ $(oc get pods | eval ${query_2} | awk '{ print $3 }' | \
		      grep ${query_2_status} | wc -l) -eq $replica_2 ]; then
	oc get pods
	convert::print_pass "All pods are Running now. kompose up is successful.\n"
    fi
}

function convert::kompose_down_check () {
    # Checks if the pods have been successfully deleted
    # with 'kompose down'.
    # Usage: convert::kompose_down_check <num_of_pods>

    local retry_down=0
    local pod_count=$1

    convert::print_msg "Waiting for the pods to go down ..."

    while [ $(oc get pods | wc -l ) != 0 ] &&
	  [ $(oc get pods | grep -v deploy | grep 'Terminating' | wc -l ) != $pod_count ]; do
	if [ $retry_down -lt 120 ]; then
	    retry_down=$(($retry_down + 1))
	    sleep 1
	else
	    convert::print_fail "kompose down has failed\n"
	    oc get pods
	    exit 1
	fi
    done

    # Wait
    sleep 2

    # Print a message if all the pods are down
    if [ $(oc get pods | wc -l ) == 0 ] ||
       [ $(oc get pods | grep -v deploy | grep 'Terminating' | wc -l ) == $pod_count ]; then
	convert::print_pass "All pods are down now. kompose down successful.\n"
	oc get pods
    fi
}

function convert::oc_cleanup () {
    oc delete bc,rc,rs,svc,is,dc,deploy,ds,builds,route --all > /dev/null
}

function convert::oc_check_route () {
    local route_key=$1
    if [ $route_key == 'true' ]; then
	route_key='nip.io'
    fi

    if [ $(oc get route | grep ${route_key} | wc -l ) -gt 0 ]; then
	convert::print_pass "Route *.${route_key} has been exposed\n"
    else
	convert::print_fail "Route *.${route_key} has not been exposed\n"
    fi

    echo ""
    oc get route
}

function convert::kompose_up () {
    # Function for running 'kompose up'
    # Usage: convert::kompose_up <docker_compose_file>
    local compose_file=$1
    convert::print_msg "Running kompose up ..."
    ./kompose up --provider=openshift --volumes emptyDir -f $compose_file
    exit_status=$?

    if [ $exit_status -ne 0 ]; then
	convert::print_fail "kompose up has failed\n"
	exit 1
    fi
}

function convert::kompose_down () {
    # Function for running 'kompose down'
    # Usage: convert::kompose_down <docker_compose_file>
    local compose_file=$1
    convert::print_msg "Running kompose down ..."
    ./kompose --provider=openshift -f $compose_file down
    exit_status=$?

    if [ $exit_status -ne 0 ]; then
	convert::print_fail "kompose down has failed\n"
	exit 1
    fi
}
