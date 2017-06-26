#!/bin/bash
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
data_dir="/mint/data"
if [ ! -d $data_dir ]; then 
		mkdir $data_dir
fi
cd $data_dir
dd if=/dev/zero of=SmallFile bs=1024 count=10
dd if=/dev/zero of=FileOfSize1B bs=1 count=1
dd if=/dev/zero of=FileOfSize1MB bs=1024 count=1024
dd if=/dev/zero of=FileOfSizeGt1MB bs=1024 count=1056

dd if=/dev/zero of=FileOfSize5MB bs=1024 count=5120
dd if=/dev/zero of=FileOfSize6MB bs=1024 count=6144
dd if=/dev/zero of=FileOfSize11MB bs=1024 count=11264
dd if=/dev/zero of=FileOfSize65MB bs=1024 count=66560
dd if=/dev/zero of=FileOfSizeGt32KB bs=1024 count=33
dd if=/dev/zero of=FileOfSize100KB bs=1024 count=100
cd ../