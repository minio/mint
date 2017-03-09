#!/bin/sh

# Fail if any of the commands exit with a non zero status.
# Halt the further execution of the script if any of the programs fail.
set -e
# run.sh controls the builds and run of entire test.
# Since conditional setting of env is not possible in Dockerfile, such checks are done here.
# This gives us fine grained control on running the tests and setting more options.
# If S3_ADDRESS is not set the tests are run on play.minio.io by default.

# S3_ADDRESS is passed on as env variables while starting the docker container.
# see README.md for info on options.
#  Note: https://play.minio.io hosts publicly available Minio server.
if [ -z "$S3_ADDRESS" ]; then
	    echo "env  S3_ADDRESS not set, running the tests on play.minio.io"
	    export S3_ADDRESS="play.minio.io:9000"
	    export ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
	    export SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	    export ENABLE_HTTPS=1
    fi

# Execute the top level build.
# build.sh builds main.go
chmod +x build.sh
# This is to avoid https://github.com/docker/docker/issues/9547
sync
# run build 
./build.sh    
# runs the `main` program which performs the intial checks.
# Further builds are not done and the test halts if 
# a. Server is not reachable.
# b. Credentials are wrong.
./main


# Build and Execute sdk-tests
chmod +x sdk-tests/build.sh
chmod +x sdk-tests/run.sh

sdk-tests/build.sh
sdk-tests/run.sh

# Build and Execute functional test 
chmod +x functional-test/build.sh
chmod +x functional-test/run.sh

# This is to avoid https://github.com/docker/docker/issues/9547
sync
# build and run the functional test.
functional-test/build.sh
functional-test/run.sh

