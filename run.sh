#!/usr/bin/env bash

ROOT_DIR="$PWD"
TEST_DIR="Apps"

let "errorCounter = 0"

# Setup environment variables for the run.
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
	    export S3_SECURE=1
	fi
}

# Run the current SDK Test
currTest() {
	./$TEST_DIR/$1/run.sh  $ROOT_DIR  $TEST_DIR $(basename $1)
}

# Cycle through the sdk directories and run sdk tests
runTests() {
	for i in $(yq  -r '.Apps[]' $ROOT_DIR/config.yaml ); 
		do 
			f=$ROOT_DIR/Apps/$i
			if [ -d ${f} ]; then
		        # Will not run if no directories are available
		        sdk="$(basename $f)"

		        # Clear log directories before run.
		        LOG_DIR=$ROOT_DIR/log/$sdk/
		        if [ ! -d $LOG_DIR ]
			  		then
			  			 mkdir $LOG_DIR
			  		else 
			  			rm -rf $LOG_DIR/*
				fi

				# Run test
				currTest "$sdk" -s  2>&1  >| $LOG_DIR/"$sdk"_log.txt

				# Count failed runs
				if [ -s "$LOG_DIR/error.log" ] 
		 		 then 
		     		let "errorCounter = errorCounter + 1" 
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