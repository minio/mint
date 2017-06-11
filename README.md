# Mint
Collection of tests to detect resource leaks, gauge performance problems and overall quality of Minio server.

## Goals
- To run tests in self contained manner, with various tools pre-installed.
- To assess the quality of the Minio server product.

## How to Run
The project will be published in Docker hub after further more testing. Till then the docker image has to be built locally and run.

### Build

```sh
$ git clone https://github.com/minio/mint.git
$ cd mint
$ docker build -t minio/mint:mint2 .
```

### Options

Options are provided as environment variables to the docker container. Supported envs:

 - `SERVER_ENDPOINT`     - <IP/URL>:<PORT> of the Minio server on which the tests has to be run.
 - `ACCESS_KEY`   - Access Key of the server.
 - `SECRET_KEY`   - Secret Key of the server.
 - `ENABLE_HTTPS` - Optional value when set to 1 sends HTTPS requests on SSL enabled deployment.


### Run

```sh
$ docker run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F  -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1  -v /path/to/local/src/dir:/home  minio/mint:mint2
```

Note: With no env variables provided the tests are run on play.minio.io by default

### Current tests
- SDK Tests (Contains tests using S3 compatible client libraries)
  - mc functional tests.
  - Minio-go functional tests.
  - Minio-java functional tests.
  - Minio-js functional tests.
  - Minio-py functional tests.
 
### Adding tests. 

## Directory structure

```
/
  apps/
    sdk_folder/
       run.sh
  log/
    sdk_folder/
       output.log
       error.log
  config.yaml
  run.sh
```
- To add tests to an existing sdk folder: 
  - follow README.MD within respective sdk folder
- To add new sdk or test category to Mint, 
  - add an entry in config.yaml with name of folder to find tests, e.g test_folder
  - create a folder in apps/ directory with above name. Add a run.sh method for tests to run.
  - All output and errors will be saved in sdk_folder/test_folder 
