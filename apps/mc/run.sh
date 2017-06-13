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

cleanUp(){
    # remove mc 
    rm mc
}

downloadMC() {
    # Download latest MC release
    curl -s -o mc https://dl.minio.io/client/mc/release/linux-amd64/mc

build() {
        
    if [ $ENABLE_HTTPS -eq "1" ]; then
        target_address="https://"$SERVER_ENDPOINT
    else
        target_address="http://"$SERVER_ENDPOINT
    fi

    # Download latest MC release
    # curl -s -o mc https://dl.minio.io/client/mc/release/linux-amd64/mc
    curl -s -o mc https://dl.minio.io/client/mc/release/darwin-amd64/mc

    res=$?
    if test "$res" != "0"; then
        echo "curl command to download mc failed with: $res"
        exit 1
    else 
        chmod +x ./mc
        echo "Downloaded mc $(./mc version | grep Version)"
        echo "Adding mc host alias target $target_address"
        ./mc config host add target $target_address $ACCESS_KEY $SECRET_KEY
    fi
}

# Execute test.sh 
run() {
    chmod +x ./test.sh
    ./test.sh
}

main() {
    # Build test file binary
    build -s  2>&1  >| $1

    # run the tests
    run -s  2>&1  >| $1

    # remove the executable
    cleanUp

    grep -q 'Error:|FAIL' $1 > $2

    return 0
}

# invoke the script
main "$@"

