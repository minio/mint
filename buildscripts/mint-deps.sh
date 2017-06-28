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

# Minimum required versions for various languages
_init() {
    GO_VERSION="1.7.5"
    INIT_PATH="/mint/run/core/init"
    MINIO_GO_SDK_PATH="/mint/run/core/minio-go"
    
    MINIO_JAVA_SDK_PATH="/mint/run/core/minio-java"
    MINIO_JAVA_SDK_VERSION="3.0.5"

    MC_VERSION="RELEASE.2017-06-15T03-38-43Z"
    MC_TEST_PATH="/mint/run/core/mc"

    MINIO_JS_SDK_PATH="/mint/run/core/minio-js"
    MINIO_JS_SDK_VERSION="3.1.3"

    export SUDO_FORCE_REMOVE=yes
}

# Install general dependencies
installGeneralDeps() {
    apt-get update -yq
    apt-get install -yq \
    curl \
    openssl
}

removeGeneralDeps() {
    apt-get purge -yq curl git
    apt-get autoremove -yq 
}

buildMain() {
    installGeneralDeps
    
    /mint/buildscripts/go-deps.sh $INIT_PATH $MINIO_GO_SDK_PATH $GO_VERSION

    /mint/buildscripts/java-deps.sh $MINIO_JAVA_SDK_PATH $MINIO_JAVA_SDK_VERSION

    /mint/buildscripts/cli-tools-deps.sh $MC_TEST_PATH $MC_VERSION

    /mint/buildscripts/js-deps.sh $MINIO_JS_SDK_PATH $MINIO_JS_SDK_VERSION

    /mint/buildscripts/py/install.sh

    # Remove all the used deps
    removeGeneralDeps
}

_init && buildMain
