#!/bin/bash -e
#
#  Mint (C) 2017 Minio, Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

SPOTBUGS_VERSION="4.8.6"        ## SpotBugs annotations version
JUNIT_VERSION="5.11.3"          ## JUnit 5 version
JUNIT_PLATFORM_VERSION="1.11.3" ## JUnit Platform version

MINIO_JAVA_VERSION=$(curl --retry 10 -s "https://repo1.maven.org/maven2/io/minio/minio/maven-metadata.xml" | sed -n "/<latest>/{s/<.[^>]*>//g;p;q}" | sed "s/  *//g")
if [ -z "$MINIO_JAVA_VERSION" ]; then
	echo "unable to get latest minio-java version from maven"
	exit 1
fi

test_run_dir="$MINT_RUN_CORE_DIR/minio-java"
git clone --quiet https://github.com/minio/minio-java.git "$test_run_dir/minio-java.git"
(
	cd "$test_run_dir/minio-java.git"
	git checkout --quiet "tags/${MINIO_JAVA_VERSION}"
)

# Download main dependencies
$WGET --output-document="$test_run_dir/minio-${MINIO_JAVA_VERSION}-all.jar" \
	"https://repo1.maven.org/maven2/io/minio/minio/${MINIO_JAVA_VERSION}/minio-${MINIO_JAVA_VERSION}-all.jar"
$WGET --output-document="$test_run_dir/minio-admin-${MINIO_JAVA_VERSION}-all.jar" \
	"https://repo1.maven.org/maven2/io/minio/minio-admin/${MINIO_JAVA_VERSION}/minio-admin-${MINIO_JAVA_VERSION}-all.jar"
$WGET --output-document="$test_run_dir/spotbugs-annotations-${SPOTBUGS_VERSION}.jar" \
	"https://repo1.maven.org/maven2/com/github/spotbugs/spotbugs-annotations/${SPOTBUGS_VERSION}/spotbugs-annotations-${SPOTBUGS_VERSION}.jar"

# Download JUnit 5 dependencies
$WGET --output-document="$test_run_dir/junit-jupiter-api-${JUNIT_VERSION}.jar" \
	"https://repo1.maven.org/maven2/org/junit/jupiter/junit-jupiter-api/${JUNIT_VERSION}/junit-jupiter-api-${JUNIT_VERSION}.jar"
$WGET --output-document="$test_run_dir/junit-jupiter-engine-${JUNIT_VERSION}.jar" \
	"https://repo1.maven.org/maven2/org/junit/jupiter/junit-jupiter-engine/${JUNIT_VERSION}/junit-jupiter-engine-${JUNIT_VERSION}.jar"
$WGET --output-document="$test_run_dir/junit-platform-commons-${JUNIT_PLATFORM_VERSION}.jar" \
	"https://repo1.maven.org/maven2/org/junit/platform/junit-platform-commons/${JUNIT_PLATFORM_VERSION}/junit-platform-commons-${JUNIT_PLATFORM_VERSION}.jar"
$WGET --output-document="$test_run_dir/junit-platform-engine-${JUNIT_PLATFORM_VERSION}.jar" \
	"https://repo1.maven.org/maven2/org/junit/platform/junit-platform-engine/${JUNIT_PLATFORM_VERSION}/junit-platform-engine-${JUNIT_PLATFORM_VERSION}.jar"

# Download API Guardian (required by JUnit 5)
$WGET --output-document="$test_run_dir/apiguardian-api-1.1.2.jar" \
	"https://repo1.maven.org/maven2/org/apiguardian/apiguardian-api/1.1.2/apiguardian-api-1.1.2.jar"

# Download OpenTest4J (required by JUnit 5)
$WGET --output-document="$test_run_dir/opentest4j-1.3.0.jar" \
	"https://repo1.maven.org/maven2/org/opentest4j/opentest4j/1.3.0/opentest4j-1.3.0.jar"

# Build classpath
CLASSPATH="$test_run_dir/minio-${MINIO_JAVA_VERSION}-all.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/minio-admin-${MINIO_JAVA_VERSION}-all.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/spotbugs-annotations-${SPOTBUGS_VERSION}.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/junit-jupiter-api-${JUNIT_VERSION}.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/junit-jupiter-engine-${JUNIT_VERSION}.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/junit-platform-commons-${JUNIT_PLATFORM_VERSION}.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/junit-platform-engine-${JUNIT_PLATFORM_VERSION}.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/apiguardian-api-1.1.2.jar"
CLASSPATH="$CLASSPATH:$test_run_dir/opentest4j-1.3.0.jar"

# Compile tests
javac -cp "$CLASSPATH" "${test_run_dir}/minio-java.git/functional"/*.java

# Copy compiled classes
cp -a "${test_run_dir}/minio-java.git/functional"/*.class "$test_run_dir/"

# Cleanup
rm -fr "$test_run_dir/minio-java.git"
