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
	"math/rand"
	"os"
	"time"

	minio "github.com/minio/minio-go"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyz01234569"
)

// randString generates random names and prepends them with a known prefix.
func randString(n int, src rand.Source, prefix string) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return prefix + string(b[0:30-len(prefix)])
}

func main() {
	endpoint := os.Getenv("SERVER_ENDPOINT")
	accessKeyID := os.Getenv("ACCESS_KEY")
	secretAccessKey := os.Getenv("SECRET_KEY")
	useSSL := false
	// check if ENABLE_HTTPS env flag is set.
	if os.Getenv("ENABLE_HTTPS") != "" {
		useSSL = true
	}

	// Instantiate new minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	// Make a new bucket to see if the server is reachable.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "minio-go-test")
	location := "us-east-1"

	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		log.Fatalln(err)
	}

	err = minioClient.RemoveBucket(bucketName)
	if err != nil {
		log.Fatalln(err)
	}

	// log the success message.
	log.Println("Target server: " + endpoint + " is reachable. Starting the tests...")
}
