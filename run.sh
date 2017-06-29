#!/bin/bash
#
#  Mint (C) 2017 Minio, Inc.
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


root_dir="$PWD"
log_dir="log"
error_file_name="error.log"
log_file_name="output.log"

# Utility methods

# Prints message after an error
printMsg() {
	echo ""
	echo "Use 'docker ps -a' to find container-id"
	echo "Export run logs from the container using 'docker cp container-id:/mint/log /tmp/mint-logs'"
}

# checks if an elements is not present in an array
containsElement () {
    local e
    for e in "${@:2}"; do [[ "$e" == "$1" ]] && return 0; done
    return 1
}

# Setup environment variables for the run.
_init() {
	set -e

	# If SERVER_ENDPOINT is not set the tests are run on play.minio.io by default.
	# SERVER_ENDPOINT is passed on as env variables while starting the docker container.
	if [ -z "$SERVER_ENDPOINT" ]; then
	    export SERVER_ENDPOINT="play.minio.io:9000"
	    export ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
	    export SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	    export ENABLE_HTTPS=1
	fi

        if [ -z "$SERVER_REGION" ]; then
            export SERVER_REGION="us-east-1"
        fi

	if [ -z "$ENABLE_HTTPS" ]; then
		ENABLE_HTTPS=0
	fi

	# mode is set via env vars
	if [ -z "$MINT_MODE" ]; then
		export MINT_MODE=core
	fi

	# init log directory
	if [ ! -d $log_dir ]; then
		mkdir $log_dir
	fi

	if [ -z "$DATA_DIR" ]; then 
		export DATA_DIR="/mint/data"
	fi
}

# Run the current SDK Test
runCoreTest() {

	# Clear log directories before run.
	local sdk_log_dir=$root_dir/$log_dir/$1

	# make and clean SDK specific log directories.
	if [ ! -d $sdk_log_dir ]
		then
			mkdir $sdk_log_dir
		else
			rm -rf $sdk_log_dir/*
	fi

	# move to SDK directory
	cd $test_dir/$1/

	# run the test
	./run.sh "$sdk_log_dir/$log_file_name" "$sdk_log_dir/$error_file_name"  && \

	# move back to top level directory
	cd ../../..
}

# Cycle through the sdk directories and run sdk/cli tests
coreMain() {

	test_dir="run/core"

    # check if any tests are ignored
    if [ -z "${IGNORE_TESTS}" ]; then
        IFS=';' read -ra SDK_TO_IGNORE <<< "${IGNORE_TESTS}"
    fi

	# read the SDKs to run
	for i in ${root_dir}/${test_dir}/*;
		do  
            # if directory exists and is not excluded
			if [ -d "${i}" ] && [ ! $(containsElement "${i}" "${SDK_TO_IGNORE[@]}") ]; then

		        # Will not run if no directories are available
		        sdk="$(basename $i)"

				echo "Running $sdk tests ..."
				# log start time
				start=$(date +%s)
				runCoreTest "$sdk" "$MINT_MODE" || { printMsg; exit 2; }
				# log end time
				end=$(date +%s)
				# get diif
				diff=$(( $end - $start ))
				echo "Finished $sdk tests in $diff seconds"
			fi
		done

	echo "Mint ran all core tests successfully. To view logs, use 'docker cp container-id:/mint/log /tmp/mint-logs'"
}

# calls subsequent test methods based on the mode.
# core is run in all modes
main() {
	# set the directories to run
	if [ "$MINT_MODE" == "core" ]; then
		coreMain
	elif [ "$MINT_MODE" == "stress" ]; then
		coreMain
		# Add stressMain here
	elif [ "$MINT_MODE" == "bench" ]; then
		coreMain
		# Add benchMain here
	else
		coreMain
		# Add other modes here
	fi
}

_init && main
