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

MC_CMD='./mc'
SERVER_ALIAS="target"
HASH_1_MB=$(md5sum "${MINT_DATA_DIR}/datafile-1-MB" | awk '{print $1}')
HASH_65_MB=$(md5sum "${MINT_DATA_DIR}/datafile-65-MB" | awk '{print $1}')

function get_time() {
    date +%s%N
}

function get_duration() {
    start_time=$1
    end_time=$(get_time)

    echo $(( (end_time - start_time) / 1000000 ))
}

function log_success() {
    function=$(python -c 'import sys,json; print(json.dumps(sys.stdin.read()))' <<<"$2")
    printf '{"name": "mc", "duration": %d, "function": %s, "status": "PASS"}\n' "$1" "$function"
}

function log_failure() {
    function=$(python -c 'import sys,json; print(json.dumps(sys.stdin.read()))' <<<"$2")
    printf '{"name": "mc", "duration": %d, "function": %s, "status": "FAIL", "error": "%s"}\n' "$1" "$function" "$3"
}

function log_alert() {
    function=$(python -c 'import sys,json; print(json.dumps(sys.stdin.read()))' <<<"$2")
    printf '{"name": "mc", "duration": %d, "function": %s, "status": "FAIL", "alert": "%d", error": "%s"}\n' "$1" "$function" "$3" "$4"
}

function make_bucket() {
    # Make bucket
    bucket_name="mc-mint-test-bucket-$RANDOM"
    function="${MC_CMD} mb ${SERVER_ALIAS}/${bucket_name}"

    # execute the test
    out=$($function 2>&1)
    rv=$?

    # if command is successful print bucket_name or print error
    if [ $rv -eq 0 ]; then
        echo "${bucket_name}"
    else    
        echo "${out}"
    fi

    return $rv
}

function delete_bucket() {
    # Delete bucket
    function="${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/${1}"
    out=$($function 2>&1)
    rv=$?

    # echo the output
    echo "${out}"

    return $rv
}

