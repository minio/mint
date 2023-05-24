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

MINT_DATA_DIR="$MINT_ROOT_DIR/data"

declare -A data_file_map
data_file_map["datafile-0-b"]="0"
data_file_map["datafile-1-b"]="1"
data_file_map["datafile-1-kB"]="1K"
data_file_map["datafile-10-kB"]="10K"
data_file_map["datafile-33-kB"]="33K"
data_file_map["datafile-100-kB"]="100K"
data_file_map["datafile-1.03-MB"]="1056K"
data_file_map["datafile-1-MB"]="1M"
data_file_map["datafile-5-MB"]="5M"
data_file_map["datafile-5243880-b"]="5243880"
data_file_map["datafile-6-MB"]="6M"
data_file_map["datafile-10-MB"]="10M"
data_file_map["datafile-11-MB"]="11M"
data_file_map["datafile-65-MB"]="65M"
data_file_map["datafile-129-MB"]="129M"

mkdir -p "$MINT_DATA_DIR"
for filename in "${!data_file_map[@]}"; do
	echo "creating $MINT_DATA_DIR/$filename"
	if ! shred -n 1 -s "${data_file_map[$filename]}" - 1>"$MINT_DATA_DIR/$filename" 2>/dev/null; then
		echo "unable to create data file $MINT_DATA_DIR/$filename"
		exit 1
	fi
done
