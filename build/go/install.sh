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
    GO_VERSION="1.7.5"
    GO_INSTALL_PATH="/usr/local"
    MINIO_GO_PATH="/mint/run/core/minio-go"
}

# Install Go dependencies
installGo() {
    curl https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar -C "${GO_INSTALL_PATH}" -xzf -
}

installGoPkgs() {
    go get -u github.com/minio/minio-go/...
    go get -u github.com/sirupsen/logrus/...
}

# Build Minio Go tests
buildGoTests() {
    CGO_ENABLED=0 go build -o "${MINIO_GO_PATH}/minio-go" "${MINIO_GO_PATH}/minio-go-tests.go"
}

# Remove Go dependencies
cleanup() {
    rm -rf "${GO_INSTALL_PATH}"

    # Use "${var:?}" to ensure this never expands to /* . [SC2115]
    rm -rf "${GOPATH:?}/*"
}

main() {
    installGo
    installGoPkgs

    buildGoTests
    cleanup
}

_init "$@" && main
