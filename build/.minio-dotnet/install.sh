#!/bin/bash
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

set -e

MINIO_DOTNET_SDK_PATH="$MINT_RUN_CORE_DIR/minio-dotnet"

MINIO_DOTNET_SDK_VERSION=$(curl --retry 10 -Ls -o /dev/null -w "%{url_effective}" https://github.com/minio/minio-dotnet/releases/latest | sed "s/https:\/\/github.com\/minio\/minio-dotnet\/releases\/tag\///")
if [ -z "$MINIO_DOTNET_SDK_VERSION" ]; then
	echo "unable to get minio-dotnet version from github"
	exit 1
fi

out_dir="$MINIO_DOTNET_SDK_PATH/out"
if [ -z "$out_dir" ]; then
	mkdir "$out_dir"
fi

temp_dir="$MINIO_DOTNET_SDK_PATH/temp"
git clone --quiet https://github.com/minio/minio-dotnet.git "${temp_dir}"
pushd "${temp_dir}" >/dev/null
git checkout --quiet "tags/${MINIO_DOTNET_SDK_VERSION}"

dotnet publish Minio.Functional.Tests --configuration Mint --framework net6.0 --output ../out

popd >/dev/null
rm -fr "${temp_dir}"
