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
    MINIO_JS_SDK_PATH="/mint/run/core/minio-js"
    MINIO_JS_SDK_VERSION="3.1.3"
}

# Compile test files
install() {
    #TODO - Change this to release based URL once we make a minio-js release
    mkdir "${MINIO_JS_SDK_PATH}"/test
    curl https://raw.githubusercontent.com/minio/minio-js/master/src/test/functional/functional-tests.js > "${MINIO_JS_SDK_PATH}"/test/functional-tests.js
    npm --prefix "$MINIO_JS_SDK_PATH" install --save "minio@$MINIO_JS_SDK_VERSION"
    npm --prefix "$MINIO_JS_SDK_PATH" install
}

main() {
    install
}

_init "$@" && main
