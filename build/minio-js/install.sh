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

set -ex

test_run_dir="$MINT_RUN_CORE_DIR/minio-js"

rm -rf "${test_run_dir}/minio-js"

git clone https://github.com/minio/minio-js.git "${test_run_dir}/minio-js"

cd "${test_run_dir}/minio-js"

LATEST=$(git tag | tail -1)

git checkout "${LATEST}" --force

npm i
