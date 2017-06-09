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

# Add Run instructions for SDK tests here.
# The Build instructions should be added to build.sh

setLogEnv() {
    export INIT_LOG_DIRECTORY=$(echo $LOG_DIRECTORY/${PWD##*/})
    export INIT_ERROR_LOG_FILE=$(echo $INIT_LOG_DIRECTORY/"error.log")
    export INIT_LOG_FILE=$(echo $INIT_LOG_DIRECTORY/"run.log")
}

prepareLogDir() {
    # clear old logs 
    rm -r echo $INIT_LOG_DIRECTORY 2> /dev/null

    # create log directory
    mkdir $INIT_LOG_DIRECTORY 2> /dev/null

    # create log files
    touch $INIT_ERROR_LOG_FILE
    touch $INIT_LOG_FILE
}

initCheck() {
  # Get minio admin package  
  go get -u github.com/minio/minio/pkg/madmin
  
  # Build the endpoint checker program.
  go build initCheck.go

  # This is to avoid https://github.com/docker/docker/issues/9547
  sync
  
  # Run the check
  ./initCheck
}

# Set log folders
setLogEnv

# Create log folders
prepareLogDir

# Run the init tests
initCheck