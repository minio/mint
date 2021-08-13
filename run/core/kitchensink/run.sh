#!/bin/bash
#
#  Minio Cloud Storage, (C) 2021 Minio, Inc.
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

if [ $# -ne 2 ]; then
    echo "usage: run.sh <OUTPUT-LOG-FILE> <ERROR-LOG-FILE>"
    exit 1
fi

output_log_file="$1"
error_log_file="$2"

BUCKET="my-test-bucket"

(./kitchensink create $SERVER_ENDPOINT $ACCESS_KEY $SECRET_KEY $BUCKET  1>>"$output_log_file" 2>"$error_log_file")
rv=$?
if [ "$rv" -ne 0 ]; then
    exit 1
    
./kitchensink verify $SERVER_ENDPOINT $ACCESS_KEY $SECRET_KEY $BUCKET  1>>"$output_log_file" 2>"$error_log_file"
