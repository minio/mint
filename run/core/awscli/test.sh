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

# Tests creating, stat and delete on a bucket.
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

# Tests creating and deleting an object.
createObject_02(){
    echo "Running createObject_02"
    local bucketName

    bucketName=$(create_random_string)

    # save md5 hash
    hash1=$(md5sum "${MINT_DATA_DIR}/datafile-1-MB" | awk '{print $1}')

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # copy the file
    echo "Upload an object"
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "datafile-1-MB"

    echo "Download the file"
    ${AWS} s3api get-object --bucket "${bucketName}" --key "datafile-1-MB" "/tmp/datafile-1-MB-downloaded"

    #save md5 hash of downloaded file
    hash2=$(md5sum "/tmp/datafile-1-MB-downloaded" | awk '{print $1}')

    echo "Testing if the downloaded file is same as local file"
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "datafile-1-MB"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# Tests listing objects for both v1 and v2 API.
listObjects_03() {
    echo "Running listObjects_03"
    local bucketName
    local fileName
    local baseFileName

    bucketName=$(create_random_string)

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    fileName="${MINT_DATA_DIR}/datafile-1-MB"
    baseFileName="$(basename "${fileName}")"

    # upload a file
    echo "Upload an object"
    ${AWS} s3api put-object --body "${fileName}" --bucket "${bucketName}" --key "${baseFileName}"

    keyName=$(${AWS} s3api list-objects --bucket "${bucketName}" --prefix "${baseFileName}" | jq -r .Contents[].Key)
    if [ "$keyName" != "datafile-1-MB" ]; then
        echo "Unexpected $keyName, expecting datafile-1-MB"
        return 1
    fi

    keyName=$(${AWS} s3api list-objects --bucket "${bucketName}" --prefix "linux" | jq -r .Contents[].Key)
    if [ "$keyName" != "" ]; then
        echo "Unexpected $keyName found, expecting empty key"
        return 1
    fi

    keyName=$(${AWS} s3api list-objects-v2 --bucket "${bucketName}" --prefix "${baseFileName}" | jq -r .Contents[].Key)
    if [ "$keyName" != "datafile-1-MB" ]; then
        echo "Unexpected $keyName, expecting datafile-1-MB"
        return 1
    fi

    keyName=$(${AWS} s3api list-objects-v2 --bucket "${bucketName}" --prefix "linux" | jq -r .Contents[].Key)
    if [ "$keyName" != "" ]; then
        echo "Unexpected $keyName found, expecting empty key"
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "${baseFileName}"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# Tests multipart API by making each individual calls.
multipart_04() {
    echo "Running multipart_04"
    local bucketName
    local fileName1
    local fileName2

    bucketName=$(create_random_string)
    objectName=$(create_random_string)
    fileName1="${MINT_DATA_DIR}/datafile-5-MB"
    fileName2="${MINT_DATA_DIR}/datafile-1-MB"

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # create multipart
    uploadID=$(${AWS} s3api create-multipart-upload --bucket "${bucketName}" --key "${objectName}" | jq -r .UploadId)

    # Capture etag for part-number 1
    etag1=$(${AWS} s3api upload-part --bucket "${bucketName}" --key "${objectName}" --body "${fileName1}" --upload-id "${uploadID}" --part-number 1 | jq -r .ETag)

    # Capture etag for part-number 2
    etag2=$(${AWS} s3api upload-part --bucket "${bucketName}" --key "${objectName}" --body "${fileName2}" --upload-id "${uploadID}" --part-number 2 | jq -r .ETag)

    # Create a multipart struct file for completing multipart transaction
    echo "{
        \"Parts\": [
            {
                \"ETag\": ${etag1},
                \"PartNumber\": 1
            },
            {
                \"ETag\": ${etag2},
                \"PartNumber\": 2
            }
        ]
    }" >> /tmp/multipart

    # Use saved etags to complete the multipart transaction
    finalETag=$(${AWS} s3api complete-multipart-upload --multipart-upload file:///tmp/multipart --bucket "${bucketName}" --key "${objectName}" --upload-id "${uploadID}" | jq -r .ETag)
    if [ "${finalETag}" == "" ]; then
        echo "Unexpected empty etag, expecting a non etag"
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "${objectName}"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# Copy object tests for server side copy
# of the object, validates returned md5sum.
copyObject_05() {
    echo "Running copyObject_05"
    local bucketName

    bucketName=$(create_random_string)

    # save md5 hash
    hash1=$(md5sum "${MINT_DATA_DIR}/datafile-1-MB" | awk '{print $1}')

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # upload an object
    echo "Upload an object"
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "datafile-1-MB"

    # copy object server side
    echo "Copy server side object"
    hash2=$(${AWS} s3api copy-object --bucket "${bucketName}" --key "datafile-1-MB-copy" --copy-source "${bucketName}/datafile-1-MB" | jq -r .CopyObjectResult.ETag | sed -e 's/^"//' -e 's/"$//')

    echo "Testing if the copied object is same as source"
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "datafile-1-MB"

    echo "Remove the copied object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "datafile-1-MB-copy"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# Tests for presigned URL success case, presigned URL
# is correct and accessible - we calculate md5sum of
# the object and validate it against a local files md5sum.
presignedObject_06() {
    echo "Running presignedObject_06"
    local bucketName

    bucketName=$(create_random_string)

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # save md5 hash
    hash1=$(md5sum "${MINT_DATA_DIR}/datafile-1-MB" | awk '{print $1}')

    # upload an object
    echo "Upload an object"
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "datafile-1-MB"

    url=$(${AWS} s3 presign "s3://${bucketName}/datafile-1-MB")
    hash2=$(curl "${url}" | md5sum -)

    echo "Testing if the copied object is same as source"
    if [ "$hash1" != "$hash2" ]; then
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "datafile-1-MB"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# Tests `aws s3 cp` by uploading a local file.
copy_07() {
    local bucketName
    local objectName

    bucketName=$(create_random_string)
    objectName=$(create_random_string)
    fileName="${MINT_DATA_DIR}/datafile-65-MB"

    echo "Running copy_05"
    # Create a bucket.
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    echo "Upload a large file"
    ${AWS} s3 cp "$fileName" "s3://${bucketName}/$(basename "$fileName")"

    echo "Remove an object"
    ${AWS} s3 rm "s3://${bucketName}/$(basename "$fileName")"

    echo "Remove bucket"
    ${AWS} s3 rb "s3://${bucketName}/"
}

# Tests `aws s3 sync` by mirroring all the
# local content to remove bucket.
sync_08() {
    local bucketName
    local objectName

    bucketName=$(create_random_string)
    objectName=$(create_random_string)

    echo "Running sync_06"
    # Create a bucket.
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    echo "Upload a set of files"
    ${AWS} s3 sync "$MINT_DATA_DIR" "s3://${bucketName}/"

    echo "Remove bucket contents recursively"
    ${AWS} s3 rm --recursive "s3://${bucketName}/"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# list objects negative test - tests for following conditions.
# v1 API with max-keys=-1 and max-keys=0
# v2 API with max-keys=-1 and max-keys=0
listObjects_error_01() {
    echo "Running listObjects_03"
    local bucketName
    local fileName
    local baseFileName

    bucketName=$(create_random_string)
    fileName="${MINT_DATA_DIR}/datafile-1-MB"
    baseFileName="$(basename "${fileName}")"

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # upload a file
    echo "Upload an object"
    ${AWS} s3api put-object --body "${fileName}" --bucket "${bucketName}" --key "${baseFileName}"

    # Server replies an error for v1 with max-key=-1
    ${AWS} s3api list-objects --bucket "${bucketName}" --prefix "${baseFileName}" --max-keys=-1
    if [ $? -ne 255 ]; then
        return 1
    fi

    # Server replies an error for v2 with max-keys=-1
    ${AWS} s3api list-objects-v2 --bucket "${bucketName}" --prefix "${baseFileName}" --max-keys=-1
    if [ $? -ne 255 ]; then
        return 1
    fi

    # Server returns success with no keys when max-keys=0
    if [ "$(${AWS} s3api list-objects-v2 --bucket "${bucketName}" --prefix "${baseFileName}" --max-keys=0)" -ne 0 ]; then
        return 1
    fi

    echo "Remove the object"
    ${AWS} s3api delete-object --bucket "${bucketName}" --key "${baseFileName}"

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# put object negative test - tests for following conditions.
# - invalid object name.
# - invalid Content-Md5
# - invalid Content-Length
putObject_error_02() {
    echo "Running putObject_error_02"
    local bucketName

    bucketName=$(create_random_string)

    # save md5 hash
    hash1=$(md5sum "${MINT_DATA_DIR}/datafile-1-MB" | awk '{print $1}')

    # create a bucket
    echo "Create a bucket"
    ${AWS} s3api create-bucket --bucket "${bucketName}"

    # upload an object failure with invalid object name.
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "/2123123\123"
    if [ $? -ne 255 ]; then
        return 1
    fi

    # upload an object without content-md5.
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "datafile-1-MB" --content-md5 "invalid"
    if [ $? -ne 255 ]; then
        return 1
    fi

    # upload an object without content-length.
    ${AWS} s3api put-object --body "${MINT_DATA_DIR}/datafile-1-MB" --bucket "${bucketName}" --key "datafile-1-MB" --content-length -1
    if [ $? -ne 255 ]; then
        return 1
    fi

    echo "Remove the bucket"
    ${AWS} s3api delete-bucket --bucket "${bucketName}"
}

# main handler for all the tests.
main() {
    # Success tests
    createBucket_01
    createObject_02
    listObjects_03
    multipart_04
    copyObject_05
    presignedObject_06

    # Success cli ops.
    copy_07
    sync_08

    # Error tests
    listObjects_error_01
    putObject_error_02
}

_init "$@" && main
