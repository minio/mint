// +build mint

/*
*
*  Mint, (C) 2025 Minio, Inc.
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
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// testBucketVersioningExcludedPrefixes tests that excluded_prefixes are properly
// set and retrieved via bucket versioning configuration APIs.
// This test replicates the minio-py test_set_get_bucket_versioning test to verify
// that EOS properly returns excluded_prefixes in GetBucketVersioning API response.
func testBucketVersioningExcludedPrefixes() {
	startTime := time.Now()
	testName := "testBucketVersioningExcludedPrefixes"
	function := "GetBucketVersioning/SetBucketVersioning"

	// Initialize minio client
	endpoint := os.Getenv("SERVER_ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	secure := os.Getenv("ENABLE_HTTPS") == "1"

	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		logError(testName, function, nil, startTime, "", "MinIO client creation failed", err)
		return
	}

	// Generate unique bucket name
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "minio-go-test-")
	args := map[string]interface{}{
		"bucketName": bucketName,
	}

	ctx := context.Background()

	// Create bucket
	err = c.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		logError(testName, function, args, startTime, "", "MakeBucket failed", err)
		return
	}
	defer c.RemoveBucket(ctx, bucketName)

	// Test 1: Set versioning with excluded_prefixes
	excludedPrefixes := []minio.ExcludedPrefix{
		{Prefix: "prefix1"},
		{Prefix: "prefix2"},
	}

	versioningConfig := minio.BucketVersioningConfiguration{
		Status:           "Enabled",
		ExcludedPrefixes: excludedPrefixes,
		ExcludeFolders:   true,
	}

	err = c.SetBucketVersioning(ctx, bucketName, versioningConfig)
	if err != nil {
		logError(testName, function, args, startTime, "", fmt.Sprintf("SetBucketVersioning with excluded_prefixes failed: %v", err), err)
		return
	}

	// Get versioning configuration
	retrievedConfig, err := c.GetBucketVersioning(ctx, bucketName)
	if err != nil {
		logError(testName, function, args, startTime, "", "GetBucketVersioning failed", err)
		return
	}

	// Verify status
	if retrievedConfig.Status != "Enabled" {
		logError(testName, function, args, startTime, "", fmt.Sprintf("GetBucketVersioning status mismatch: expected 'Enabled', got '%s'", retrievedConfig.Status), nil)
		return
	}

	// Verify exclude_folders
	if !retrievedConfig.ExcludeFolders {
		logError(testName, function, args, startTime, "", fmt.Sprintf("GetBucketVersioning exclude_folders mismatch: expected true, got false"), nil)
		return
	}

	// Verify excluded_prefixes - THIS IS WHERE THE EOS BUG MANIFESTS
	if len(retrievedConfig.ExcludedPrefixes) != len(excludedPrefixes) {
		logError(testName, function, args, startTime, "",
			fmt.Sprintf("GetBucketVersioning excluded_prefixes count mismatch: expected %d, got %d. "+
				"Expected: %v, Got: %v. "+
				"EOS BUG: GetBucketVersioning returns empty excluded_prefixes array instead of configured values",
				len(excludedPrefixes), len(retrievedConfig.ExcludedPrefixes),
				excludedPrefixes, retrievedConfig.ExcludedPrefixes), nil)
		return
	}

	// Compare prefix values
	for i, expectedPrefix := range excludedPrefixes {
		if retrievedConfig.ExcludedPrefixes[i].Prefix != expectedPrefix.Prefix {
			logError(testName, function, args, startTime, "",
				fmt.Sprintf("GetBucketVersioning excluded_prefix[%d] mismatch: expected '%s', got '%s'",
					i, expectedPrefix.Prefix, retrievedConfig.ExcludedPrefixes[i].Prefix), nil)
			return
		}
	}

	// Test 2: Suspend versioning (should clear excluded_prefixes)
	suspendConfig := minio.BucketVersioningConfiguration{
		Status: "Suspended",
	}

	err = c.SetBucketVersioning(ctx, bucketName, suspendConfig)
	if err != nil {
		logError(testName, function, args, startTime, "", "SetBucketVersioning suspend failed", err)
		return
	}

	// Get versioning configuration after suspend
	retrievedConfig2, err := c.GetBucketVersioning(ctx, bucketName)
	if err != nil {
		logError(testName, function, args, startTime, "", "GetBucketVersioning after suspend failed", err)
		return
	}

	// Verify status is suspended
	if retrievedConfig2.Status != "Suspended" {
		logError(testName, function, args, startTime, "",
			fmt.Sprintf("GetBucketVersioning status after suspend mismatch: expected 'Suspended', got '%s'", retrievedConfig2.Status), nil)
		return
	}

	// Verify exclude_folders is reset
	if retrievedConfig2.ExcludeFolders {
		logError(testName, function, args, startTime, "",
			fmt.Sprintf("GetBucketVersioning exclude_folders after suspend: expected false, got true"), nil)
		return
	}

	// Verify excluded_prefixes is empty
	if len(retrievedConfig2.ExcludedPrefixes) != 0 {
		logError(testName, function, args, startTime, "",
			fmt.Sprintf("GetBucketVersioning excluded_prefixes after suspend: expected empty, got %d prefixes: %v",
				len(retrievedConfig2.ExcludedPrefixes), retrievedConfig2.ExcludedPrefixes), nil)
		return
	}

	logSuccess(testName, function, args, startTime)
}
