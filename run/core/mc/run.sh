#!/bin/bash
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

# Execute test.sh 
run() {
    ./test.sh
}

setupMCTarget() {
        
    [ "$ENABLE_HTTPS" -eq "1" ] && scheme="https" || scheme="http" 

    target_address=$scheme://$SERVER_ENDPOINT
    
    echo "Adding mc host alias target $target_address"

    ./mc config host add target "$target_address" "$ACCESS_KEY" "$SECRET_KEY"
}

main() {
    
    logfile=$1
    errfile=$2

    # run the tests
    rc=0
    
    # setup MC alias target to point to SERVER_ENDPOINT
    setupMCTarget >>"$logfile"  2>&1 || { echo 'mc setup failed' ; exit 1; }

    run 2>>"$errfile" 1>>"$logfile" || { echo 'mc run failed.'; rc=1; } 
    grep -e '<ERROR>' "$logfile" >> "$errfile"
    return $rc
}

# invoke the script
main "$@"