#!/usr/bin/env bash
#!/usr/bin/expect -f


build() {
	npm i -g npm-check-updates
	npm-check-updates -u
	npm install
	cd $CURRENT_DIR
	npm link 

	# TODO: set this var in top level config.yaml
    # export FUNCTIONAL_TEST_TRACE=$LOG_DIR/error.log
}

run() {
	npm test
}

main () {

    # Build test file binary
    build -s  2>&1  >| $1

    # run the tests
    run -s  2>&1  >| $1

    grep -q 'Error:|FAIL' $1 > $2

    return 0
}

# invoke the script
main "$@"
