#!/bin/sh

# Add build instructions for SDK tests here.
# The Run instructions should be added to run.sh
# minio/pkg/admin is a dependency for main.go 
# admin api from the package is used to check whether Minio server with given 
# credentials is reachable.
go get -u github.com/minio/minio/pkg/madmin
# first build the main program.
go build main.go



