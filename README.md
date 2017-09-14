# Mint [![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/mint.svg?maxAge=604800)](https://hub.docker.com/r/minio/mint/)

Collection of tests to detect overall correctness of Minio server.

## Goals

- To run tests in self contained manner, with various tools pre-installed.
- To assess the quality of the Minio server/gateway products.

## Roadmap

- `minio-go`, `minio-dotnet` functional tests should be pulled from respective SDKs instead of local test copy in Mint repository. `minio-js`, `minio-py` & `minio-java` tests are already pulled from respective SDKs.
- Add test cases under categories like correctness, stress/load, etc.
- Add specific tests for distributed mode, shared-backend mode, gateway mode
- Add other SDK/Client side tools to increase the test case variety
- Add bench-marking tools

## Supported Environment variables

Set environment variables to pass test target server details to the docker container. Supported environment variables:

- `SERVER_ENDPOINT`- <IP/URL>:<PORT> of the Minio server on which the tests has to be run. Defaults to [Minio Play Server](play.minio.io:9000/minio/).
- `ACCESS_KEY`     - Access Key of the server. Defaults to Minio Play Access Key.
- `SECRET_KEY`     - Secret Key of the server. Defaults to Minio Play Secret Key.
- `ENABLE_HTTPS`   - Set to 1 to send HTTPS requests on SSL enabled deployment. Defaults to 0.
- `MINT_DATA_DIR`  - Data directory for SDK tests. Defaults to data directory created by `build/data/install.sh` script.

Note: With no env variables provided the tests are run on play.minio.io by default

## Run Mint

Docker is needed to run mint. Install and setup Docker by following the steps [here](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/). 

To run Mint image, use the `docker run` command. For example, to run Mint with Minio Play server as test target use the below command

```sh
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1 minio/mint:latest
```

After the tests are run, output is stored in `/mint/log` directory inside the container. You can access these logs via `docker cp` command. For example to store logs to `/tmp/logs` directory on your host, run

```sh
docker cp <container-id>:/mint/log /tmp/logs
```

Then navigate to `/tmp/logs` directory to access the test logs.

## Current tests

Following SDKs/CLI tools are available:

- awscli
- aws-sdk-php
- aws-sdk-ruby
- mc
- minio-go
- minio-java
- minio-js
- minio-py

## Adding tests

To add tests to an existing SDK folder:

- Add tests to respective SDK repository functional test.

To add new SDK/CLI tool to Mint:

- Check if the environment for the programming language is already set in `build` directory, if not, add a new directory for the language and add set up steps (including SDK/CLI tool) in `install.sh` directory.
- Create new directory in `run/core/<sdk_name>` directory with corresponding tool name.
- Add a `run.sh` script. This script should set up the SDK/CLI tool and then execute the tests

## Building Mint Docker image

### Build locally

```sh
$ git clone https://github.com/minio/mint.git
$ cd mint
$ docker build -t minio/mint . -f Dockerfile.dev
```

Developers can also customize `Dockerfile.dev` to generate smaller build images. For example, removing the following lines from `Dockerfile.dev` will avoid shipping Golang SDK tests in the image hence a faster build and tests execution time.

```docker
COPY build/go/ /mint/build/go/
RUN /mint/build/go/install.sh
```

### Build using Travis

Each pull request when submitted to Github `travis-ci` runs build on mint to create new docker image on `play.minio.io`, our private docker registry. You can get the `mint` image associated with your pull request by just running `docker pull play.minio.io/mint:$PULL_REQUEST_SHA`. For example

```sh
$ docker pull play.minio.io/mint:travis-f9f519cefc25f2eeb210847e782a47e466a6b79e
```

## Test data

All test data used by SDK tests will reside in `/mint/data/` directory on the container. To add additional test files, edit `build/data/install.sh` script

| File name |  Size
|:--- |:--- |
| datafile-1-b | 1B |
| datafile-10-kB   |10KB
| datafile-33-kB |33KB
| datafile-100-kB |100KB
| datafile-1-MB |1MB
| datafile-1.03-MB |1.03MB
| datafile-5-MB |5MB
| datafile-6-MB |6MB
| datafile-11-MB |11MB
| datafile-65-MB |65MB
