#!/bin/bash
#
#  Minio Cloud Storage, (C) 2017 Minio, Inc.
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

# handle command line arguments
if [ $# -ne 2 ]; then
	echo "usage: run.sh <OUTPUT-LOG-FILE> <ERROR-LOG-FILE>"
	exit 1
fi

output_log_file="$1"
error_log_file="$2"

mc_variant="${MINT_MC_VARIANT:-mc}"

case "$mc_variant" in
ec)
	ec_repo_dir="${MINT_MC_EC_REPO_DIR:-./ec}"
	ec_runner="${MINT_MC_EC_RUNNER:-${ec_repo_dir}/mint/run.sh}"

	if [ ! -f "$ec_runner" ]; then
		echo "MINT_MC_VARIANT=ec but missing runner: $ec_runner" >&2
		echo "mount the ec repo at ./ec (so ${ec_repo_dir}/mint/run.sh exists) or set MINT_MC_EC_RUNNER" >&2
		exit 1
	fi

	# Use exec for the ec tests
	exec bash "$ec_runner" "$output_log_file" "$error_log_file"
	;;
mc)
	# Run upstream mc tests
	./functional-tests.sh 1>>"$output_log_file" 2>"$error_log_file"
	;;
*)
	echo "unknown MINT_MC_VARIANT: $mc_variant (supported: mc, ec)" >&2
	exit 1
	;;
esac
