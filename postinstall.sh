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

# remove all packages listed in remove-packages.list
xargs --arg-file="${MINT_ROOT_DIR}/remove-packages.list" apt --quiet --yes purge
${APT} autoremove

# remove unwanted files
rm -fr "$GOROOT" "$GOPATH/src" /var/lib/apt/lists/*

# flush to disk
sync
