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
    minio_js_sdk_path=$1
    minio_js_sdk_version=$2
}

# Install JS dependencies
installMinioJSDeps() {
    curl -sL https://deb.nodesource.com/setup_6.x | bash - && \
    apt-get install -y nodejs
}

# Compile test files
buildMinioJSTests() {
    npm --prefix $minio_js_sdk_path install --save minio@$minio_js_sdk_version && \
    npm --prefix $minio_js_sdk_path install
}

jsMain() {
    installMinioJSDeps && \
    buildMinioJSTests
}

_init "$@" && jsMain