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
    init_test_path=$1
    minio_go_sdk_path=$2
    go_version=$3
}

# Install Go dependencies
installMinioGoDeps() {
    curl -O https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz && \
    tar -xf go${go_version}.linux-amd64.tar.gz && \
    rm go${go_version}.linux-amd64.tar.gz && \
    mv go /usr/local
}
 
# Build init tests
buildInitTests() {
    go get -u github.com/minio/minio/pkg/madmin && \
    go build -o ${init_test_path}/initCheck ${init_test_path}/initCheck.go
}

# Build Minio Go tests
buildMinioGoTests() {
    go get -u github.com/minio/minio-go && \
	go test -o ${minio_go_sdk_path}/minio.test -c ${minio_go_sdk_path}/api_functional_v4_test.go
}

# Remove Go dependencies
cleanMinioGoDeps() {
    rm -rf /usr/local/go
}

goMain() {
    installMinioGoDeps && \
    buildInitTests && \
    buildMinioGoTests && \
    cleanMinioGoDeps
}

_init "$@" && goMain