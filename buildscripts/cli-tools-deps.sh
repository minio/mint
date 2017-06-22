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

set -e

_init() {
    mc_test_path=$1
    mc_version=$2
}

downloadMC() {
    # Download MC specific version
    curl -s -o ${mc_test_path}/mc https://dl.minio.io/client/mc/release/linux-amd64/mc.${mc_version}
    
    res=$?
    if test "$res" != "0"; then
        echo "curl command to download mc failed with: $res"
        exit 1
    else 
        chmod +x ${mc_test_path}/mc
        sync
        echo "Downloaded mc $(${mc_test_path}/mc version | grep Version)"
    fi
}

cliToolsMain() {
    downloadMC
}

_init "$@" && cliToolsMain