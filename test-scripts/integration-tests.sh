#!/usr/bin/env bash

SCRIPT_DIR=$(cd "$(dirname "$0")"; pwd)
TEST_OUTPUT_FILE=$SCRIPT_DIR/test_output.txt
PID=-1

function check_test_output() {
    # We are going to give up to 10 seconds for lostromos to populate and flush to the file buffer. After that we will
    # assume a failure occurred.
    COUNTER=10
    until [ $COUNTER == 0 ]; do
        output=`cat $TEST_OUTPUT_FILE | sed -n "$1p"`
        if [ "$2" == "$output" ]; then
            echo "Matched expected output $2."
            return
        else
            # If we didn't match the output, then first check if we got an empty string. If so, then we likely haven't
            # cleared the buffer yet and should loop. If not, then we just failed so end the loop.
            if [ "$output" == "" ]; then
                sleep 1
                let COUNTER=COUNTER-1
            else
                let COUNTER=0
            fi
        fi
    done

    kill $PID
    if [ "$output" == "" ]; then
        echo "Failed to get any output, expected '$2'."
    else
        echo "Failed to get correct output. Found '$output', expected '$2'."
    fi
    echo "Contents of the output file:"
    echo
    cat $TEST_OUTPUT_FILE
    exit 1
}

# Start out with a clean slate before running the tests.
function remove_files_and_configs() {
    rm -f $TEST_OUTPUT_FILE
    kubectl delete -f test-data/cr_things.yml || true
    kubectl delete -f test-data/cr_nemo.yml || true
}


# Create the initial custom resource definition as well as "thing1" and "thing2" to be picked up by the watcher.
function setup_initial_test_data() {
    kubectl apply -f test-data/crd.yml
    kubectl apply -f test-data/cr_things.yml
}

# Start the watcher and pipe output to test_output.txt to keep track of for testing. Checks to see that "thing1" and
# "thing2" exist.
function start_watcher() {
    go run main.go start --config test-data/config.yaml > $TEST_OUTPUT_FILE &
    PID=`echo $!`
    echo "Watcher running with PID: $PID ..."

    # Check every 5 seconds to see if the watcher has started outputting to a file. Waits for up to 60 seconds.
    COUNTER=12
    until [ $COUNTER == 0 ]; do
        if [ -s $TEST_OUTPUT_FILE ]; then
            echo "Watcher started..."
            let COUNTER=0
        else
            echo "Watcher hasn't started yet, waiting 5 seconds..."
            sleep 5
            let COUNTER=COUNTER-1
        fi
    done

    check_test_output 2 "INFO: resource added, cr: thing1"
    check_test_output 6 "INFO: resource added, cr: thing2"
}

# Add a new custom resource "nemo" and ensure the watcher picks up the creation.
function create_new_test_data() {
    kubectl apply -f test-data/cr_nemo.yml
    check_test_output 10 "INFO: resource added, cr: nemo"
}

# Modify the custom resource "nemo" and ensure the watcher picks up the update.
function edit_test_data() {
    cp test-data/cr_nemo.yml test-data/cr_nemo_update.yml
    echo "  When: 2000 Something" >> test-data/cr_nemo_update.yml
    kubectl replace -f test-data/cr_nemo_update.yml
    check_test_output 14 "INFO: resource updated, cr: nemo"
    rm -f test-data/cr_nemo_update.yml
}

# Remove each of the custom resources and the definition. Checks to make sure first that "nemo" was removed, then makes
# sure "thing1" and "thing2" are also removed.
function remove_test_data() {
    kubectl delete -f test-data/cr_nemo.yml
    check_test_output 18 "INFO: resource deleted, cr: nemo"
    kubectl delete -f test-data/cr_things.yml
    check_test_output 22 "INFO: resource deleted, cr: thing1"
    check_test_output 26 "INFO: resource deleted, cr: thing2"
}

# Run the integration test step by step
remove_files_and_configs
setup_initial_test_data
start_watcher
create_new_test_data
edit_test_data
remove_test_data
kill $PID
rm -f $TEST_OUTPUT_FILE