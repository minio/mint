# Mint [![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/mint.svg?maxAge=604800)](https://hub.docker.com/r/minio/mint/)

Collection of tests to detect overall correctness of Minio server.

## Goals

- To run tests in self contained manner, with various tools pre-installed
- To assess the quality of the Minio server product

## Roadmap

- Add test cases under categories like correctness, stress/load, etc.
- Add specific tests for distributed mode, shared-backend mode, gateway mode
- Add other SDK/Client side tools to increase the test case variety
- Add bench-marking tools

## How to Run

The project will be published in Docker hub after further testing. Till then the docker image has to be built locally and run.

### Build

```sh
$ git clone https://github.com/minio/mint.git
$ cd mint
$ docker build -t minio/mint .
```

### Options

#### Env variables

Set environment variables to pass test target server details to the docker container. Supported environment variables:

- `SERVER_ENDPOINT`- <IP/URL>:<PORT> of the Minio server on which the tests has to be run. Defaults to [Minio Play Server](play.minio.io:9000/minio/).
- `ACCESS_KEY`     - Access Key of the server. Defaults to Minio Play Access Key.
- `SECRET_KEY`     - Secret Key of the server. Defaults to Minio Play Secret Key.
- `ENABLE_HTTPS`   - Set to 1 to send HTTPS requests on SSL enabled deployment. Defaults to 0.
- `DATA_DIR`       - Data directory for SDK tests. Defaults to data directory created by `build/data/install.sh` script.
- `SKIP_TESTS`     - `','` separated list of SDKs to ignore running. Empty by default. For example, to skip `minio-js` and `aws-cli` tests, use `export SKIP_TESTS=minio-js,aws-cli`. 

Note: With no env variables provided the tests are run on play.minio.io by default

### Run

To run Mint image, use the `docker run` command. For example, to run Mint with Minio Play server as test target use the below command

```sh
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1 minio/mint 
```

After the tests are run, output is stored in `/mint/log` directory inside the container. You can access these logs via `docker cp` command. For example to store logs to `/tmp/logs` directory on your host, run

```sh
docker cp minio/mint:/mint/log /tmp/logs
```

Then navigate to `/tmp/logs` directory to access the test logs.

### Current tests

Following SDKs/CLI tools are available:

- aws-cli
- mc
- minio-go
- minio-java
- minio-js
- minio-py

### Adding tests

To add tests to an existing SDK folder:

- Navigate to specific SDK test file in the path `apps/<sdk_name>/`.
- Add test cases and update `main` method if applicable.
- Refer test data section for using existing test data.

To add new SDK/CLI to Mint:

- Create new directory in `apps/` directory with corresponding tool name
- Add a `run.sh` script. This script should set up the SDK/CLI tool and then execute the tests
- Add an entry in `config.yaml` with name of folder, e.g test_folder

### Test data

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