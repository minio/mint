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

MC_VERSION=$(curl --retry 10 -Ls -o /dev/null -w "%{url_effective}" https://github.com/minio/mc/releases/latest | sed "s/https:\/\/github.com\/minio\/mc\/releases\/tag\///")
if [ -z "$MC_VERSION" ]; then
	echo "unable to get mc version from github"
	exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/mc"
$WGET --output-document="${test_run_dir}/mc" "https://dl.minio.io/client/mc/release/linux-amd64/mc.${MC_VERSION}"
chmod a+x "${test_run_dir}/mc"

git clone --quiet https://github.com/minio/mc.git "$test_run_dir/mc.git"
(
	cd "$test_run_dir/mc.git"
	git checkout --quiet "tags/${MC_VERSION}"
)
cp -a "${test_run_dir}/mc.git/functional-tests.sh" "$test_run_dir/"
rm -fr "$test_run_dir/mc.git"
