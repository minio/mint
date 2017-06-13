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

cleanUp() {
    # remove executable 
    rm initCheck 2> /dev/null
}

build() {
    go build -o initCheck ./initCheck.go
}

run() {
	chmod +x initCheck
    ./initCheck
}

main() {
    # Build test file binary
    build >>$1  2>&1 || { echo 'initCheck build failed' ; exit 1; }
    
    # run the tests
    run >>$1  2>&1 && cleanUp || { echo 'Server not reachable. Exiting...'; exit 1;}
}

# invoke the script
main "$@"
