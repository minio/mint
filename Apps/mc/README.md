## `mc` tests.
This directory serves as the location for Mint tests using `mc`. To add new test case to Mint `mc` app, just add new method in the `test.sh` file.

- `mc` is already configured to point to alias `target` that points to the endpoint this Mint instance is testing.
- `run.sh` script runs all the test cases. 
- A normal response should be redirected to `$MC_LOG_FILE`, and error response should be redirected to `$MC_ERROR_LOG_FILE`.
- Do not proceed to next testcase in case of an error. 
