/*
*
*  Mint, (C) 2021 Minio, Inc.
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software

*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
*
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/aws/smithy-go"
)

func testStatObject() {
	startTime := time.Now()
	function := "testStatObject"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "versioning-test-")
	object := "testObject"
	expiry := 1 * time.Minute
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
		"expiry":     expiry,
	}
	ctx := context.Background()

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "CreateBucket failed", err).Fatal()
		return
	}
	defer cleanupBucket(bucket, function, args, startTime)

	putVersioningInput := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	}

	_, err = s3Client.PutBucketVersioning(ctx, putVersioningInput)
	if err != nil {
		if strings.Contains(err.Error(), "NotImplemented: A header you provided implies functionality that is not implemented") {
			ignoreLog(function, args, startTime, "Versioning is not implemented").Info()
			return
		}
		failureLog(function, args, startTime, "", "Put versioning failed", err).Fatal()
		return
	}

	putInput1 := &s3.PutObjectInput{
		Body:   strings.NewReader("my content 1"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	_, err = s3Client.PutObject(ctx, putInput1)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
		return
	}
	putInput2 := &s3.PutObjectInput{
		Body:   strings.NewReader("content file 2"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	_, err = s3Client.PutObject(ctx, putInput2)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
		return
	}

	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}

	_, err = s3Client.DeleteObject(ctx, deleteInput)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("Delete expected to succeed but got %v", err), err).Fatal()
		return
	}

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}

	result, err := s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}

	testCases := []struct {
		size         int64
		versionId    string
		contentType  string
		deleteMarker bool
	}{
		{0, *result.DeleteMarkers[0].VersionId, "", true},
		{14, *result.Versions[0].VersionId, "binary/octet-stream", false},
		{12, *result.Versions[1].VersionId, "binary/octet-stream", false},
	}

	for i, testCase := range testCases {
		headInput := &s3.HeadObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(object),
			VersionId: aws.String(testCase.versionId),
		}

		result, err := s3Client.HeadObject(ctx, headInput)
		if testCase.deleteMarker && err == nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) expected to fail but succeeded", i+1), nil).Fatal()
			return
		}

		if !testCase.deleteMarker && err != nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) expected to succeed but failed", i+1), err).Fatal()
			return
		}

		if testCase.deleteMarker {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				if apiErr.ErrorCode() != "MethodNotAllowed" {
					failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected error code with delete marker", i+1), err).Fatal()
					return
				}
			} else {
				failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected error with delete marker", i+1), err).Fatal()
				return
			}
			continue
		}

		if *result.ContentLength != testCase.size {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected Content-Length", i+1), err).Fatal()
			return
		}

		if !etagRegex.MatchString(*result.ETag) {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected ETag", i+1), err).Fatal()
			return
		}

		if *result.ContentType != testCase.contentType {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected Content-Type", i+1), err).Fatal()
			return
		}

		if result.DeleteMarker != nil && *result.DeleteMarker {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected DeleteMarker", i+1), err).Fatal()
			return
		}

		if time.Since(*result.LastModified) > time.Hour {
			failureLog(function, args, startTime, "", fmt.Sprintf("StatObject (%d) unexpected LastModified", i+1), err).Fatal()
			return
		}
	}

	successLogger(function, args, startTime).Info()
}
