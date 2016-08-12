#!/bin/bash

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

    if [ "$match" = false ]; then FAIL_MSGS=$FAIL_MSGS"converted output does not match\n"; return 1;
    else SUCCESS_MSGS=$SUCCESS_MSGS"converted output matches\n"; return 0; fi
}
readonly -f convert::match_output

# function called from outside which accecpts cmd to run and
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
