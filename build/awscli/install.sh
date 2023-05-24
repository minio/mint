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

die() {
	echo "$*" 1>&2
	exit 1
}

# shellcheck disable=SC2086
ROOTDIR="$(dirname "$(realpath $0)")"
TMPDIR="$(mktemp -d)"

cd "$TMPDIR"

# Download botocore and apply @y4m4's expect 100 continue fix
(git clone --depth 1 -b 1.27.1 https://github.com/boto/botocore &&
	cd botocore &&
	patch -p1 <"$ROOTDIR/expect-100.patch" &&
	python3 -m pip install .) ||
	die "Unable to install botocore.."

# Download and install aws cli
(git clone --depth 1 -b 1.25.1 https://github.com/aws/aws-cli &&
	cd aws-cli &&
	python3 -m pip install .) ||
	die "Unable to install aws-cli.."

# Clean-up
rm -r "$TMPDIR"
