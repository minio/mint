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

_init() {
    AWS="aws --endpoint-url $1"
}

create_random_string() {
    random_str=$(tr -dc 'a-z0-9' < /dev/urandom  | fold -w 32 | head -n 1)
    echo "$random_str"
}

createBucket_01(){
    local bucketName
    
    bucketName=$(create_random_string)

    echo "Running createBucket_01"
    # Create a bucket.
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # Stat a bucket.
    echo "Testing if the bucket was indeed created"
    ${AWS} s3api head-bucket --bucket "${bucketName}"

    echo "Removing the bucket"
    # remove bucket
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

createObject_02(){
    echo "Running createObject_02"
    local bucketName

    bucketName=$(create_random_string)

    # Create a temp 2m file
    echo "Creating a 2mb temp file for upload"
    tmpfile=`tempfile`
    truncate -s 2m "$tmpfile"

    # save md5 hash
    hash1=`md5sum $tmpfile | awk '{print $1}'`

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # copy the file
    echo "Upload an object"
    ${AWS} s3api put-object --body "${tmpfile}" --bucket "${bucketName}" --key "$(basename "${tmpfile}")"

    echo "Download the file"
    ${AWS} s3api get-object --bucket "${bucketName}" --key "$(basename "${tmpfile}")" "${tmpfile}_downloaded"

    #save md5 hash of downloaded file
    hash2=$(md5sum "/tmp/${tmpfile}_downloaded" | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file"
    if [ "$hash1" -ne "$hash2" ]; then
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "$(basename "${tmpfile}")"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

main() {
    # Run tests
    createBucket_01
    createObject_02
}

_init "$@" && main
