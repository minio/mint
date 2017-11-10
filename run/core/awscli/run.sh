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

# handle command line arguments
if [ $# -ne 2 ]; then
    echo "usage: run.sh <OUTPUT-LOG-FILE> <ERROR-LOG-FILE>"
    exit -1
fi

output_log_file="$1"
error_log_file="$2"

# configure awscli
aws configure set aws_access_key_id "$ACCESS_KEY"
aws configure set aws_secret_access_key "$SECRET_KEY"
aws configure set default.region "$SERVER_REGION"

# Setup endpoint scheme
endpoint="http://$SERVER_ENDPOINT"
if [ "$ENABLE_HTTPS" -eq 1 ]; then
    endpoint="https://$SERVER_ENDPOINT"
fi

# check the access style and run tests if virtual style is set
if [ "$ENABLE_VIRTUAL_STYLE" -eq 1 ]; then
    aws configure set default.s3.addressing_style virtual
    ./test.sh "$endpoint"  1>>"$output_log_file" 2>"$error_log_file"
fi

# run path style tests
aws configure set default.s3.addressing_style path
./test.sh "$endpoint"  1>>"$output_log_file" 2>"$error_log_file"
