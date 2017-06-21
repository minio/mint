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
    minio_java_sdk_path=$1
    minio_java_version=$2

    MINIO_JAR="/usr/local/minio.jar"
    OKHTTP_JAR="/usr/local/okhttp.jar"
}

# Install Java dependencies and Download Minio.jar / okhttp and 
installMinioJavaDeps() {
    apt-get update && apt-get install -yq default-jre default-jdk

    curl -s -o $MINIO_JAR http://repo1.maven.org/maven2/io/minio/minio/${minio_java_version}/minio-${minio_java_version}-all.jar 
	curl -s -o $OKHTTP_JAR http://central.maven.org/maven2/com/squareup/okhttp3/okhttp/3.7.0/okhttp-3.7.0.jar
}

# Compile test files
buildMinioJavaTests() {
	javac -cp $MINIO_JAR ${minio_java_sdk_path}/FunctionalTest.java ${minio_java_sdk_path}/PutObjectRunnable.java
}

# Remove Java dependencies
cleanMinioJavaDeps() {
    apt-get purge -yq default-jdk && \
    apt-get autoremove -yq
}

javaMain() {
    installMinioJavaDeps && \
    buildMinioJavaTests && \
    cleanMinioJavaDeps
}

_init "$@" && javaMain