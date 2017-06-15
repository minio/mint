#!/usr/bin/env bash
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

declare MINIO_JAR=minio.jar
declare OK_HTTP_JAR=okhttp.jar

cleanUp(){
    # remove binary 
    rm ./minio.jar && \
    rm ./okhttp.jar
}

build() {
	# Download Minio.jar / okhttp and compile test files.
	curl -s -o $MINIO_JAR http://repo1.maven.org/maven2/io/minio/minio/3.0.4/minio-3.0.4-all.jar && \
	curl -s -o $OK_HTTP_JAR http://central.maven.org/maven2/com/squareup/okhttp3/okhttp/3.7.0/okhttp-3.7.0.jar && \
	javac -cp $MINIO_JAR FunctionalTest.java PutObjectRunnable.java
}

run() {
	if [ -n $MINIO_JAR ]; then
		[[ "$ENABLE_HTTPS" == "1" ]] && scheme="https" || scheme="http"  
		ENDPOINT_URL=$scheme://"${SERVER_ENDPOINT}"
		java -cp $MINIO_JAR":."  FunctionalTest  "$ENDPOINT_URL" "${ACCESS_KEY}" "${SECRET_KEY}" "${S3_REGION}"
	fi
}

main() {
    logfile=$1
    errfile=$2
    
    # Build test file binary
    build >>$logfile  2>&1 || { echo "minio-java build failed."; exit 1;}
 
    # run the tests
    rc=0
    run 2>>$errfile 1>>$logfile && cleanUp || { echo "minio-java run failed."; rc=1; }
    return $rc
}

# invoke the script
main "$@"