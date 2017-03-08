package main

import (
	"github.com/minio/minio/pkg/madmin"
	"log"
	"os"
)

func main() {
	// obtain endpoint from env.
	// it is set using `-e` when starting the container
	// check README.md for instructions.
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("ACCESS_KEY")
	secretAccessKey := os.Getenv("SECRET_KEY")
	useSSL := false
	// check if ENABLE_HTTPS env flag is set.
	if os.Getenv("ENABLE_HTTPS") != "" {
		useSSL = true
	}

	// Initialize minio admin client object.
	madmClnt, err := madmin.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		// Print() followed by a call to os.Exit(1).
		log.Fatalln(err)
	}
	// check the status of the Minio server.
	_, err = madmClnt.ServiceStatus()
	if err != nil {
		// Print() followed by a call to os.Exit(1).
		log.Fatalln(err)
	}
}
