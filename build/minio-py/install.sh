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

# Using master branch temporarily until 7.2.19 is released (contains type annotation fixes and new API)
# TO BE FIXED
MINIO_PY_VERSION="371a384ff31cc72db3d44bf61725c1091b315f99"
test_run_dir="$MINT_RUN_CORE_DIR/minio-py"
# Using --break-system-packages for Ubuntu 24.04+ (PEP 668) - safe in containers
pip3 install --break-system-packages --user faker
pip3 install --break-system-packages --no-cache-dir git+https://github.com/rraulinio/minio-py.git@$MINIO_PY_VERSION

$WGET --output-document="$test_run_dir/tests.py" "https://raw.githubusercontent.com/rraulinio/minio-py/${MINIO_PY_VERSION}/tests/functional/tests.py"
