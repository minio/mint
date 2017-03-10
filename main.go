package main

/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"log"
	"os"

	"github.com/minio/minio/pkg/madmin"
)

func main() {
	// obtain endpoint from env.
	// it is set using `-e` when starting the container
	// check README.md for instructions.
	endpoint := os.Getenv("S3_ADDRESS")
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
	// log the success message.
	log.Println("Minio server is reachable and credentials right. Starting the tests...")
}
