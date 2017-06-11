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
	endpoint := os.Getenv("S3_ADDRESS")
	accessKeyID := os.Getenv("ACCESS_KEY")
	secretAccessKey := os.Getenv("SECRET_KEY")
	useSSL := false
	// check if ENABLE_HTTPS env flag is set.
	if os.Getenv("ENABLE_HTTPS") != "" {
		useSSL = true
	}

	madmClnt, err := madmin.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		addErrorLog(err.Error())
		os.Exit(1)
	}

	// check the status of the Minio server.
	_, err = madmClnt.ServiceStatus()
	if err != nil {
		addErrorLog(err.Error())
		os.Exit(1)
	}

	// log the success message.
	addRunLog("Target server: " + endpoint + " is reachable. Starting the tests...")
}

func addErrorLog(message string) {
	errorLogFile := os.Getenv("INIT_ERROR_LOG_FILE")
	file, err := os.OpenFile(errorLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// write the log to file
	log.SetOutput(file)
	log.Println(message)
}

func addRunLog(message string) {
	logFile := os.Getenv("INIT_LOG_FILE")
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// write the log to file
	log.SetOutput(file)
	log.Println(message)
}
