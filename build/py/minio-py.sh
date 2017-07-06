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
    MINIO_PY_SDK_PATH="/mint/run/core/minio-py"
    MINIO_PY_SDK_VERSION="2.2.4"
}

installDeps() {
    pip3 install --user -r ${MINIO_PY_SDK_PATH}/requirements.txt
    pip3 install minio==$MINIO_PY_SDK_VERSION
}

installFunctionalTest() {
    curl https://raw.githubusercontent.com/minio/minio-py/${MINIO_PY_SDK_VERSION}/tests/functional/tests.py > /usr/bin/minio-py-functional
    # This is needed until we make a new minio-py release.
    sed -i 's/DATA_DIR/MINT_DATA_DIR/g' /usr/bin/minio-py-functional
    chmod +x /usr/bin/minio-py-functional
    sync
}

main() {
    installDeps
    installFunctionalTest
}

_init "$@" && main
