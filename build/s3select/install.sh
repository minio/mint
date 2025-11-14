#!/bin/bash -e
#
#  Mint (C) 2020 Minio, Inc.
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

# Using --break-system-packages for Ubuntu 24.04+ (PEP 668) - safe in containers
# Install minio-py from master with type annotation fixes (commit cbac53b) until 7.2.19 is released
# TO BE FIXED
MINIO_PY_VERSION="968031f57675c9de81fda3397f6761c7fda05dea"
python -m pip install --break-system-packages --no-cache-dir git+https://github.com/rraulinio/minio-py.git@$MINIO_PY_VERSION
