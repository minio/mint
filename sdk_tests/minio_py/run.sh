#!/usr/bin/env bash
#!/usr/bin/expect -f

echo $@

# settings / change this to your config
currentDir="$1/$2/$3"
LOG_DIR="$1/log/$3"


build() {
	echo $currentDir/requirements.txt
	pip3 install --user  -r  $currentDir/requirements.txt
	pip3 install minio
}

run() {
	python3 $currentDir/functional_test.py "$LOG_DIR"
}

#build -s  2>&1  >| $LOG_DIR/build.log
run   -s  2>&1  >| $LOG_DIR/temp.log
cat $LOG_DIR/temp.log $LOG_DIR/output.log  | grep 'ERROR' > $LOG_DIR/error.log
rm $LOG_DIR/temp.log
