## `minio-java` tests
This directory serves as the location for Mint tests using `minio-java`.  Top level `mint.sh` calls `run.sh` to execute tests.

## Adding new tests
New tests is added in functional tests of minio-java.  Please check https://github.com/minio/minio-java

## Running tests manually
- Set environment variables `MINT_DATA_DIR`, `MINT_MODE`, `SERVER_ENDPOINT`, `ACCESS_KEY`, `SECRET_KEY`, `SERVER_REGION`, `ENABLE_HTTPS` and `RUN_ON_FAIL`
- Call `run.sh` with output log file and error log file. for example
```bash
export MINT_DATA_DIR=~/my-mint-dir
export MINT_MODE=core
export SERVER_ENDPOINT="play.minio.io:9000"
export ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
export SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
export ENABLE_HTTPS=1
export SERVER_REGION=us-east-1
export RUN_ON_FAIL=1
./run.sh /tmp/output.log /tmp/error.log
```
