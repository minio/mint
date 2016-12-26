# Mint
Collection of tests to detect resource leaks, gauge performance problems and overall quality of Minio server.

# Design.
- Mint is designed to enable easy running of series of tests built using a wide variety of tools.
- To make the task easier Mint is built as a single docker container with all the necessary dependencies and tools installed.


# How to run.
  The project will be published in Docker hub after further more testing. Till then the docker image has to be build locally and run.

- Clone the repo
  
  ```sh
  $ git clone https://github.com/minio/mint.git 
  ```

- Build and Run

  ```sh
  $ docker build -t minio/mint:alpha
  $ docker run -e ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F  -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1  mint:alpha
  ```

# Options.
  Options are passed in as environment variables to the docker containers as seen above.
  
  - `ENDPOINT`     - <IP/URL>:<PORT> of the Minio server on which the tests has to be run.
  - `ACCESS_KEY`   - Access Key of the server. 
  - `SECRET_KEY`   - Secret Key of the server.
  - `ENABLE_HTTPS` - Optional value when set to 1 sends HTTPS requests on SSL enabled deployment.
  
# Test Run.

- Minio-go functional test.

