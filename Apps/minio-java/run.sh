#!/usr/bin/env bash
#!/usr/bin/expect -f
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

# settings / change this to your config
ROOT_DIR=$1
SDK_DIR=$2
SDK_NAME=$3

CURRENT_DIR="$ROOT_DIR/$SDK_DIR/$SDK_NAME"
LOG_DIR="$ROOT_DIR/log/$SDK_NAME"

declare MINIO_JAR_NAME
declare OK_HTTP_JAR_NAME

build() {
	MINIO_JAR_NAME=`find $ROOT_DIR/bin -maxdepth 1 -mindepth 1  -name 'minio*.jar'`
	OK_HTTP_JAR_NAME=`find $ROOT_DIR/bin -maxdepth 1 -mindepth 1  -name 'okhttp*.jar'`

	if [ -n $MINIO_JAR_NAME ]; then 
		javac -cp $MINIO_JAR_NAME $CURRENT_DIR/FunctionalTest.java $CURRENT_DIR/PutObjectRunnable.java
	fi
}

run() {
	if [ -n $MINIO_JAR_NAME ]; then
		[[ "$S3_SECURE" == "1" ]] && scheme="https" || scheme="http"  
		cd $CURRENT_DIR
		ENDPOINT_URL=$scheme://"${S3_ADDRESS}"
		java -cp $MINIO_JAR_NAME":."  FunctionalTest  "$ENDPOINT_URL" "${ACCESS_KEY}" "${SECRET_KEY}" "${S3_REGION}"
	fi
}

build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/output.log
cat $LOG_DIR/output.log   | grep -E "Error:|FAIL" > $LOG_DIR/error.log
exit 0