#!/usr/bin/env bash

ROOT_DIR="$PWD"
TEST_DIR="sdk_tests"

let "errorCounter = 0"

setup() {
	set -e

	# If S3_ADDRESS is not set the tests are run on play.minio.io by default.

	# S3_ADDRESS is passed on as env variables while starting the docker container.
	# see README.md for info on options.
	#  Note: https://play.minio.io hosts publicly available Minio server.
	if [ -z "$S3_ADDRESS" ]; then
	    export S3_ADDRESS="play.minio.io:9000"
	    export ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
	    export SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	    export ENABLE_HTTPS=1
	fi
	
}
currTest() {
	./$TEST_DIR/$1/run.sh  $ROOT_DIR  $TEST_DIR $(basename $1)

	if [ $? -ne 0 ]; then 
     echo "Error running SDK: $1" 
     let "errorCounter = errorCounter + 1" 
    else
    	echo "Successfully ran $1 SDK"
	fi
}

runTests() {
	for f in sdk_tests/*; do
    if [ -d ${f} ]; then
        # Will not run if no directories are available
       # echo "sending:"  $MINIO_ROOT_DIR "src:" $SRC_DIR "test:" $TEST_DIR  "::f=" $(basename $f)
        sdk="$(basename $f)"
        log_dir=$ROOT_DIR/log/$sdk/
        if [ ! -d $log_dir ]
  			then mkdir $log_dir
		fi
		currTest "$sdk" -s  2>&1  >| $log_dir/"$sdk"_log.txt
		if [ -s "$log_dir/error.log" ] 
 		 then 
     		#echo "$sdk tests failed to complete" 
     		let "errorCounter = errorCounter + 1" 
     	 #else
     	 	#echo "$sdk tests run successfully"
		 fi
    fi
	done
}

setup
runTests
if [ $errorCounter -ne 0 ]; then 
	exit 1
fi

exit 0