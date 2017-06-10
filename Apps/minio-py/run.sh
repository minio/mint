#!/usr/bin/env bash
#!/usr/bin/expect -f

# settings / change this to your config

ROOT_DIR=$1
SDK_DIR=$2
SDK_NAME=$3

CURRENT_DIR="$ROOT_DIR/$SDK_DIR/$SDK_NAME"
LOG_DIR="$ROOT_DIR/log/$SDK_NAME"

# This will be factored out when build environment is readied
build() {
	echo $CURRENT_DIR/requirements.txt
	pip3 install --user  -r  $CURRENT_DIR/requirements.txt
	pip3 install minio
}

# Run the test
run() {
	python3 $CURRENT_DIR/functional_test.py "$LOG_DIR" 
}

build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/temp.log  
cat $LOG_DIR/temp.log $LOG_DIR/output.log  | grep 'ERROR' > $LOG_DIR/error.log
rm $LOG_DIR/temp.log
exit 0