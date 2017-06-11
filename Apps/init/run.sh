#!/bin/sh
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


# setLogEnv() {
#     export INIT_LOG_DIR=$(echo ../../$LOG_DIR/${PWD##*/})
#     export INIT_ERROR_LOG_FILE=$(echo $INIT_LOG_DIR/"error.log")
#     export INIT_LOG_FILE=$(echo $INIT_LOG_DIR/"output.log")
# }

# prepareLogDir() {
#     # clear old logs 
#     rm -r echo $INIT_LOG_DIR 2> /dev/null

#     # create log directory
#     mkdir $INIT_LOG_DIR 2> /dev/null

#     # create log files
#     touch $INIT_ERROR_LOG_FILE
#     touch $INIT_LOG_FILE
# }

cleanUp() {
    # remove executable 
    rm initCheck 2> /dev/null
}

build() {
    go build -o initCheck ./initCheck.go
}

# Set log folders
setLogEnv

# Create log folders
prepareLogDir

# Run the init tests
initCheck