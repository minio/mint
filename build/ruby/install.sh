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

# Install ruby deps
install() {
     apt-get update && \
     apt-get install -y ruby ruby-dev ruby-bundler && \
     rm -rf /var/lib/apt/lists/*
}
 

installPkgs() {
    ## Execute all scripts present in ruby/* other than `install.sh`
    for i in $(echo /mint/build/ruby/*.sh | tr ' ' '\n' | grep -v install.sh); do
        $i
    done
}

main() {
    # Start with installing ruby.
    install

    # Install all the dependent packages which are used
    # for running tests
    installPkgs

    # Cleanup any unnecessary packages not required at runtime.
    
}

main
