## `mc` tests
This directory serves as the location for Mint tests using `mc`.  Top level `mint.sh` calls `run.sh` to execute tests.

## Adding new tests
Upstream `mc` tests live in `functional-tests.sh` (fetched from `minio/mc` during the Mint image build).

## Running tests manually
- Set environment variables `MINT_DATA_DIR`, `MINT_MODE`, `SERVER_ENDPOINT`, `ACCESS_KEY`, `SECRET_KEY`, `SERVER_REGION` and `ENABLE_HTTPS`
- Call `run.sh` with output log file and error log file. for example
```bash
export MINT_DATA_DIR=~/my-mint-dir
export MINT_MODE=core
export SERVER_ENDPOINT="play.minio.io:9000"
export ACCESS_KEY="Q3AM3UQ867SPQQA43P2F"
export SECRET_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
export ENABLE_HTTPS=1
export SERVER_REGION=us-east-1
./run.sh /tmp/output.log /tmp/error.log
```

## Running `ec` (enterprise) mc variant
If you want to run `mc` tests using the `ec` fork, set `MINT_MC_VARIANT=ec` and provide the `ec` repo inside the container.

By default `run.sh` looks for:
- `./ec/` (a checkout of the `ec` repository)

It then runs an `ec`-provided runner script at `./ec/mint/run.sh` (relative to this directory).

Override locations/behavior with:
- `MINT_MC_EC_REPO_DIR` (path to the checked out `ec` repo; default `./ec`)
- `MINT_MC_EC_RUNNER` (path to an executable runner script; default `./ec/mint/run.sh`)
