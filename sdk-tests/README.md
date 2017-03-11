# SDK Tests
Collection of tests derived from S3 Compatible client libraries.

# Current tests
- Minio-go functional test.

# Adding tests.
- Add the test in separate folder.
- Add the test to the list of Current tests above.
- The tests can read the server endpoint and credentials from the environment variables.
- Add build instructions for the test in build.sh
- If a dependency or a package is necessary for multiple tests add it to the Dockerfile in the root of the project, otherwise add it to build.sh in the current folder.
- Be aware that the Docker environment is Alpine while installing packages.
- Add run instruction for the test in run.sh
