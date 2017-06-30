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

# Check script argument
[[ $# -ne 1 ]] && echo "USAGE: ./run.sh /path/to/target/app" && exit -1

APPNAME=$1

pid=0

sig_handler() {
    echo "EXIT signal captured.."
    kill $pid
    exit -1
}

# Register signal handler
trap 'sig_handler' SIGINT

# Run application in background
$APPNAME &
pid=$!

wait $pid
