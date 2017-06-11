#!/usr/bin/env bash
#!/usr/bin/expect -f

# settings / change this to your config
ROOT_DIR=$1
SDK_DIR=$2
SDK_NAME=$3

CURRENT_DIR="$ROOT_DIR/$SDK_DIR/$SDK_NAME"
LOG_DIR="$ROOT_DIR/log/$SDK_NAME"

build() {
	cp $CURRENT_DIR/package.json $ROOT_DIR/bin
	cd $ROOT_DIR/bin
	npm i -g npm-check-updates
	npm-check-updates -u
	npm install
	cd $CURRENT_DIR
	npm link 

	export FUNCTIONAL_TEST_TRACE=$LOG_DIR/error.log
}

run() {
	cd $CURRENT_DIR
	npm test
}

build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/output.log
cat $LOG_DIR/output.log   | grep -E "Error:|FAIL" >> $LOG_DIR/error.log
exit 0