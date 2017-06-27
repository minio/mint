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

set -e

# Install PHP deps
install() {
    apt-get install -yq php php7.0-curl php-xml
}

# Remove PHP dependencies
cleanup() {
    apt-get autoremove -yq
}

installPkgs() {
    ## Execute all scripts present in py/* other than `install.sh`
    for i in $(echo /mint/build/php/*.sh | tr ' ' '\n' | grep -v install.sh); do
        $i
    done
}

main() {
    # Start with installing PHP.
    install

    # Install all the dependent packages which are used
    # for running tests
    installPkgs

    # Cleanup any unnecessary packages not required at runtime.
    cleanup
}

main
