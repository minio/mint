# Mint [![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/mint.svg?maxAge=604800)](https://hub.docker.com/r/minio/mint/)

Mint is a testing framework for Minio object server, available as a docker image. It runs correctness, benchmarking and stress tests. Following are the SDKs/tools used in correctness tests.

- awscli
- aws-sdk-go
- aws-sdk-php
- aws-sdk-ruby
- mc
- minio-go
- minio-java
- minio-js
- minio-py
- minio-dotnet
- s3cmd

## Running Mint

Mint is run by `docker run` command which requires Docker to be installed. For Docker installation follow the steps [here](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/).

To run Mint with Minio Play server as test target,

```sh
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F \
             -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1 minio/mint
```

After the tests are run, output is stored in `/mint/log` directory inside the container. To get these logs, use `docker cp` command. For example
```sh
docker cp <container-id>:/mint/log /tmp/logs
```

### Mint environment variables

Below environment variables are required to be passed to the docker container. Supported environment variables:

| Environment variable | Description | Example |
|:--- |:--- |:--- |
| `SERVER_ENDPOINT` | Endpoint of Minio server in the format `HOST:PORT` | `play.minio.io:9000` |
| `ACCESS_KEY` | Access key of access `SERVER_ENDPOINT` | `Q3AM3UQ867SPQQA43P2F` |
| `SECRET_KEY` | Secret Key of access `SERVER_ENDPOINT` | `zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG` |
| `ENABLE_HTTPS` | (Optional) Set `1` to indicate to use HTTPS to access `SERVER_ENDPOINT`. Defaults to `0` (HTTP) | `1` |
| `MINT_MODE` | (Optional) Set mode indicating what category of tests to be run by values `core` or `full`.  Defaults to `core` | `full` |

### Mint log format

All test logs are stored in `/mint/log/log.json` as multiple JSON document.  Below is the JSON format for every entry in the log file.

| JSON field | Type | Description | Example |
|:--- |:--- |:--- |:--- |
| `name` | _string_ | Testing tool/SDK name | `"aws-sdk-php"` |
| `function` | _string_ | Test function name | `"getBucketLocation ( array $params = [] )"` |
| `args` | _object_ | (Optional) Key/Value map of arguments passed to test function | `{"Bucket":"aws-sdk-php-bucket-20341"}` |
| `duration` | _int_ | Time taken in milliseconds to run the test | `384` |
| `status` | _string_ | one of `PASS`, `FAIL` or `NA` | `"PASS"` |
| `alert` | _string_ | (Optional) Alert message indicating test failure | `"I/O error on create file"` |
| `message` | _string_ | (Optional) Any log message | `"validating checksum of downloaded object"` |
| `error` | _string_ | Detailed error message including stack trace on status `FAIL` | `"Error executing \"CompleteMultipartUpload\" on ...` |

## For Developers

### Running Mint development code

After making changes to Mint source code a local docker image can be built/run by

```sh
$ docker build -t minio/mint . -f Dockerfile.dev
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F \
             -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG \
             -e ENABLE_HTTPS=1 -e MINT_MODE=full minio/mint:latest
```

### Adding tests with new tool/SDK

Below are the steps need to be followed

* Create new app directory under [build](https://github.com/minio/mint/tree/master/build) and [run/core](https://github.com/minio/mint/tree/master/run/core) directories.
* Create `install.sh` which does installation of required tool/SDK under app directory.
* Any build and install time dependencies should be added to [install-packages.list](https://github.com/minio/mint/blob/master/install-packages.list).
* Build time dependencies should be added to [remove-packages.list](https://github.com/minio/mint/blob/master/remove-packages.list) for removal to have clean Mint docker image.
* Add `run.sh` in app directory under `run/core` which execute actual tests.

#### Test data

Tests may use pre-created data set to perform various object operations on Minio server.  Below data files are available under `/mint/data` directory.

| File name |  Size |
|:--- |:--- |
| datafile-1-b | 1B |
| datafile-10-kB |10KiB |
| datafile-33-kB |33KiB |
| datafile-100-kB |100KiB |
| datafile-1-MB |1MiB |
| datafile-1.03-MB |1.03MiB |
| datafile-5-MB |5MiB |
| datafile-6-MB |6MiB |
| datafile-11-MB |11MiB |
| datafile-65-MB |65MiB |

#### Mint image of github pull request

**Note for Developers**: On each PR sent to [Mint repository](https://github.com/minio/mint), `travis-ci` builds `mint` docker image and pushes it to `play.minio.io`, our private docker registry. You can get the `mint` image associated with your pull request by just running `docker pull play.minio.io/mint:$PULL_REQUEST_SHA`. For example

```sh
$ docker pull play.minio.io/mint:travis-f9f519cefc25f2eeb210847e782a47e466a6b79e
```
