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
    MINIO_JAVA_SDK_PATH="/mint/run/core/minio-java"
    MINIO_JAVA_SDK_VERSION="3.0.6"
}

# install java dependencies.
install() {
    curl https://raw.githubusercontent.com/minio/minio-java/"${MINIO_JAVA_SDK_VERSION}"/functional/ContentInputStream.java > "${MINIO_JAVA_SDK_PATH}"/ContentInputStream.java
    curl https://raw.githubusercontent.com/minio/minio-java/"${MINIO_JAVA_SDK_VERSION}"/functional/PutObjectRunnable.java > "${MINIO_JAVA_SDK_PATH}"/PutObjectRunnable.java
    curl https://raw.githubusercontent.com/minio/minio-java/"${MINIO_JAVA_SDK_VERSION}"/functional/FunctionalTest.java > "${MINIO_JAVA_SDK_PATH}"/FunctionalTest.java
    curl https://raw.githubusercontent.com/minio/minio-java/"${MINIO_JAVA_SDK_VERSION}"/functional/MintLogger.java > "${MINIO_JAVA_SDK_PATH}"/MintLogger.java
    curl -s -o "$MINIO_JAVA_SDK_PATH/minio-${MINIO_JAVA_SDK_VERSION}-all.jar" "http://repo1.maven.org/maven2/io/minio/minio/${MINIO_JAVA_SDK_VERSION}/minio-${MINIO_JAVA_SDK_VERSION}-all.jar"
 }

# Compile test files
build() {
    javac -cp "$MINIO_JAVA_SDK_PATH/minio-${MINIO_JAVA_SDK_VERSION}-all.jar" "${MINIO_JAVA_SDK_PATH}/*.java" 
}

main() {
    install
    build
}

_init "$@" && main
