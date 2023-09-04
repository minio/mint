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

MINIO_GO_VERSION=$(curl --retry 10 -Ls -o /dev/null -w "%{url_effective}" https://github.com/minio/minio-go/releases/latest | sed "s/https:\/\/github.com\/minio\/minio-go\/releases\/tag\///")
if [ -z "$MINIO_GO_VERSION" ]; then
	echo "unable to get minio-go version from github"
	exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/minio-go"
curl -sL -o "${test_run_dir}/main.go" "https://raw.githubusercontent.com/minio/minio-go/${MINIO_GO_VERSION}/functional_tests.go"
(cd "$test_run_dir" && CGO_ENABLED=0 go build --ldflags "-s -w" -o minio-go main.go)
