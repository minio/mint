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

create_random_string() {
    random_str=$(tr -dc 'a-z0-9' < /dev/urandom  | fold -w 32 | head -n 1)
    echo "$random_str"
}

remove_bucket() {
    ./mc rm --force --recursive "target/$1"
    rm -rf /tmp/*
}

# Create a bucket and check if it exists on server
makeBucket(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # mc returns status 0 if bucket is created
    if [ "$(./mc mb "target/${bucketName}")" -ne 0 ]; then
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Upload an object, download it and check if it matches the uploaded object
putObject(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    # save md5 hash
    hash1=$(md5sum "$DATA_DIR"/datafile-1-MB | awk '{print $1}')
    

    # upload the file
    echo "Uploading the 1MB temp file" 
    ./mc cp "$DATA_DIR"/datafile-1-MB "target/${bucketName}" 

    echo "Download the file"      
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-1-MB" /tmp | jq -r .target)")" != "datafile-1-MB" ]; then
        return 1
    fi

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file" 
    if [ "$hash1" != "$hash2" ]; then 
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Upload an object > 64MB (MC uses multipart for more than 64MB), download it and check if it matches the uploaded object
putObjectMultipart(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    # save md5 hash
    hash1=$(md5sum "$DATA_DIR"/datafile-65-MB | awk '{print $1}')

    # upload the file
    echo "Uploading a 65MB temp file" 
    ./mc cp "$DATA_DIR"/datafile-65-MB "target/${bucketName}" 

    echo "Download the file" 
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-65-MB" /tmp | jq -r .target)")" != "datafile-65-MB" ]; then
        return 1
    fi 
    
    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-65-MB | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file" 
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Tests `mc mirror` by mirroring all the local content to remove bucket.
mirrorObject() {
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    echo "Upload a set of files"
    ./mc mirror -q "$DATA_DIR" "target/${bucketName}"   

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Tests for presigned URL upload success case, presigned URL
# is correct and accessible - we calculate md5sum of
# the object and validate it against a local files md5sum.
presignedUploadObject() {
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    fileName="${DATA_DIR}/datafile-1-MB"

    # save md5 hash
    hash1=$(md5sum "$fileName" | awk '{print $1}')

    # create presigned URL object
    echo "Create presigned file upload" 
    url=$(./mc share --json upload "target/${bucketName}/$(basename "$fileName")" | jq -r .share)
    
    # upload the file
    curlUrl=$($url | sed "s@<FILE>@$fileName@g")

    eval "$curlUrl"

    echo "Download the file"      
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-1-MB" /tmp | jq -r .target)")" != "datafile-1-MB" ]; then
        return 1
    fi

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file" 
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Tests for presigned URL download success case, presigned URL
# is correct and accessible - we calculate md5sum of
# the object and validate it against a local files md5sum.
presignedDownloadObject(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    fileName="${DATA_DIR}/datafile-1-MB"

    # save md5 hash
    hash1=$(md5sum "$fileName" | awk '{print $1}')

    # upload the file
    echo "Uploading a 1MB temp file" 
    ./mc cp "${fileName}" "target/${bucketName}" 

    # create presigned URL download
    echo "Create presigned file download URL" 
    url=$(./mc share --json download "target/${bucketName}/$(basename "$fileName")" | jq -r .share)
    
    # download the file
    echo "Download the file"
    curl "$url" -o /tmp/"$(basename "$fileName")"

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/"$(basename "$fileName")" | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file" 
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Upload an object, with invalid object name
putObjectError(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

     # Make bucket
    ./mc mb "target/${bucketName}" 

    # upload the file
    echo "Uploading file with invalid object name" 
    ./mc cp "$DATA_DIR"/datafile-1-MB "target/${bucketName}//2123123\123" 

    # mc returns status 1 if case of invalid object name
    if [ $? -ne 1 ]; then
        return 1
    fi

    # remove bucket and cleanup
    echo "Removing the bucket" 
    remove_bucket "${bucketName}"
}

# Create a bucket and check if it exists on server
makeBucketError(){
    # Make bucket
    local bucketName
    bucketName="Abcd"

     # Make bucket
    ./mc mb "target/${bucketName}" 

    # mc returns status 1 if bucket is created
    if [ $? -ne 1 ]; then
        return 1
    fi

}

# main handler for all the tests.
main() {
    # Succes tests
    makeBucket
    putObject
    putObjectMultipart
    mirrorObject
    presignedUploadObject
    presignedDownloadObject

    # TODO Add Policy tests once supported on GCS

    # Error tests
    putObjectError
    makeBucketError
}

main
