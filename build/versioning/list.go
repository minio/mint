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
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Test regular listing result with simple use cases:
//
//	Upload an object ten times, delete it once (delete marker)
//	and check listing result
func testListObjectVersionsSimple() {
	startTime := time.Now()
	function := "testListObjectVersionsSimple"
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

	for i := 0; i < 10; i++ {
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
	}

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
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

	// Accumulate all versions IDs
	versionIDs := make(map[string]struct{})

	result, err := s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}

	// Check the delete marker entries
	if len(result.DeleteMarkers) != 1 {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}
	dm := result.DeleteMarkers[0]
	if !*dm.IsLatest {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}
	if *dm.Key != object {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}
	if time.Since(*dm.LastModified) > time.Hour {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}
	if *dm.VersionId == "" {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}
	versionIDs[*dm.VersionId] = struct{}{}

	// Check versions entries
	if len(result.Versions) != 10 {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected result", nil).Fatal()
		return
	}

	for _, version := range result.Versions {
		v := version
		if *v.IsLatest {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected IsLatest field", nil).Fatal()
			return
		}
		if *v.Key != object {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected Key field", nil).Fatal()
			return
		}
		if time.Since(*v.LastModified) > time.Hour {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected LastModified field", nil).Fatal()
			return
		}
		if *v.VersionId == "" {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected VersionId field", nil).Fatal()
			return
		}
		if !etagRegex.MatchString(*v.ETag) {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected ETag field", nil).Fatal()
			return
		}
		if *v.Size != 12 {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected Size field", nil).Fatal()
			return
		}
		if v.StorageClass != types.ObjectVersionStorageClassStandard {
			failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected StorageClass field", nil).Fatal()
			return
		}

		versionIDs[*v.VersionId] = struct{}{}
	}

	// Ensure that we have 11 distinct versions IDs
	if len(versionIDs) != 11 {
		failureLog(function, args, startTime, "", "ListObjectVersions didn't return 11 different version IDs", nil).Fatal()
		return
	}

	// Error cases

	// bucket-listobjects-handlers.go > ListObjectVersionsHandler > listObjectVersions
	lovInput := &s3.ListObjectVersionsInput{
		Bucket:          aws.String(bucket),
		VersionIdMarker: aws.String("test"),
	}
	_, err = s3Client.ListObjectVersions(ctx, lovInput)
	if err == nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to fail but got %v", err), err).Fatal()
		return
	}

	// bucket-listobjects-handlers.go > ListObjectVersionsHandler > validateListObjectsArgs
	lovInput.EncodingType = types.EncodingType("test")
	_, err = s3Client.ListObjectVersions(ctx, lovInput)
	if err == nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to fail but got %v", err), err).Fatal()
		return
	}

	// Second client with test credentials
	creds := credentials.NewStaticCredentialsProvider("test", "test", "")
	cfg2, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(creds))
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("LoadDefaultConfig expected to succeed but got %v", err), err).Fatal()
		return
	}
	s3ClientTest := s3.NewFromConfig(cfg2)

	// Check with a second client: bucket-listobjects-handlers.go > ListObjectVersionsHandler > checkRequestAuthType
	_, err = s3ClientTest.ListObjectVersions(ctx, lovInput)
	if err == nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to fail but got %v", err), err).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testListObjectVersionsWithPrefixAndDelimiter() {
	startTime := time.Now()
	function := "testListObjectVersionsWithPrefixAndDelimiter"
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

	for _, objectName := range []string{"dir/object", "dir/dir/object", "object"} {
		putInput := &s3.PutObjectInput{
			Body:   strings.NewReader("my content 1"),
			Bucket: aws.String(bucket),
			Key:    aws.String(objectName),
		}
		_, err = s3Client.PutObject(ctx, putInput)
		if err != nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
			return
		}
	}

	type objectResult struct {
		name     string
		isLatest bool
	}
	type listResult struct {
		versions       []objectResult
		commonPrefixes []string
	}

	simplifyListingResult := func(out *s3.ListObjectVersionsOutput) (result listResult) {
		for _, commonPrefix := range out.CommonPrefixes {
			result.commonPrefixes = append(result.commonPrefixes, *commonPrefix.Prefix)
		}
		for _, version := range out.Versions {
			result.versions = append(result.versions, objectResult{name: *version.Key, isLatest: *version.IsLatest})
		}
		return
	}

	// Recursive listing
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	result, err := s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult := simplifyListingResult(result)
	expectedResult := listResult{
		versions: []objectResult{
			{name: "dir/dir/object", isLatest: true},
			{name: "dir/object", isLatest: true},
			{name: "object", isLatest: true},
		},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	// Listing with delimiter
	input = &s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	}
	result, err = s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult = simplifyListingResult(result)
	expectedResult = listResult{
		versions: []objectResult{
			{name: "object", isLatest: true},
		},
		commonPrefixes: []string{"dir/"},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	// Listing with prefix and delimiter
	input = &s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String("dir/"),
	}
	result, err = s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult = simplifyListingResult(result)
	expectedResult = listResult{
		versions: []objectResult{
			{name: "dir/object", isLatest: true},
		},
		commonPrefixes: []string{"dir/dir/"},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

// Test if key marker continuation works in listing works well
func testListObjectVersionsKeysContinuation() {
	startTime := time.Now()
	function := "testListObjectKeysContinuation"
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

	for i := 0; i < 10; i++ {
		putInput1 := &s3.PutObjectInput{
			Body:   strings.NewReader("my content 1"),
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("testobject-%d", i)),
		}
		_, err = s3Client.PutObject(ctx, putInput1)
		if err != nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
			return
		}
	}

	input := &s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(5),
	}

	type resultPage struct {
		versions      []string
		nextKeyMarker string
		lastPage      bool
	}

	var gotResult []resultPage
	var numPages int

	paginator := s3.NewListObjectVersionsPaginator(s3Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		numPages++
		lastPage := !paginator.HasMorePages()
		rp := resultPage{lastPage: lastPage}
		if page.NextKeyMarker != nil {
			rp.nextKeyMarker = *page.NextKeyMarker
		}
		for _, v := range page.Versions {
			rp.versions = append(rp.versions, *v.Key)
		}
		if rp.nextKeyMarker != "" {
			rp.nextKeyMarker = "set"
		}
		gotResult = append(gotResult, rp)
	}

	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}

	if numPages != 2 {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected number of pages", nil).Fatal()
		return
	}

	expectedResult := []resultPage{
		{versions: []string{"testobject-0", "testobject-1", "testobject-2", "testobject-3", "testobject-4"}, nextKeyMarker: "set", lastPage: false},
		{versions: []string{"testobject-5", "testobject-6", "testobject-7", "testobject-8", "testobject-9"}, nextKeyMarker: "", lastPage: true},
	}

	if !reflect.DeepEqual(expectedResult, gotResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", fmt.Errorf("want %+v, got %+v", expectedResult, gotResult)).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

// Test if version id marker continuation works in listing works well
func testListObjectVersionsVersionIDContinuation() {
	startTime := time.Now()
	function := "testListObjectVersionIDContinuation"
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

	for i := 0; i < 10; i++ {
		putInput1 := &s3.PutObjectInput{
			Body:   strings.NewReader("my content 1"),
			Bucket: aws.String(bucket),
			Key:    aws.String("testobject"),
		}
		_, err = s3Client.PutObject(ctx, putInput1)
		if err != nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
			return
		}
	}

	input := &s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(5),
	}

	type resultPage struct {
		versions            []string
		nextVersionIDMarker string
		lastPage            bool
	}

	var gotResult []resultPage
	var gotNextVersionIDMarker string
	var numPages int

	paginator := s3.NewListObjectVersionsPaginator(s3Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		numPages++
		lastPage := !paginator.HasMorePages()
		rp := resultPage{lastPage: lastPage}
		if page.NextVersionIdMarker != nil {
			rp.nextVersionIDMarker = *page.NextVersionIdMarker
		}
		for _, v := range page.Versions {
			rp.versions = append(rp.versions, *v.Key)
		}
		if !lastPage {
			// There is only two pages, so here we are saving the version id
			// of the last element in the first page of listing
			gotNextVersionIDMarker = *page.Versions[len(page.Versions)-1].VersionId
		}
		gotResult = append(gotResult, rp)
	}

	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}

	if numPages != 2 {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected number of pages", nil).Fatal()
		return
	}

	expectedResult := []resultPage{
		{versions: []string{"testobject", "testobject", "testobject", "testobject", "testobject"}, nextVersionIDMarker: gotNextVersionIDMarker, lastPage: false},
		{versions: []string{"testobject", "testobject", "testobject", "testobject", "testobject"}, lastPage: true},
	}

	if !reflect.DeepEqual(expectedResult, gotResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

// Test listing object when there is some empty directory object
func testListObjectsVersionsWithEmptyDirObject() {
	startTime := time.Now()
	function := "testListObjectsVersionsWithEmptyDirObject"
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

	for _, objectName := range []string{"dir/object", "dir/"} {
		putInput := &s3.PutObjectInput{
			Body:   strings.NewReader(""),
			Bucket: aws.String(bucket),
			Key:    aws.String(objectName),
		}
		_, err = s3Client.PutObject(ctx, putInput)
		if err != nil {
			failureLog(function, args, startTime, "", fmt.Sprintf("PUT expected to succeed but got %v", err), err).Fatal()
			return
		}
	}

	type objectResult struct {
		name     string
		isLatest bool
	}
	type listResult struct {
		versions       []objectResult
		commonPrefixes []string
	}

	simplifyListingResult := func(out *s3.ListObjectVersionsOutput) (result listResult, err error) {
		for _, commonPrefix := range out.CommonPrefixes {
			result.commonPrefixes = append(result.commonPrefixes, *commonPrefix.Prefix)
		}
		for _, version := range out.Versions {
			if !etagRegex.MatchString(*version.ETag) {
				return listResult{}, errors.New("unexpected ETag")
			}
			result.versions = append(result.versions, objectResult{
				name:     *version.Key,
				isLatest: *version.IsLatest,
			})
		}
		return
	}

	// Recursive listing
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	result, err := s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult, err := simplifyListingResult(result)
	if err != nil {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", err).Fatal()
		return
	}
	expectedResult := listResult{
		versions: []objectResult{
			{name: "dir/", isLatest: true},
			{name: "dir/object", isLatest: true},
		},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", fmt.Errorf("want %+v, got %+v", expectedResult, gotResult)).Fatal()
		return
	}

	// Listing with delimiter
	input = &s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	}
	result, err = s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult, err = simplifyListingResult(result)
	if err != nil {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", err).Fatal()
	}
	expectedResult = listResult{
		commonPrefixes: []string{"dir/"},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	// Listing with prefix and delimiter
	input = &s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String("dir/"),
	}
	result, err = s3Client.ListObjectVersions(ctx, input)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("ListObjectVersions expected to succeed but got %v", err), err).Fatal()
		return
	}
	gotResult, err = simplifyListingResult(result)
	if err != nil {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", err).Fatal()
		return
	}
	expectedResult = listResult{
		versions: []objectResult{
			{name: "dir/", isLatest: true},
			{name: "dir/object", isLatest: true},
		},
	}
	if !reflect.DeepEqual(gotResult, expectedResult) {
		failureLog(function, args, startTime, "", "ListObjectVersions returned unexpected listing result", nil).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}
