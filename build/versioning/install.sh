#!/bin/bash -e
#
#  Mint (C) 2017-2021 Minio, Inc.
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

test_run_dir="$MINT_RUN_CORE_DIR/versioning"
test_build_dir="$MINT_RUN_BUILD_DIR/versioning"

(cd "$test_build_dir" && CGO_ENABLED=0 go build --ldflags "-s -w" -o "$test_run_dir/tests")
