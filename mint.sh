#!/bin/bash
#
#  Mint (C) 2017 Minio, Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

CONTAINER_ID=$(awk -F / '{ print substr($NF, 1, 12) }' /proc/1/cpuset)
MINT_DATA_DIR=${MINT_DATA_DIR:-/mint/data}
MINT_MODE=${MINT_MODE:-core}
SERVER_REGION=${SERVER_REGION:-us-east-1}
ENABLE_HTTPS=${ENABLE_HTTPS:-0}
RUN_LIST=( "$@" )

if [ -z "$SERVER_ENDPOINT" ]; then
    SERVER_ENDPOINT="play.minio.io:9000"
    ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
    SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
    ENABLE_HTTPS=1
fi

ROOT_DIR="$PWD"
TESTS_DIR="$ROOT_DIR/run/core"

BASE_LOG_DIR="$ROOT_DIR/log"
LOG_FILE="log.json"
ERROR_FILE="error.log"
mkdir -p "$BASE_LOG_DIR"

function humanize_time()
{
    time="$1"
    days=$(( time / 60 / 60 / 24 ))
    hours=$(( time / 60 / 60 % 24 ))
    minutes=$(( time / 60 % 60 ))
    seconds=$(( time % 60 ))

    (( days > 0 )) && echo -n "$days days "
    (( hours > 0 )) && echo -n "$hours hours "
    (( minutes > 0 )) && echo -n "$minutes minutes "
    (( days > 0 || hours > 0 || minutes > 0 )) && echo -n "and "
    echo "$seconds seconds"
}

function run_test()
{
    if [ ! -d "$1" ]; then
        return 1
    fi

    sdk_name="$(basename "$1")"

    echo -n "Running $sdk_name tests ... "
    start=$(date +%s)

    mkdir -p "$BASE_LOG_DIR/$sdk_name"

    (cd "$sdk_dir" && ./run.sh "$BASE_LOG_DIR/$LOG_FILE" "$BASE_LOG_DIR/$sdk_name/$ERROR_FILE")
    rv=$?
    end=$(date +%s)
    duration=$(humanize_time $(( end - start )))

    if [ "$rv" -eq 0 ]; then
        echo "done in $duration"
    else
        echo "FAILED in $duration"
    fi

    return $rv
}

function main()
{
    export MINT_DATA_DIR
    export MINT_MODE
    export SERVER_ENDPOINT
    export ACCESS_KEY
    export SECRET_KEY
    export ENABLE_HTTPS
    export SERVER_REGION

    echo "Running with"
    echo "SERVER_ENDPOINT: $SERVER_ENDPOINT"
    echo "ACCESS_KEY:      $ACCESS_KEY"
    echo "SECRET_KEY:      ***REDACTED***"
    echo "ENABLE_HTTPS:    $ENABLE_HTTPS"
    echo "SERVER_REGION:   $SERVER_REGION"
    echo "MINT_DATA_DIR:   $MINT_DATA_DIR"
    echo "MINT_MODE:       $MINT_MODE"

    ## $MINT_MODE is used inside every sdks.
    echo "To get intermittent logs, 'sudo docker cp ${CONTAINER_ID}:/mint/log /tmp/mint-logs'"

    if [ "${#RUN_LIST[@]}" -ne 0 ]; then 
        for sdk in "${RUN_LIST[@]}"; do
            sdk_dir="$TESTS_DIR"/$sdk
            run_test "$sdk_dir"
        done
    else 
        for sdk_dir in "$TESTS_DIR"/*; do
            run_test "$sdk_dir"
        done
    fi

    echo "Finished running all tests."
    echo "To get logs, run 'sudo docker cp ${CONTAINER_ID}:/mint/log /tmp/mint-logs'"
}

main "$@"
