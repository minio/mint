#!/bin/sh

# Add Build instructions for SDK tests here.
# The Run instructions should be added to run.sh

# build minio-go functional test.
go test -c sdk-tests/minio-go-functional-test/api_functional_v4_test.go 

