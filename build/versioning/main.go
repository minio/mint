//  Mint, (C) 2021 Minio, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	log "github.com/sirupsen/logrus"
)

// S3 client for testing
var s3Client *s3.Client

func cleanupBucket(bucket string, function string, args map[string]interface{}, startTime time.Time) {
	start := time.Now()
	ctx := context.Background()

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}

	for time.Since(start) < 30*time.Minute {
		paginator := s3.NewListObjectVersionsPaginator(s3Client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				break
			}

			for _, v := range page.Versions {
				input := &s3.DeleteObjectInput{
					Bucket:                    &bucket,
					Key:                       v.Key,
					VersionId:                 v.VersionId,
					BypassGovernanceRetention: aws.Bool(true),
				}
				_, err := s3Client.DeleteObject(ctx, input)
				if err != nil {
					break
				}
			}
			for _, v := range page.DeleteMarkers {
				input := &s3.DeleteObjectInput{
					Bucket:                    &bucket,
					Key:                       v.Key,
					VersionId:                 v.VersionId,
					BypassGovernanceRetention: aws.Bool(true),
				}
				_, err := s3Client.DeleteObject(ctx, input)
				if err != nil {
					break
				}
			}
		}

		_, err := s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			time.Sleep(30 * time.Second)
			continue
		}
		return
	}

	failureLog(function, args, startTime, "", "Unable to cleanup bucket after compliance tests", nil).Fatal()
	return
}

func main() {
	endpoint := os.Getenv("SERVER_ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	secure := os.Getenv("ENABLE_HTTPS")
	sdkEndpoint := "http://" + endpoint
	if secure == "1" {
		sdkEndpoint = "https://" + endpoint
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create an S3 service object with custom endpoint resolver
	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(sdkEndpoint)
		o.UsePathStyle = true
	})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// create custom formatter
	mintFormatter := mintJSONFormatter{}
	// set custom formatter
	log.SetFormatter(&mintFormatter)
	// log Info or above -- success cases are Info level, failures are Fatal level
	log.SetLevel(log.InfoLevel)

	testMakeBucket()
	testPutObject()
	testPutObjectWithTaggingAndMetadata()
	testGetObject()
	testStatObject()
	testDeleteObject()
	testDeleteObjects()
	testListObjectVersionsSimple()
	testListObjectVersionsWithPrefixAndDelimiter()
	testListObjectVersionsKeysContinuation()
	testListObjectVersionsVersionIDContinuation()
	testListObjectsVersionsWithEmptyDirObject()
	testTagging()
	testLockingLegalhold()
	testPutGetRetentionCompliance()
	testPutGetDeleteRetentionGovernance()
	testLockingRetentionGovernance()
	testLockingRetentionCompliance()
}
