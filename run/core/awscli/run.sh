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

# Execute test.sh
run() {
    [ "$ENABLE_HTTPS" == "1" ] && scheme="https" || scheme="http"
    endpoint_url=$scheme://$SERVER_ENDPOINT

    echo "Starting aws cli tests on ${endpoint_url}"
    ./test.sh "$endpoint_url"
}

configure() {
    echo "Configure aws cli secrets."

    aws configure set aws_access_key_id "$ACCESS_KEY"
    aws configure set aws_secret_access_key "$SECRET_KEY"
    aws configure set default.region "$SERVER_REGION"
}

main() {

    logfile=$1
    errfile=$2

    # run the tests
    rc=0

    # configure aws cli
    configure >>"$logfile" 2>&1 || { echo 'aws cli setup failed'; exit 1; }

    run 2>>"$errfile" 1>>"$logfile" || { echo 'aws cli run failed.'; rc=1; }
    grep -e '<ERROR>' "$logfile" >> "$errfile"

    return $rc
}

# invoke the script
main "$@"
