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

KOMPOSE_ROOT=$(readlink -f $(dirname "${BASH_SOURCE}")/../../..)
source $KOMPOSE_ROOT/script/test/cmd/globals.sh

# setup all the things needed to run tests
function convert::init() {
    mkdir -p $TEMP_DIR
    SUCCESS_MSGS=""
    FAIL_MSGS=""
}
readonly -f convert::init

# remove all the temporary files created for test
function convert::teardown() {
    rm -rf $TEMP_DIR
    SUCCESS_MSGS=""
    FAIL_MSGS=""
}
readonly -f convert::teardown

# print about test start information
function convert::start_test() {
    convert::init
    echo -e "\n\n===> Starting test <==="
    echo $@
}
readonly -f convert::start_test

# print in green about the test being passed
function convert::print_pass() {
    tput setaf 2
    tput bold
    echo -en "PASS: $@"
    tput sgr0
}
readonly -f convert::print_pass

# print in red about the test failed
function convert::print_fail() {
    tput setaf 1
    tput bold
    echo -en "FAIL: $@"
    tput sgr0

}
readonly -f convert::print_fail

# run a cmd, which saves stdout output to TEMP_STDOUT
# and saves errors to TEMP_STDERR files and returns exit status
function convert::run_cmd() {
    cmd=$@

    $cmd 2>$TEMP_STDERR >$TEMP_STDOUT
    return $?
}
readonly -f convert::run_cmd


function convert::expect_cmd_success() {
    local cmd=$1

    convert::start_test "convert::expect_cmd_success: Running: '${cmd}'"

    convert::run_cmd $cmd
    exit_status=$?
    if [ $exit_status -ne 0 ]; then FAIL_MSGS=$FAIL_MSGS"exit status: $exit_status\n"; return $exit_status; fi
}
# run the command and match the output to the existing file
# if error then save error string in FAIL_MSGS
# if success save pass string in SUCCESS_MSGS
function convert::match_output() {
    local cmd=$1
    local expected_output=$2

    convert::run_cmd $cmd
    exit_status=$?
    if [ $exit_status -ne 0 ]; then FAIL_MSGS=$FAIL_MSGS"exit status: $exit_status\n"; return $exit_status; fi

    match=$(jq --argfile a $TEMP_STDOUT --argfile b $expected_output -n 'def post_recurse(f): def r: (f | select(. != null) | r), .; r; def post_recurse: post_recurse(.[]?); ($a | (post_recurse | arrays) |= sort) as $a | ($b | (post_recurse | arrays) |= sort) as $b | $a == $b')
    $cmd > /tmp/test.json
    diff /tmp/test.json $expected_output > /tmp/diff
    rm /tmp/test.json
    if [ "$match" = true ]; then SUCCESS_MSGS=$SUCCESS_MSGS"converted output matches\n"; return 0;
    else FAIL_MSGS=$FAIL_MSGS"converted output does not match\n"; cat /tmp/diff; rm /tmp/diff; return 1; fi
}
readonly -f convert::match_output

# function called from outside which accepts cmd to run and
# file to compare output with
function convert::expect_success() {
    local cmd=$1
    local expected_output=$2

    convert::start_test "convert::expect_success: Running: '${cmd}' expected_output: '${expected_output}'"

    convert::match_output "$cmd" "$expected_output"
    if [ $? -ne 0 ]; then convert::print_fail $FAIL_MSGS; convert::teardown; EXIT_STATUS=1; return 1;
    else convert::print_pass $SUCCESS_MSGS; fi

    # check if no warnings are generated? If yes then fail
    warnings=$(stat -c%s $TEMP_STDERR)
    if [ $warnings -ne 0 ]; then convert::print_fail "warnings given: $(cat $TEMP_STDERR)"; EXIT_STATUS=1; fi

    convert::teardown
}
readonly -f convert::expect_success

# function called from outside, which accepts cmd to run,
# expected output file and warnings if any
function convert::expect_success_and_warning() {
    local cmd=$1
    local expected_output=$2
    local expected_warning=$3

    convert::start_test "convert::expect_success_and_warning: Running: '${cmd}' expected_output: '${expected_output}' expected_warning: '${expected_warning}'"

    convert::match_output "$cmd" "$expected_output"
    if [ $? -ne 0 ]; then convert::print_fail $FAIL_MSGS; convert::teardown; EXIT_STATUS=1; return 1;
    else convert::print_pass $SUCCESS_MSGS; fi

    grep -i "$expected_warning" $TEMP_STDERR > /dev/null
    local exit_status=$?
    if [ $exit_status -ne 0 ]; then convert::print_fail "no warning found: '$expected_warning'"; EXIT_STATUS=1;
    else convert::print_pass "warning found: '$expected_warning'"; fi

    convert::teardown
    return $exit_status
}
readonly -f convert::expect_success_and_warning

# function called from outside, which accepts cmd to run,
# expects warning, without caring if the cmd ran passed or failed
function convert::expect_warning() {
    local cmd=$1
    local expected_warning=$2

    convert::start_test "convert::expect_warning: Running: '${cmd}' expected_warning: '${expected_warning}'"

    $cmd 2>$TEMP_STDERR >$TEMP_STDOUT

    grep -i "$expected_warning" $TEMP_STDERR > /dev/null
    local exit_status=$?
    if [ $exit_status -ne 0 ]; then convert::print_fail "no warning found: '$expected_warning'"; EXIT_STATUS=1;
    else convert::print_pass "warning found: '$expected_warning'"; fi

    convert::teardown
    return $exit_status
}
readonly -f convert::expect_warning

# function called from outside, which accepts cmd to run,
# expects failure, if the command passes then errors out.
function convert::expect_failure() {
    local cmd=$1

    convert::start_test "convert::expect_failure: Running: '${cmd}'"
    convert::run_cmd $cmd
    exit_status=$?
    if [ $exit_status -eq 0 ]; then convert::print_fail "no error output, returned exit status 0"; EXIT_STATUS=1;
    else convert::print_pass "errored out with exit status: $exit_status"; fi

    convert::teardown
    return $exit_status
}
readonly -f convert::expect_failure

# see if the given files exists
function utils::file_exists() {
    for file in "$@"
    do
        exit_status=$([ -f $file ]; echo $?)
        if [ $exit_status -ne 0 ]; then convert::print_fail "$file does not exist\n"; EXIT_STATUS=1;
        else convert::print_pass "$file exists\n"; fi
    done
}
readonly -f utils::file_exists

# delete given files one by one
function utils::remove_files() {
    for file in "$@"
    do
        rm $file
    done
}
readonly -f utils::remove_files

function convert::check_artifacts_generated() {
    local cmd=$1

    convert::start_test "convert::check_artifacts_generated: Running: '${cmd}'"
    convert::run_cmd $cmd
    # passing all args except the first one
    utils::file_exists "${@:2}"
    utils::remove_files "${@:2}"

    convert::teardown
    return $exit_status
}
readonly -f convert::check_artifacts_generated