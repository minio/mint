#!/bin/bash -e
#
#  Mint (C) 2017-2020 Minio, Inc.
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

MINIO_PY_VERSION="7.2.19"
test_run_dir="$MINT_RUN_CORE_DIR/minio-py"

pip3 install --break-system-packages --user faker
pip3 install --break-system-packages --no-cache-dir minio=="${MINIO_PY_VERSION}"
$WGET --output-document="$test_run_dir/tests.py" "https://raw.githubusercontent.com/minio/minio-py/${MINIO_PY_VERSION}/tests/functional/tests.py"
