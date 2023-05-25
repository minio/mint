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

# Checkout at /mint/test-run/minio-js/
# During run of the test copy it to the the /min/run/core/minio-js/minio-js

install_path="./test-run/minio-js/"
rm -rf $install_path

git clone https://github.com/minio/minio-js.git $install_path

cd $install_path || exit 0

# Get new tags from remote
git fetch --tags
# Get latest tag name
# shellcheck disable=SC2046
LATEST=$(git describe --tags $(git rev-list --tags --max-count=1))

echo "Using minio-js RELEASE $LATEST"

git checkout "${LATEST}" --force &>/dev/null

npm install --quiet &>/dev/null
