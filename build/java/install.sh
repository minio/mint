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

# install java dependencies.
install() {
    apt-get update && apt-get install -yq default-jre default-jdk
}

installPkgs() {
    ## Execute all scripts present in java/* other than `install.sh`
    for i in $(echo /mint/build/java/*.sh | tr ' ' '\n' | grep -v install.sh); do
        $i
    done
}

# remove java dependencies.
cleanup() {
    apt-get purge -yq default-jdk
    apt-get autoremove -yq
}

main() {
    install

    installPkgs

    cleanup
}

main

