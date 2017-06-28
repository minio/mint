#!/usr/bin/env bash
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

create_random_string() {
    bucketName=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 32 | head -n 1)
}

createBuckets_01(){
    create_random_string
    echo "Running createBuckets_01" 

    # Make bucket
    ./mc mb "target/${bucketName}" 

    echo "Testing if the bucket was created" 
    # list buckets
    ./mc ls target 

    echo "Removing the bucket" 
    # remove bucket
    ./mc rm "target/${bucketName}" 
}

createFile_02(){
    create_random_string
    echo "Running createFile_02" 

    # save md5 hash
    hash1=$(md5sum "$DATA_DIR"/datafile-1-MB | awk '{print $1}')
    
    # create a bucket
    echo "Creating a bucket" 
    ./mc mb "target/${bucketName}" 

    # copy the file
    echo "Uploading the 1MB temp file" 
    ./mc cp "$DATA_DIR"/datafile-1-MB "target/${bucketName}" 

    echo "Download the file" 
    ./mc cp "target/${bucketName}/datafile-1-MB" /tmp/datafile-1-MB-downloaded 
    
    #save md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-1-MB-downloaded | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file" 
    if [ "${hash1}" -ne "${hash2}" ]; then
        return 1
    fi

    echo "Removing the bucket" 
    # remove bucket
    ./mc rm --force --recursive "target/${bucketName}" 
}

# Run tests
createBuckets_01
createFile_02