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

if [ -z "$SERVER_ENDPOINT" ]; then
    SERVER_ENDPOINT="play.minio.io:9000"
    ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
    SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
    ENABLE_HTTPS=1
fi

ROOT_DIR="$PWD"
ERROR_LOG_FILE="error.log"
OUTPUT_LOG_FILE="output.log"
TESTS_DIR="$ROOT_DIR/run/core"

BASE_LOG_DIR="$ROOT_DIR/log"
mkdir -p "$BASE_LOG_DIR"

function ignore_sdk()
{
    IFS=',' read -ra ignore_list <<<"${SKIP_TESTS}"
    for sdk in "${ignore_list[@]}"; do
        if [ "$1" == "$sdk" ]; then
            return 0
        fi
    done

    return 1
}

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

function main()
{
    export MINT_DATA_DIR
    export MINT_MODE
    export SERVER_ENDPOINT
    export ACCESS_KEY
    export SECRET_KEY
    export ENABLE_HTTPS
    export SERVER_REGION

    ## $MINT_MODE is used inside every sdks.
    echo "To get intermittent logs, 'sudo docker cp ${CONTAINER_ID}:/mint/log /tmp/mint-logs'"

    for sdk_dir in "$TESTS_DIR"/*; do
        if [ ! -d "$sdk_dir" ]; then
            continue
        fi

        sdk_name="$(basename "$sdk_dir")"
        if ignore_sdk "$sdk_name"; then
            echo "Ignoring $sdk_name tests"
            continue
        fi

        echo -n "Running $sdk_name tests ... "
        start=$(date +%s)

        mkdir -p "$BASE_LOG_DIR/$sdk_name"

        (cd "$sdk_dir" && ./run.sh "$BASE_LOG_DIR/$sdk_name/$OUTPUT_LOG_FILE" "$BASE_LOG_DIR/$sdk_name/$ERROR_LOG_FILE")
        rv=$?
        end=$(date +%s)
        duration=$(humanize_time $(( end - start )))

        if [ "$rv" -eq 0 ]; then
            echo "done in $duration"
        else
            echo "FAILED in $duration"
        fi
    done

    echo "Finished running all tests."
    echo "To get logs, run 'sudo docker cp ${CONTAINER_ID}:/mint/log /tmp/mint-logs'"
}

main "$@" &
main_pid=$!
trap 'echo -e "\nAborting Mint..."; kill $main_pid' SIGINT SIGTERM

# wait for main to complete
wait
