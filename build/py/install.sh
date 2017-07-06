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

# Install python deps
install() {
    apt-get install -yq python3
    apt-get install -yq python3-pip
    # python3 installation doesn't create /usr/bin/python symlink,
    # create this pro-actively.
    update-alternatives --install /usr/bin/python python /usr/bin/python3.5 1
}

# Remove python dependencies
cleanup() {
    apt-get purge -yq python3-pip && \
    apt-get autoremove -yq
}

installPkgs() {
    ## Execute all scripts present in py/* other than `install.sh`
    for i in $(echo /mint/build/py/*.sh | tr ' ' '\n' | grep -v install.sh); do
        $i
    done
}

main() {
    # Start with installing python3.
    install

    # Install all the dependent packages which are used
    # for running tests
    installPkgs

    # Cleanup any unnecessary packages not required at runtime.
    cleanup
}

main
