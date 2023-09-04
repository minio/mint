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

# Checkout at /mint/test-run/minio-js/
# During run of the test copy it to the the /min/run/core/minio-js/minio-js

install_path="./test-run/minio-js/"
rm -rf $install_path

git clone https://github.com/minio/minio-js.git $install_path

cd $install_path || exit 0

# Get new tags from remote
git fetch --tags
# Get latest tag name
# shellcheck disable=SC2046
LATEST=$(git describe --tags $(git rev-list --tags --max-count=1))

echo "Using minio-js RELEASE $LATEST"

git checkout "${LATEST}" --force &>/dev/null

npm install --quiet &>/dev/null
