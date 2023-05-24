#!/bin/bash -e
#
#  Mint (C) 2017-2023 MinIO, Inc.
#
#  This file is part of MinIO Object Storage stack
#
#  This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as published by
#  the Free Software Foundation, either version 3 of the License, or
#  (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU Affero General Public License for more details.
#
#  You should have received a copy of the GNU Affero General Public License
#  along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

export APT="apt --quiet --yes"
export WGET="wget --quiet --no-check-certificate"

# install nodejs source list
if ! $WGET --output-document=- https://deb.nodesource.com/setup_20.x | bash -; then
	echo "unable to set nodejs repository"
	exit 1
fi

$APT install apt-transport-https

if ! $WGET --output-document=packages-microsoft-prod.deb https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb | bash -; then
	echo "unable to download dotnet packages"
	exit 1
fi

dpkg -i packages-microsoft-prod.deb
rm -f packages-microsoft-prod.deb

$APT update
$APT install gnupg ca-certificates

# download and install golang
GO_VERSION="1.20.4"
GO_INSTALL_PATH="/usr/local"
download_url="https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz"
if ! $WGET --output-document=- "$download_url" | tar -C "${GO_INSTALL_PATH}" -zxf -; then
	echo "unable to install go$GO_VERSION"
	exit 1
fi

xargs --arg-file="${MINT_ROOT_DIR}/install-packages.list" apt --quiet --yes install

# set python 3.10 as default
update-alternatives --install /usr/bin/python python /usr/bin/python3.10 1

sync
