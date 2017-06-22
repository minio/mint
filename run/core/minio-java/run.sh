#!/usr/bin/env bash
#
#  Minio Cloud Storage, (C) 2017 Minio, Inc.
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

run() {
	[ "$ENABLE_HTTPS" == "1" ] && scheme="https" || scheme="http" 
    ENDPOINT_URL=$scheme://"$SERVER_ENDPOINT"

    java -cp /usr/local/minio.jar":." FunctionalTest "$ENDPOINT_URL" "$ACCESS_KEY" "$SECRET_KEY" "$S3_REGION"
}

main() {
    logfile=$1
    errfile=$2

    # run the tests
    rc=0
    run 2>>$errfile 1>>$logfile || { echo "minio-java run failed."; rc=1; }
    return $rc
}

# invoke the script
main "$@"