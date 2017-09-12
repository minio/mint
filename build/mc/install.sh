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

MC_VERSION="RELEASE.2017-06-15T03-38-43Z"

test_run_dir="$MINT_RUN_CORE_DIR/mc"
$WGET --output-document="${test_run_dir}/mc" "https://dl.minio.io/client/mc/release/linux-amd64/mc.${MC_VERSION}"
chmod a+x "${test_run_dir}/mc"
