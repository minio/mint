#!/bin/sh
#
#  Minio Cloud Storage, (C) 2017 Minio, Inc.
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

# Add build instructions for SDK tests here.
# The Run instructions should be added to run.sh
# minio/pkg/admin is a dependency for main.go
# admin api from the package is used to check whether Minio server with given
# credentials is reachable.
go get -u github.com/minio/minio/pkg/madmin
# first build the main program.
go build main.go



