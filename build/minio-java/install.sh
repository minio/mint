#!/bin/bash -e
#
#  Mint (C) 2017-2023 MinIO, Inc.
#
#  This file is part of MinIO Object Storage stack
#
#  This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as published by
#  the Free Software Foundation, either version 3 of the License, or
#  (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU Affero General Public License for more details.
#
#  You should have received a copy of the GNU Affero General Public License
#  along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

SPOTBUGS_VERSION="4.2.2" ## needed since 8.0.2 release
JUNIT_VERSION="4.12"     ## JUNIT version
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
$WGET --output-document="$test_run_dir/minio-${MINIO_JAVA_VERSION}-all.jar" "https://repo1.maven.org/maven2/io/minio/minio/${MINIO_JAVA_VERSION}/minio-${MINIO_JAVA_VERSION}-all.jar"
$WGET --output-document="$test_run_dir/minio-admin-${MINIO_JAVA_VERSION}-all.jar" "https://repo1.maven.org/maven2/io/minio/minio-admin/${MINIO_JAVA_VERSION}/minio-admin-${MINIO_JAVA_VERSION}-all.jar"
$WGET --output-document="$test_run_dir/spotbugs-annotations-${SPOTBUGS_VERSION}.jar" "https://repo1.maven.org/maven2/com/github/spotbugs/spotbugs-annotations/${SPOTBUGS_VERSION}/spotbugs-annotations-${SPOTBUGS_VERSION}.jar"
$WGET --output-document="$test_run_dir/junit-${JUNIT_VERSION}.jar" "https://repo1.maven.org/maven2/junit/junit/${JUNIT_VERSION}/junit-${JUNIT_VERSION}.jar"
javac -cp "$test_run_dir/minio-${MINIO_JAVA_VERSION}-all.jar:$test_run_dir/minio-admin-${MINIO_JAVA_VERSION}-all.jar:$test_run_dir/spotbugs-annotations-${SPOTBUGS_VERSION}.jar:$test_run_dir/junit-${JUNIT_VERSION}.jar" "${test_run_dir}/minio-java.git/functional"/*.java
cp -a "${test_run_dir}/minio-java.git/functional"/*.class "$test_run_dir/"
rm -fr "$test_run_dir/minio-java.git"
