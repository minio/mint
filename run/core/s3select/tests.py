#!/usr/bin/env python
# -*- coding: utf-8 -*-
# MinIO Python Library for Amazon S3 Compatible Cloud Storage,
# (C) 2015-2023 MinIO, Inc.
#
# This file is part of MinIO Object Storage stack
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

import os
import sys
from csv import (test_csv_input_custom_quote_char,
                 test_csv_output_custom_quote_char)

from minio import Minio

from sql_ops import (test_sql_datatypes, test_sql_functions_agg_cond_conv,
                     test_sql_functions_date, test_sql_functions_string,
                     test_sql_operators, test_sql_operators_precedence,
                     test_sql_select, test_sql_select_csv_no_header,
                     test_sql_select_json)
from utils import LogOutput


def main():
    """
    Functional testing for S3 select.
    """

    try:
        access_key = os.getenv('ACCESS_KEY', 'Q3AM3UQ867SPQQA43P2F')
        secret_key = os.getenv('SECRET_KEY',
                               'zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG')
        server_endpoint = os.getenv('SERVER_ENDPOINT', 'play.min.io')
        secure = os.getenv('ENABLE_HTTPS', '1') == '1'
        if server_endpoint == 'play.min.io':
            access_key = 'Q3AM3UQ867SPQQA43P2F'
            secret_key = 'zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG'
            secure = True

        client = Minio(server_endpoint, access_key, secret_key, secure=secure)

        log_output = LogOutput(client.select_object_content,
                               'test_csv_input_quote_char')
        test_csv_input_custom_quote_char(client, log_output)

        log_output = LogOutput(client.select_object_content,
                               'test_csv_output_quote_char')
        test_csv_output_custom_quote_char(client, log_output)

        log_output = LogOutput(
            client.select_object_content, 'test_sql_operators')
        test_sql_operators(client, log_output)

        log_output = LogOutput(client.select_object_content,
                               'test_sql_operators_precedence')
        test_sql_operators_precedence(client, log_output)

        log_output = LogOutput(client.select_object_content,
                               'test_sql_functions_agg_cond_conv')
        test_sql_functions_agg_cond_conv(client, log_output)

        log_output = LogOutput(
            client.select_object_content, 'test_sql_functions_date')
        test_sql_functions_date(client, log_output)

        log_output = LogOutput(client.select_object_content,
                               'test_sql_functions_string')
        test_sql_functions_string(client, log_output)

        log_output = LogOutput(
            client.select_object_content, 'test_sql_datatypes')
        test_sql_datatypes(client, log_output)

        log_output = LogOutput(client.select_object_content, 'test_sql_select')
        test_sql_select(client, log_output)

        log_output = LogOutput(
            client.select_object_content, 'test_sql_select_json')
        test_sql_select_json(client, log_output)

        log_output = LogOutput(
            client.select_object_content, 'test_sql_select_csv')
        test_sql_select_csv_no_header(client, log_output)

    except Exception as err:
        print(log_output.json_report(err))
        sys.exit(1)


if __name__ == "__main__":
    # Execute only if run as a script
    main()
