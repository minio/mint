#!/bin/sh 

# Add Run instructions for SDK tests here.
# The Build instructions should be added to build.sh
./minio.test -test.timeout 3600s -test.v 
