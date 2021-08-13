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

KITCHENSINK_VERSION=$(curl --retry 10 -Ls -o /dev/null -w "%{url_effective}" https://github.com/minio/kitchensink/releases/latest | sed "s/https:\/\/github.com\/minio\/kitchensink\/releases\/tag\///")
if [ -z "$KITCHENSINK_VERSION" ]; then
    echo "unable to get kitchensink version from github"
    exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/kitchensink"
$WGET --output-document="${test_run_dir}/kitchensink" "https://github.com/minio/kitchensink/releases/download/${KITCHENSINK_VERSION}/kitchensink-linux-amd64"
chmod a+x "${test_run_dir}/kitchensink"
