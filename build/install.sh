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

# Install general dependencies
installDeps() {
    apt-get update -yq
    apt-get install -yq curl openssl
}

cleanupDeps() {
    apt-get purge -yq curl git
    apt-get autoremove -yq
}

main() {
    installDeps

    for i in $(echo /mint/build/*/install.sh | tr ' ' '\n'); do
        $i
    done

    # Remove all the used deps
    cleanupDeps
}

main
