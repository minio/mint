# Functional tests. 
Collection of tests with handcrafted http requests.

# Current tests
- Minio functional test. 
  Derived and modified from the functional tests in Minio server.
  Contains rich set of tests for wide variety of functionalities.

# Adding tests.
- Add the test in separate folder.
- Add the test to the list of Current tests above.
- The tests can read the server endpoint and credentials from the environment variables.
- Add build instructions for the test in build.sh
- If a dependency or a package is necessary for multiple tets add it to the Dockerfile in the root of the project, otherwise add it to build.sh in the current folder.
- Be aware that the Docker environment is Alpine while installing packages.
- Add run instruction for the test in run.sh
