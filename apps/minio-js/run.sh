#!/usr/bin/env bash
#!/usr/bin/expect -f


build() {
	npm i -g npm-check-updates && \
	npm-check-updates -u   && \
	npm install   && \
	npm link  
}

run() {
	npm test
}

main () {
    
    logfile=$1
    errfile=$2
    
    # Build test file binary
    build >>$logfile  2>&1 || { echo "minio-js build failed."; exit 1;}

    # run the tests
    rc=0
    run 2>>$errfile 1>>$logfile || { echo "minio-js run failed.";rc=1;}
    return $rc
}

# invoke the script
main "$@"
