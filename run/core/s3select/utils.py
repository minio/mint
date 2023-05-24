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

import inspect
import json
import time
import traceback
import uuid


class LogOutput(object):
    """
    LogOutput is the class for log output. It is required standard for all
    SDK tests controlled by mint.
    Here are its attributes:
            'name': name of the SDK under test, e.g. 's3select'
            'function': name of the method/api under test with its signature
                        The following python code can be used to
                        pull args information of a <method> and to
                        put together with the method name:
                        <method>.__name__+'('+', '.join(args_list)+')'
                        e.g. 'remove_object(bucket_name, object_name)'
            'args': method/api arguments with their values, in
                    dictionary form: {'arg1': val1, 'arg2': val2, ...}
            'duration': duration of the whole test in milliseconds,
                        defaults to 0
            'alert': any extra information user is needed to be alerted about,
                     like whether this is a Blocker/Gateway/Server related
                     issue, etc., defaults to None
            'message': descriptive error message, defaults to None
            'error': stack-trace/exception message(only in case of failure),
                     actual low level exception/error thrown by the program,
                     defaults to None
            'status': exit status, possible values are 'PASS', 'FAIL', 'NA',
                      defaults to 'PASS'
    """

    PASS = 'PASS'
    FAIL = 'FAIL'
    NA = 'NA'

    def __init__(self, meth, test_name):
        self.__args_list = inspect.getargspec(meth).args[1:]
        self.__name = 's3select:'+test_name
        self.__function = meth.__name__+'('+', '.join(self.__args_list)+')'
        self.__args = {}
        self.__duration = 0
        self.__alert = ''
        self.__message = None
        self.__error = None
        self.__status = self.PASS
        self.__start_time = time.time()

    @property
    def name(self): return self.__name

    @property
    def function(self): return self.__function

    @property
    def args(self): return self.__args

    @name.setter
    def name(self, val): self.__name = val

    @function.setter
    def function(self, val): self.__function = val

    @args.setter
    def args(self, val): self.__args = val

    def json_report(self, err_msg='', alert='', status=''):
        self.__args = {k: v for k, v in self.__args.items() if v and v != ''}
        entry = {'name': self.__name,
                 'function': self.__function,
                 'args': self.__args,
                 'duration': int(round((time.time() - self.__start_time)*1000)),
                 'alert': str(alert),
                 'message': str(err_msg),
                 'error': traceback.format_exc() if err_msg and err_msg != '' else '',
                 'status': status if status and status != '' else
                 self.FAIL if err_msg and err_msg != '' else self.PASS
                 }
        return json.dumps({k: v for k, v in entry.items() if v and v != ''})


def generate_bucket_name():
    return "s3select-test-" + str(uuid.uuid4())


def generate_object_name():
    return str(uuid.uuid4())
