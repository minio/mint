#!/bin/bash
#  Mint (C) 2017 Minio, Inc.
#

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
    ./mc rm --force --recursive "target/$1" > /dev/null
    rm -rf /tmp/*
}

# Create a bucket and check if it exists on server
makeBucket(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # mc returns status 0 if bucket is created
    printf "\tEntering Make Bucket Test 1\n"

    if ! ./mc mb "target/${bucketName}" > /dev/null
    then 
        >&2 echo "Make Bucket Test 1 Failure"
        return 1
    fi

    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Upload an object, download it and check if it matches uploaded object
putObject(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # Make bucket
    ./mc mb "target/${bucketName}" > /dev/null

    # save md5 hash
    hash1=$(md5sum "$MINT_DATA_DIR"/datafile-1-MB | awk '{print $1}')

    # upload the file
    ./mc cp "$MINT_DATA_DIR"/datafile-1-MB "target/${bucketName}" > /dev/null

    printf "\tEntering Put Object Test 1\n"
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-1-MB" /tmp | jq -r .target)")" != "datafile-1-MB" ]; then
        >&2 echo "Put Object Test 1 Failure"
        return 1
    fi

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')

    printf "\tEntering Put Object Test 2\n"
    if [ "$hash1" != "$hash2" ]; then 
        >&2 echo "Put Object Test 2 Failure"
        return 1
    fi

    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Upload an object > 64MB (MC uses multipart for more than 64MB), download it and check if it matches the uploaded object
putObjectMultipart(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # Make bucket
    ./mc mb "target/${bucketName}" > /dev/null

    # save md5 hash
    hash1=$(md5sum "$MINT_DATA_DIR"/datafile-65-MB | awk '{print $1}')

    # upload the file
    ./mc cp "$MINT_DATA_DIR"/datafile-65-MB "target/${bucketName}" > /dev/null

    printf "\tEntering Put Object Multipart Test 1\n"
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-65-MB" /tmp | jq -r .target)")" != "datafile-65-MB" ]; then
        >&2 echo "Put Object Multipart Test 1 Failure"
        return 1
    fi 

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-65-MB | awk '{print $1}')

    printf "\tEntering Put Object Multipart Test 2\n"
    if [ "$hash1" != "$hash2" ]; then
        >&2 echo "Put Object Multipart Test 2 Failure"
        return 1
    fi

    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Tests for presigned URL upload success case, presigned URL
# is correct and accessible - we calculate md5sum of
# the object and validate it against a local files md5sum.
presignedUploadObject() {
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # Make bucket
    ./mc mb "target/${bucketName}" > /dev/null

    fileName="${MINT_DATA_DIR}/datafile-1-MB"

    # save md5 hash
    hash1=$(md5sum "$fileName" | awk '{print $1}')

    # create presigned URL object
    url=$(./mc share --json upload "play/${bucketName}/$(basename "$fileName")" | jq -r .share | sed "s|<FILE>|$fileName|g" | sed "s|curl||g")
    url="curl -sS $url"
    ./mc policy upload "target/${bucketName}" > /dev/null

    eval "$url"> /dev/null

    printf "\tEntering Share Upload Test 1\n"
    if [ "$(basename "$(./mc cp --json "target/${bucketName}/datafile-1-MB" /tmp/ | jq -r .target)")" != "datafile-1-MB" ]; then
        printf "\tShare Upload Test 1 Failure\n"
        >&2 echo "Presigned Upload Test 1 Failure"
        >&2 echo "Error on line 164"
        return 1
    fi

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')

    printf "\tEntering Share Upload Test 2\n"
    if [ "$hash1" != "$hash2" ]; then
        printf "\tShare Upload Test 2 Faillure\n"
        >&2 echo "Share Upload Test 2 Failure"
        >&2 echo "Error on line 171"
        return 1
    fi

    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Tests for presigned URL download success case, presigned URL
# is correct and accessible - we calculate md5sum of
# the object and validate it against a local files md5sum.
presignedDownloadObject(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # Make bucket
    ./mc mb "target/${bucketName}" > /dev/null

    fileName="${MINT_DATA_DIR}/datafile-1-MB"

    # save md5 hash
    hash1=$(md5sum "$fileName" | awk '{print $1}')

    # upload the file
    ./mc cp "${fileName}" "target/${bucketName}" > /dev/null

    ./mc policy download "target/${bucketName}" > /dev/null
    # create presigned URL download
    url=$(./mc share --json download "target/${bucketName}/$(basename "$fileName")" | jq -r .share)

    # download the file
    curl -sS -X GET "$url" > /tmp/datafile-1-MB

    # calculate the md5 hash of downloaded file
    hash2=$(md5sum /tmp/"$(basename "$fileName")" | awk '{print $1}')

    printf "\tEntering Presigned Download Object Test 1\n"
    if [ "$hash1" != "$hash2" ]; then
        printf "\tShare Download Test 1 Faillure\n"
        >&2 echo "Share Download Test 1 Failure"
        return 1
    fi

    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Tests for list object success, by mirroring
# an fs store to minio, and then comparing 
# the ls results.
MirrorListObjects(){
    local bucketName
    bucketName=$(create_random_string)

    # create a new bucket and mirror all content into said bucket
    ./mc mb "target/${bucketName}" > /dev/null
    ./mc mirror -q "$MINT_DATA_DIR" "target/${bucketName}" > /dev/null


    # ignore all white space related differences when comparing using diff
    if ! diff -bB <(ls "$MINT_DATA_DIR") <(./mc ls --json "target/${bucketName}" | jq -r .key)
    then
        printf "\tList Objects Test 1 Failure\n"
        >&2 echo "List Objects Test 1 Failure"
    fi

    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Tests for the cat object success by comparing 
# the STDout of "cat run.sh", to "mc cat run.sh"
# by uploading run.sh, and comparing the two
# outputs.
catObjects(){
    local bucketName
    bucketName=$(create_random_string)

    ./mc mb "target/${bucketName}" > /dev/null

    # substitute run.sh as a txt file in upload
    ./mc cp ./run.sh "target/${bucketName}" > /dev/null

    # compare output
    if !  diff <(cat ./run.sh) <(./mc cat "target/${bucketName}/run.sh")
    then
        printf "\tCat Objects Test 1 Failure\n"
        >&2 echo "Cat Objects Test 1 Failure"
    fi
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Tests for mc watch, by running mc watch in 
# the background, creating several objects,
# checking to see if the 
# words "ObjectCreated" and "ObjectRemoved", 
# were printed to STDout,
watchObjects(){
    local bucketName
    bucketName=$(create_random_string)
    ./mc mb "target/$bucketName" >/dev/null

    # send all output to a variable myvar, and run the operation
    # in the background
    ./mc watch --json "target/$bucketName" > myvar&
    processID=$!

    # these operations should cause mc watch to print "ObjectCreated and ObjectRemoved"
    ./mc cp "$MINT_DATA_DIR/datafile-1-b play/$bucketName">/dev/null
    ./mc rm --force --recursive "target/$bucketName">/dev/null

    # run diff with flags -bB, to ignore whitespace differences

    if ! diff -bB <(jq -r .events.type myVar | tr '\n' ' ') <(printf "ObjectCreated ObjectRemoved")
    then
        printf "\tWatch Objects Test 1 Error\n"
        >&2 echo "Watch Objects Test 1 Error"
    fi

    # kill routine running in the background
    pkill $processID
    # remove the extraneous variable which was created
    rm myVar
    printf "\tTest Success\n"
}

# Upload an object, with invalid object name
putObjectError(){
    # Make bucket
    local bucketName
    bucketName=$(create_random_string)

    # Make bucket
    ./mc mb "target/${bucketName}" > /dev/null

    # upload the file
    printf "\tEntering Put Object Error Test 1\n"

    # mc returns status 1 if case of invalid object name
    if ./mc cp "$MINT_DATA_DIR"/datafile-1-MB "target/${bucketName}//2123123\123" > /dev/null 2>&1
    then
        printf "\tPut Object Error Test 1 Failure\n"
        >&2 echo "Put Object Error Test 1 Failure"
        >&2 echo "Error on line 264"
        return 1
    fi


    # remove bucket and cleanup
    remove_bucket "${bucketName}"
    printf "\tTest Success\n"
}

# Create a bucket and check if it exists on server
makeBucketError(){
    # Make bucket
    local bucketName
    bucketName="Abcd"

    # Make bucket
    printf "\tEntering Make Bucket Error Test 1\n"

    # mc returns status 1 if case of invalid object name
    if ./mc mb "target/${bucketName}" > /dev/null 2>&1
    then
        printf "\tMake Bucket Error Test 1 Failure\n"
        >&2 echo "Make Bucket Error Test 1 Failure"
        >&2 echo "Error on line 280"
        return 1
    fi

    printf "\tTest Success\n"
}

# main handler for all the tests.
main() {
    blue=$(tput setaf 4)
    normal=$(tput sgr0)

    # Succes tests
    printf "\n %s Make Bucket Tests %s \n\n" "${blue}" "${normal}"
    makeBucket
    printf "\n %s Put Object Tests %s \n\n" "${blue}" "${normal}" 
    putObject
    printf "\n %s Put Object Multipart Tests %s \n\n" "${blue}" "${normal}" 
    putObjectMultipart
    printf "\n %s Presigned Upload Object Tests %s \n\n" "${blue}" "${normal}"
    presignedUploadObject
    printf "\n %s Presigned Download Object Tests %s \n\n" "${blue}" "${normal}" 
    presignedDownloadObject
    printf "\n %s List and Mirror Object Tests %s \n\n" "${blue}" "${normal}"
    MirrorListObjects
    printf "\n %s Cat Object Tests %s \n\n" "${blue}" "${normal}"
    catObjects
    printf "\n %s Watch  Object Tests %s \n\n" "${blue}" "${normal}"
    #watchObjects 

    # TODO Add Policy tests once supported on GCS

    # Error tests
    printf "\n %s Put Object Error Tests %s \n\n" "${blue}" "${normal}" 
    putObjectError
    printf "\n %s Make Bucket Object Error Tests %s \n\n" "${blue}" "${normal}" 
    makeBucketError
    printf "\n %s End of tests %s \n\n" "${blue}" "${normal}" 1
}

main
