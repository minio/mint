#!/bin/bash -e
#
#  Minio Cloud Storage, (C) 2024 Minio, Inc.
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

export MINT_RUN_CORE_DIR="$MINT_ROOT_DIR/run/core"
export MINT_RUN_BUILD_DIR="$MINT_ROOT_DIR/build"
export APT="apt --quiet --yes"
export WGET="wget --quiet --no-check-certificate"
export WGET="wget --quiet --no-check-certificate"

## Software versions
export GO_VERSION="1.21.9"
export GRADLE_VERSION="8.5"
export GRADLE_INSTALL_PATH="/opt/gradle"
export GO_INSTALL_PATH="/usr/local"

export PATH=${GO_INSTALL_PATH}/bin:$PATH
export PATH=${GRADLE_INSTALL_PATH}/gradle-${GRADLE_VERSION}/bin:$PATH
