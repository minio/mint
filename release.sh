#!/bin/bash -e
#
#  MinIO Cloud Storage, (C) 2017-2023 MinIO, Inc.
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

export MINT_ROOT_DIR=${MINT_ROOT_DIR:-/mint}
export MINT_RUN_CORE_DIR="$MINT_ROOT_DIR/run/core"
export MINT_RUN_BUILD_DIR="$MINT_ROOT_DIR/build"
export WGET="wget --quiet --no-check-certificate"

"${MINT_ROOT_DIR}"/create-data-files.sh
"${MINT_ROOT_DIR}"/preinstall.sh

# install mint app packages
for pkg in "$MINT_ROOT_DIR/build"/*/install.sh; do
	echo "Running $pkg"
	$pkg
done

"${MINT_ROOT_DIR}"/postinstall.sh
