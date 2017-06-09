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

setLogEnv() {
    export MC_LOG_DIRECTORY=$(echo $LOG_DIRECTORY/${PWD##*/})
    export MC_ERROR_LOG_FILE=$(echo $MC_LOG_DIRECTORY/"error.log")
    export MC_LOG_FILE=$(echo $MC_LOG_DIRECTORY/"run.log")

    if [ $ENABLE_HTTPS -eq "1" ]; then
        export TARGET_ADDRESS_WITH_PROTOCOL="https://"$S3_ADDRESS
    else
        export TARGET_ADDRESS_WITH_PROTOCOL="http://"$S3_ADDRESS
    fi
}

prepareLogDir() {
    # clear old logs 
    rm -r echo $MC_LOG_DIRECTORY 2> /dev/null

    # create log directory
    mkdir $MC_LOG_DIRECTORY 2> /dev/null

    # create log files
    touch $MC_ERROR_LOG_FILE
    touch $MC_LOG_FILE
}

cleanUP(){
    # remove mc 
    rm mc
}

downloadMC() {
    # Download latest MC release
    curl -s -o mc https://dl.minio.io/client/mc/release/linux-amd64/mc
    res=$?
    if test "$res" != "0"; then
        echo "curl command to download mc failed with: $res" >> $MC_ERROR_LOG_FILE
        exit 1
    else 
        chmod +x ./mc
        echo "Downloaded mc $(./mc version | grep Version)" >> $MC_LOG_FILE
        echo "Adding mc host alias target $TARGET_ADDRESS_WITH_PROTOCOL" >> $MC_LOG_FILE
        ./mc config host add target $TARGET_ADDRESS_WITH_PROTOCOL $ACCESS_KEY $SECRET_KEY >> $MC_LOG_FILE
    fi
}

# Execute test.sh 
runMCTests() {
    ./test.sh
}

# Setup log directories 
setLogEnv

# Create the log dir and files
prepareLogDir

# Download and add alias target pointing to the server under test
downloadMC

# run the tests
runMCTests

# Remove mc binary
cleanUP

