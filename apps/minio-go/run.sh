#!/usr/bin/env bash
#!/usr/bin/expect -f

<<<<<<< HEAD
# settings / change this to your config
ROOT_DIR=$1
SDK_DIR=$2
SDK_NAME=$3

CURRENT_DIR="$ROOT_DIR/$SDK_DIR/$SDK_NAME"
LOG_DIR="$ROOT_DIR/log/$SDK_NAME"

build() {
	go test -c $CURRENT_DIR/api_functional_v4_test.go -o $ROOT_DIR/bin/minio.test
	go get -u github.com/minio/minio-go

}

run() {
	$ROOT_DIR/bin/minio.test -test.short 
}

build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/output.log
cat $LOG_DIR/output.log   | grep -E "Error:|FAIL" > $LOG_DIR/error.log
exit 0
=======
build() {
	go test -c api_functional_v4_test.go -o minio.test
}

run() {
	chmod +x ./minio.test
	./minio.test -test.short
}

main() {
    # Build test file binary
    build -s  2>&1  >| $1

    # run the tests
    run -s  2>&1  >| $1

    grep -q 'Error:|FAIL' $1 > $2

    return 0
}

# invoke the script
main "$@"
>>>>>>> e271a38... Cleanup run.sh scripts
