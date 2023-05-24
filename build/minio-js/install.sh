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

MINIO_JS_VERSION=$(curl --retry 10 -Ls -o /dev/null -w "%{url_effective}" https://github.com/minio/minio-js/releases/latest | sed "s/https:\/\/github.com\/minio\/minio-js\/releases\/tag\///")
if [ -z "$MINIO_JS_VERSION" ]; then
	echo "unable to get minio-js version from github"
	exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/minio-js"
mkdir "${test_run_dir}/test"
$WGET --output-document="${test_run_dir}/test/functional-tests.js" "https://raw.githubusercontent.com/minio/minio-js/${MINIO_JS_VERSION}/src/test/functional/functional-tests.js"
npm --prefix "$test_run_dir" install --save "minio@$MINIO_JS_VERSION"
npm --prefix "$test_run_dir" install
