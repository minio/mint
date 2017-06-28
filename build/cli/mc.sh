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

set -e

_init() {
    MC_VERSION="RELEASE.2017-06-15T03-38-43Z"
    MC_TEST_PATH="/mint/run/core/mc"
}

install() {
    # Download MC specific version
    curl -s -o "${MC_TEST_PATH}/mc" "https://dl.minio.io/client/mc/release/linux-amd64/mc.${MC_VERSION}"

    res=$?
    if test "$res" != "0"; then
        echo "curl command to download mc failed with: $res"
        exit 1
    else
        chmod +x "${MC_TEST_PATH}/mc"
        sync
        echo "Downloaded mc $("${MC_TEST_PATH}"/mc version | grep Version)"
    fi
}

main() {
    install
}

_init && main
