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

createBuckets_01(){
    echo "Running createBuckets_01" 
    # Make bucket
    ./mc mb target/testbucket1 

    echo "Testing if the bucket was created" 
    # list buckets
    ./mc ls target 

    echo "Removing the bucket" 
    # remove bucket
    ./mc rm target/testbucket1 
}

createFile_02(){
    echo "Running createFile_02" 

    # Create a temp 2m file
    echo "Creating a 2mb temp file for upload" 
    truncate -s 2m /tmp/file

    # create a bucket
    echo "Creating a bucket" 
    ./mc mb target/testbucket1 

    # copy the file
    echo "Uploading the 2mb temp file" 
    ./mc cp /tmp/file target/testbucket1 

    echo "Download the file" 
    ./mc cp target/testbucket1/file /tmp/file_downloaded 
    
    echo "Testing if the downloaded file is same as local file" 
    comm /tmp/file_downloaded /tmp/file 

    echo "Removing the bucket" 
    # remove bucket
    ./mc rm --force --recursive target/testbucket1 
}

createBuckets_01
createFile_02