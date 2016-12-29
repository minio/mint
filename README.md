# Mint
Collection of tests to detect resource leaks, gauge performance problems and overall quality of Minio server.

## Goals
- To run tests in self contained manner, with various tools pre-installed.
- To asses the quality of the Minio server product.

## How to Run
The project will be published in Docker hub after further more testing. Till then the docker image has to be built locally and run.

### Build

```sh
$ git clone https://github.com/minio/mint.git
$ cd mint
$ docker build -t minio/mint:alpha .
```

### Options

Options are provided as environment variables to the docker container. Supported envs:

 - `ENDPOINT`     - <IP/URL>:<PORT> of the Minio server on which the tests has to be run.
 - `ACCESS_KEY`   - Access Key of the server.
 - `SECRET_KEY`   - Secret Key of the server.
 - `ENABLE_HTTPS` - Optional value when set to 1 sends HTTPS requests on SSL enabled deployment.


### Run

```sh
$ docker run -e ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F  -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1  minio/mint:alpha
```
