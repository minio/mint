#!/bin/bash -e
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

MINIO_GO_VERSION=$(curl -s https://api.github.com/repos/minio/minio-go/releases/latest | jq -r .tag_name)
if [ -z "$MINIO_GO_VERSION" ]; then
    echo "unable to get minio-go version from github"
    exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/minio-go"
go get -u github.com/sirupsen/logrus/...
go get -u github.com/dustin/go-humanize/...
go get -u github.com/minio/minio-go/...
(cd "$GOPATH/src/github.com/minio/minio-go" && git checkout --quiet "tags/$MINIO_GO_VERSION")
CGO_ENABLED=0 go build -o "$test_run_dir/minio-go" "$GOPATH/src/github.com/minio/minio-go/functional_tests.go"
