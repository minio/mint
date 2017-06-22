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
    minio_py_sdk_path=$1
    minio_py_sdk_version=$2
}

# Install PY dependencies
installMinioPyDeps() {
    apt-get install -yq python3-pip && \
    pip3 install --user -r ${minio_py_sdk_path}/requirements.txt && \
    pip3 install minio==$minio_py_sdk_version
}

# Remove Python dependencies
cleanMinioPyDeps() {
    apt-get purge -yq python3-pip && \
    apt-get autoremove -yq
}

pyMain() {
    installMinioPyDeps && \
    cleanMinioPyDeps
}

_init "$@" && pyMain