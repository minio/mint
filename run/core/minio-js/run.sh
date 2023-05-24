#!/bin/bash
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

# handle command line arguments
if [ $# -ne 2 ]; then
	echo "usage: run.sh <OUTPUT-LOG-FILE> <ERROR-LOG-FILE>"
	exit 1
fi

output_log_file="$1"
error_log_file="$2"

# run tests
./node_modules/mocha/bin/mocha -R minioreporter -b --exit 1>>"$output_log_file" 2>"$error_log_file"
