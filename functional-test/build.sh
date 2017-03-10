#!/bin/sh 

# Add Build instructions for SDK tests here.
# The Run instructions should be added to run.sh

# Build Minio functional tests.
go test -c functional-test/minio-functional-test/server_test.go