# Create a bucket and check if it exists on server
function test_make_bucket() {

    # log start time
    start_time=$(get_time)

    function="make_bucket"
    test_function=${function}
    bucket_name=$(make_bucket)
    rv=$?

    # if make_bucket is successful remove the bucket
    if [ $rv -eq 0 ]; then
        function="delete_bucket"
        out=$(delete_bucket "${bucket_name}") 
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Upload an object, download it and check if it matches uploaded object
function test_put_object() {

    # log start time
    start_time=$(get_time)

    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$?

    # if make bucket succeeds upload a file
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} cp ${MINT_DATA_DIR}/datafile-1-MB ${SERVER_ALIAS}/${bucket_name}"
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    # if upload succeeds download the file
    if [ $rv -eq 0 ]; then 
        function="${MC_CMD} cp --json ${SERVER_ALIAS}/${bucket_name}/datafile-1-MB /tmp"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?        
        # calculate the md5 hash of downloaded file
        hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')
    fi

    # if download succeeds, verify downloaded file
    if [ $rv -eq 0 ]; then 
        if [ "$HASH_1_MB" == "$hash2" ] && [ "$(basename "$(echo "$out" | jq -r .target)")" == "datafile-1-MB" ]; then
            function="delete_bucket"
            out=$(delete_bucket "$bucket_name") 
            rv=$?
            # remove download file
            rm -rf /tmp/datafile-1-MB
        else
            rv=1
        fi
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Upload an object > 64MB (MC uses multipart for more than 64MB), download
# it and check if it matches the uploaded object
function test_put_object_multipart() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$?

    # if make bucket succeeds upload a file
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} cp ${MINT_DATA_DIR}/datafile-65-MB ${SERVER_ALIAS}/${bucket_name}"
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    # if upload succeeds download the file
    if [ $rv -eq 0 ]; then 
        function="${MC_CMD} cp --json ${SERVER_ALIAS}/${bucket_name}/datafile-65-MB /tmp"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?        
        # calculate the md5 hash of downloaded file
        hash2=$(md5sum /tmp/datafile-65-MB | awk '{print $1}')
    fi

    # if download succeeds, verify downloaded file and cleanup
    if [ $rv -eq 0 ]; then 
        if [ "$HASH_65_MB" = "$hash2" ] && [ "$(basename "$(echo "$out" | jq -r .target)")" == "datafile-65-MB" ]; then
            function="delete_bucket"
            out=$(delete_bucket "$bucket_name") 
            rv=$?        
            # remove download file
            rm -rf /tmp/datafile-65-MB
        else
            rv=1
        fi
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Tests for presigned URL upload success case, presigned URL is correct and
# accessible - we calculate md5sum of the object and validate
# it against a local files md5sum.
function test_presigned_upload_object() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$?

    # if make bucket succeeds, create a share upload
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} share --json upload ${SERVER_ALIAS}/${bucket_name}/datafile-1-MB"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi


    # if share upload succeeds, upload the file via curl and then download with mc cp
    if [ $rv -eq 0 ]; then
        url=$(echo "$out" | jq -r .share | sed "s|<FILE>|$MINT_DATA_DIR/datafile-1-MB|g" | sed "s|curl||g")
        url="curl -sS $url"
        # upload the file
        eval "$url"> /dev/null
        # download the file
        function="${MC_CMD} cp --json ${SERVER_ALIAS}/${bucket_name}/datafile-1-MB /tmp"
        out=$($function 2>&1)
        rv=$?
        # calculate the md5 hash of downloaded file
        hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')
    fi

    # if download succeeds, verify and cleanup
    if [ $rv -eq 0 ]; then 
        if [ "$HASH_1_MB" = "$hash2" ] && [ "$(basename "$(echo "$out" | jq -r .target)")" == "datafile-1-MB" ]; then
            function="delete_bucket"
            out=$(delete_bucket "$bucket_name") 
            rv=$? 
            # remove download file
            rm -rf /tmp/datafile-1-MB
        else
            rv=1
        fi
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Tests for presigned URL download success case, presigned URL is correct and 
# accessible - we calculate md5sum of the object and validate 
# it against a local files md5sum.
function test_presigned_download_object() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$?

    # if make bucket succeeds, upload a file via mc cp
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} cp $MINT_DATA_DIR/datafile-1-MB ${SERVER_ALIAS}/${bucket_name}/datafile-1-MB"
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    # if upload succeeds, generate download url via share download
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} share --json download ${SERVER_ALIAS}/${bucket_name}/datafile-1-MB"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?
    fi

    # if url generation succeeds, download via curl
    if [ $rv -eq 0 ]; then
        # get presigned URL download
        url=$(echo "$out" | jq -r .share)
        # download the file
        curl -sS -X GET "$url" > /tmp/datafile-1-MB
        # calculate the md5 hash of downloaded file
        hash2=$(md5sum /tmp/datafile-1-MB | awk '{print $1}')    
    fi

    if [ "$HASH_1_MB" = "$hash2" ]; then
        function="delete_bucket"
        out=$(delete_bucket "$bucket_name") 
        rv=$? 
        # remove download file
        rm -rf /tmp/datafile-1-MB
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Tests for list object success, by mirroring an fs store to minio, 
# and then comparing the ls results.
function test_mirror_list_objects() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$? 

    # if make bucket succeeds, start mirroring $MINT_DATA_DIR
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} mirror -q $MINT_DATA_DIR ${SERVER_ALIAS}/${bucket_name}"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    if [ $rv -eq 0 ]; then 
        if  diff -bB <(find "$MINT_DATA_DIR" -type f -printf "%f\n" | sort) <(./mc ls --json "${SERVER_ALIAS}/${bucket_name}" | jq -r .key | sort) > /dev/null; then
            function="delete_bucket"
            out=$(delete_bucket "$bucket_name") 
            rv=$?
        else
            rv=1
        fi
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Tests for the cat object success by comparing the STDOUT of "cat run.sh",
# to "mc cat run.sh" by uploading run.sh, and comparing the two outputs.
function test_cat_objects() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$? 

    # if make bucket succeeds, upload a file
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} cp ./run.sh ${SERVER_ALIAS}/${bucket_name}"
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi

    # if upload succeeds, cat the file using mc cat
    if [ $rv -eq 0 ]; then
        function="./mc cat target/${bucket_name}/run.sh"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?
    fi

    # compare output
    if [ $rv -eq 0 ]; then 
        if diff <(cat ./run.sh) <($function); then
            function="delete_bucket"
            out=$(delete_bucket "$bucket_name") 
            rv=$?
        else 
            rv=1
        fi
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Tests for mc watch, by running mc watch in the background, creating several 
# objects, checking to see if the words "ObjectCreated" and "ObjectRemoved", 
# were printed to STDOUT.
function watchObjects() {
    local bucket_name
    bucket_name=$(create_random_string)
    ./mc mb "target/$bucket_name" >/dev/null

    # send all output to a variable myvar, and run the operation
    # in the background
    ./mc watch --json "target/$bucket_name" > myvar&
    processID=$!

    # these operations should cause mc watch to print "ObjectCreated and ObjectRemoved"
    ./mc cp "$MINT_DATA_DIR/datafile-1-b play/$bucket_name">/dev/null
    ./mc rm --force --recursive "target/$bucket_name">/dev/null

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

# Create a bucket and check if it exists on server
function test_make_bucket_error() {
    
    # Make bucket - invalid bucket name
    bucket_name="Mc-mint-test-bucket-$RANDOM"
    function="${MC_CMD} mb ${SERVER_ALIAS}/${bucket_name}"

    # log start time
    start_time=$(get_time)

    # execute the test
    out=$($function 2>&1)
    rv=$?

    if [ $rv -eq 1 ]; then
        log_success "$(get_duration "$start_time")" "${function}"
        rv=0
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
}

# Try to upload an object with invalid object name, and verify if the 
# upload fails
function test_put_object_error() {

    # log start time
    start_time=$(get_time)

    # Make bucket
    function="make_bucket"
    bucket_name=$(make_bucket)  
    rv=$? 

    # if make bucket succeeds, try to upload an object with invalid name
    if [ $rv -eq 0 ]; then
        function="${MC_CMD} cp $MINT_DATA_DIR/datafile-1-MB ${SERVER_ALIAS}/${bucket_name}//2123123\123"
        # save the ref to function being tested, so it can be logged
        test_function=${function}
        out=$($function 2>&1)
        rv=$?
    else 
        # if make bucket fails, $bucket_name has the error output
        out="${bucket_name}"
    fi 
    
    # mc returns status 1 if case of invalid object name
    if [ $rv -eq 1 ]; then
        function="delete_bucket"
        out=$(delete_bucket "$bucket_name") 
        rv=$?
    else
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    if [ $rv -eq 0 ]; then
        log_success "$(get_duration "$start_time")" "${test_function}"
    else
        ${MC_CMD} rm --force --recursive ${SERVER_ALIAS}/"${bucket_name}" > /dev/null 2>&1
        log_failure "$(get_duration "$start_time")" "${function}" "${out}"
    fi

    return $rv
} 

# main handler for all the tests.
function main() {

    test_make_bucket && \
    test_put_object && \
    test_put_object_multipart && \
    test_presigned_upload_object && \
    test_presigned_download_object && \
    test_mirror_list_objects && \
    test_cat_objects && \
    test_make_bucket_error && \
    test_put_object_error

    return $?
}

main
