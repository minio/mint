#!/usr/bin/env bash
#!/usr/bin/expect -f

echo $@

# settings / change this to your config
currentDir="$1/$2/$3"
LOG_DIR="$1/log/$3"

ROOT_DIR=$1
build() {
	go test -c $currentDir/api_functional_v4_test.go -o $ROOT_DIR/exe/minio.test
	go get -u github.com/minio/minio-go

}

run() {
	$ROOT_DIR/exe/minio.test -test.v 
}

build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/output.log
cat $LOG_DIR/output.log   | grep -E "Error:|FAIL" > $LOG_DIR/error.log
