# Mint [![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Go Report Card](https://goreportcard.com/badge/minio/minio)](https://goreportcard.com/report/minio/minio) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/) [![codecov](https://codecov.io/gh/minio/minio/branch/master/graph/badge.svg)](https://codecov.io/gh/minio/minio)

Collection of tests to detect overall correctness of Minio server.

## Goals

- To run tests in self contained manner, with various tools pre-installed
- To assess the quality of the Minio server product

## Roadmap

- Add test cases under various categories
- Add other SDK/Client side tools to increase the test case variety
- Add bench-marking tools

## How to Run

The project will be published in Docker hub after further more testing. Till then the docker image has to be built locally and run.

### Build

```sh
$ git clone https://github.com/minio/mint.git
$ cd mint
$ docker build -t minio/mint:mint .
```

### Options

#### Env variables

Set environment variables to pass test target server details to the docker container. Supported envs:

 - `SERVER_ENDPOINT`- <IP/URL>:<PORT> of the Minio server on which the tests has to be run
 - `ACCESS_KEY`     - Access Key of the server
 - `SECRET_KEY`     - Secret Key of the server
 - `ENABLE_HTTPS`   - Optional value when set to 1 sends HTTPS requests on SSL enabled deployment

Note: With no env variables provided the tests are run on play.minio.io by default

#### Config file

Directories specified in the `config.yaml` are executed sequentially. You can comment out specific directories to remove them from execution or shuffle the directory names to change the execution order.

### Run

To run Mint image, use the `docker run` command. For example, to run Mint with Minio Play server as test target use the below command

```sh
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1 minio/mint:mint
```

After the tests are run, output is stored in `/mint/log` directory inside the container. You can access these logs via `docker cp` command. For example to store logs to `/tmp/logs` directory on your host, run

```sh
docker cp minio/mint:mint:/mint/log /tmp/logs
```

Then navigate to `/tmp/logs` directory to access the test logs.

### Current tests

Following SDKs/CLI tools are available.

- mc
- minio-go
- minio-java
- minio-js
- minio-py

### Adding tests

To add tests to an existing SDK folder:

- Navigate to specific SDK test file in the path `apps/<sdk_name>/`.
- Add test cases and update `main` method if applicable.

To add new SDK/CLI to Mint:

- Create new directory in `apps/` directory with corresponding tool name
- Add a `run.sh` script. This script should set up the SDK/CLI tool and then execute the tests
- Add an entry in `config.yaml` with name of folder, e.g test_folder
