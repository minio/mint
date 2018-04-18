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

export APT="apt --quiet --yes"
export WGET="wget --quiet --no-check-certificate"

# install nodejs source list
if ! $WGET --output-document=- https://deb.nodesource.com/setup_6.x | bash -; then
    echo "unable to set nodejs repository"
    exit 1
fi

# dotnetcore install commands
sh -c 'echo "deb [arch=amd64] https://apt-mo.trafficmanager.net/repos/dotnet-release/ xenial main" > /etc/apt/sources.list.d/dotnetdev.list'
apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 417A0893

$APT update

# download and install golang
GO_VERSION="1.8.3"
GO_INSTALL_PATH="/usr/local"
download_url="https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz"
if ! $WGET --output-document=- "$download_url" | tar -C "${GO_INSTALL_PATH}" -zxf -; then
    echo "unable to install go$GO_VERSION"
    exit 1
fi

xargs --arg-file=install-packages.list apt --quiet --yes install

# set python 3.5 as default
update-alternatives --install /usr/bin/python python /usr/bin/python3.5 1

sync
