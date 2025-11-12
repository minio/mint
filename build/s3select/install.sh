#!/bin/bash -e
#
#  Mint (C) 2020 Minio, Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

# Using --break-system-packages for Ubuntu 24.04+ (PEP 668) - safe in containers
# Install minio 7.2.18 which includes fix for ExcludedPrefixes XML element name bug (broken in 7.2.13, fixed in 7.2.14+)
python -m pip install --break-system-packages --no-cache-dir "minio==7.2.18"

SELECT_PY=$(python -c "import minio.select; import os; print(os.path.dirname(minio.select.__file__))")/select.py
sed -i 's/^    allow_quoted_record_delimiter = None$/    allow_quoted_record_delimiter: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    comments = None$/    comments: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    field_delimiter = None$/    field_delimiter: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    file_header_info = None$/    file_header_info: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    quote_character = None$/    quote_character: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    quote_escape_character = None$/    quote_escape_character: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    record_delimiter = None$/    record_delimiter: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    json_type = None$/    json_type: Optional[str] = None/' "$SELECT_PY"
sed -i 's/^    quote_fields = None$/    quote_fields: Optional[str] = None/' "$SELECT_PY"
