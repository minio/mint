/*
*
*  Mint, (C) 2017-2025 Minio, Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// different kinds of test failures
const (
	PASS = "PASS" // Indicate that a test passed
	FAIL = "FAIL" // Indicate that a test failed
)

type errorResponse struct {
	XMLName    xml.Name `xml:"Error" json:"-"`
	Code       string
	Message    string
	BucketName string
	Key        string
	RequestID  string `xml:"RequestId"`
	HostID     string `xml:"HostId"`

	// Region where the bucket is located. This header is returned
	// only in HEAD bucket and ListObjects response.
	Region string

	// Headers of the returned S3 XML error
	Headers http.Header `xml:"-" json:"-"`
}

type mintJSONFormatter struct{}

func (f *mintJSONFormatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(log.Fields, len(entry.Data))
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %w", err)
	}
	return append(serialized, '\n'), nil
}

// log successful test runs
func successLogger(function string, args map[string]interface{}, startTime time.Time) *log.Entry {
	// calculate the test case duration
	duration := time.Since(startTime)
	// log with the fields as per mint
	fields := log.Fields{"name": "aws-sdk-go-v2", "function": function, "args": args, "duration": duration.Nanoseconds() / 1000000, "status": PASS}
	return log.WithFields(fields)
}

// log failed test runs
func failureLog(function string, args map[string]interface{}, startTime time.Time, alert string, message string, err error) *log.Entry {
	// calculate the test case duration
	duration := time.Since(startTime)
	var fields log.Fields
	// log with the fields as per mint
	if err != nil {
		fields = log.Fields{
			"name": "aws-sdk-go-v2", "function": function, "args": args,
			"duration": duration.Nanoseconds() / 1000000, "status": FAIL, "alert": alert, "message": message, "error": err,
		}
	} else {
		fields = log.Fields{
			"name": "aws-sdk-go-v2", "function": function, "args": args,
			"duration": duration.Nanoseconds() / 1000000, "status": FAIL, "alert": alert, "message": message,
		}
	}
	return log.WithFields(fields)
}

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

func isObjectTaggingImplemented(ctx context.Context, s3Client *s3.Client) bool {
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	startTime := time.Now()
	function := "isObjectTaggingImplemented"
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return false
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:   strings.NewReader("testfile"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})

	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to success but got %v", err), err).Fatal()
		return false
	}

	_, err = s3Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotImplemented" {
				return false
			}
		}
	}
	return true
}

func cleanup(ctx context.Context, s3Client *s3.Client, bucket string, object string, function string,
	args map[string]interface{}, startTime time.Time, deleteBucket bool,
) {
	// Deleting the object, just in case it was created. Will not check for errors.
	s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})

	if deleteBucket {
		_, err := s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			failureLog(function, args, startTime, "", "AWS SDK Go V2 DeleteBucket Failed", err).Fatal()
			return
		}
	}
}

func testPresignedPutInvalidHash(ctx context.Context, s3Client *s3.Client, presignClient *s3.PresignClient) {
	startTime := time.Now()
	function := "PresignedPut"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := "presignedTest"
	expiry := 1 * time.Minute
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
		"expiry":     expiry,
	}

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	presignReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		ContentType: aws.String("application/octet-stream"),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 presigned Put request creation failed", err).Fatal()
		return
	}

	rreq, err := http.NewRequest(http.MethodPut, presignReq.URL, bytes.NewReader([]byte("")))
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 presigned PUT request failed", err).Fatal()
		return
	}

	rreq.Header.Set("X-Amz-Content-Sha256", "invalid-sha256")
	rreq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(rreq)
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 presigned put request failed", err).Fatal()
		return
	}
	defer resp.Body.Close()

	dec := xml.NewDecoder(resp.Body)
	errResp := errorResponse{}
	err = dec.Decode(&errResp)
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 unmarshalling xml failed", err).Fatal()
		return
	}

	if errResp.Code != "XAmzContentSHA256Mismatch" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 presigned PUT expected to fail with XAmzContentSHA256Mismatch but got %v", errResp.Code), errors.New("AWS S3 error code mismatch")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testListObjects(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testListObjects"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object1 := "testObject1"
	object2 := "testObject2"
	expiry := 1 * time.Minute
	args := map[string]interface{}{
		"bucketName":  bucket,
		"objectName1": object1,
		"objectName2": object2,
		"expiry":      expiry,
	}

	getKeys := func(objects []types.Object) []string {
		var rv []string
		for _, obj := range objects {
			rv = append(rv, *obj.Key)
		}
		return rv
	}
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object1, function, args, startTime, true)
	defer cleanup(ctx, s3Client, bucket, object2, function, args, startTime, false)

	listInput := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(1000),
		Prefix:  aws.String(""),
	}
	result, err := s3Client.ListObjectsV2(ctx, listInput)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 listobjects expected to success but got %v", err), err).Fatal()
		return
	}
	if result.KeyCount != nil && *result.KeyCount != 0 {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 listobjects with prefix '' expected 0 key but got %v, %v", *result.KeyCount, getKeys(result.Contents)), errors.New("AWS S3 key count mismatch")).Fatal()
		return
	}
	putInput1 := &s3.PutObjectInput{
		Body:   strings.NewReader("filetoupload"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object1),
	}
	_, err = s3Client.PutObject(ctx, putInput1)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to success but got %v", err), err).Fatal()
		return
	}
	putInput2 := &s3.PutObjectInput{
		Body:   strings.NewReader("filetoupload"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object2),
	}
	_, err = s3Client.PutObject(ctx, putInput2)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to success but got %v", err), err).Fatal()
		return
	}
	result, err = s3Client.ListObjectsV2(ctx, listInput)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 listobjects expected to success but got %v", err), err).Fatal()
		return
	}
	if result.KeyCount != nil && *result.KeyCount != 2 {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 listobjects with prefix '' expected 2 key but got %v, %v", *result.KeyCount, getKeys(result.Contents)), errors.New("AWS S3 key count mismatch")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testSelectObject(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testSelectObject"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object1 := "object1.csv"
	object2 := "object2.csv"
	args := map[string]interface{}{
		"bucketName":  bucket,
		"objectName1": object1,
		"objectName2": object2,
	}

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}

	// Test comma field separator
	inputCsv1 := `year,gender,ethnicity,firstname,count,rank
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,SOPHIA,119,1
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,CHLOE,106,2
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,EMILY,93,3
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,OLIVIA,89,4
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,EMMA,75,5
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,ISABELLA,67,6
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,TIFFANY,54,7
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,ASHLEY,52,8
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,FIONA,48,9
2011,FEMALE,ASIAN AND PACIFIC ISLANDER,ANGELA,47,10
`

	outputCSV1 := `2011
2011
2011
2011
2011
2011
2011
2011
2011
2011
`

	putInput1 := &s3.PutObjectInput{
		Body:   strings.NewReader(inputCsv1),
		Bucket: aws.String(bucket),
		Key:    aws.String(object1),
	}
	_, err = s3Client.PutObject(ctx, putInput1)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object failed %v", err), err).Fatal()
		return
	}

	defer cleanup(ctx, s3Client, bucket, object1, function, args, startTime, true)

	params := &s3.SelectObjectContentInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(object1),
		ExpressionType:  types.ExpressionTypeSql,
		Expression:      aws.String("SELECT s._1 FROM S3Object s"),
		RequestProgress: &types.RequestProgress{},
		InputSerialization: &types.InputSerialization{
			CompressionType: types.CompressionTypeNone,
			CSV: &types.CSVInput{
				FileHeaderInfo:  types.FileHeaderInfoIgnore,
				FieldDelimiter:  aws.String(","),
				RecordDelimiter: aws.String("\n"),
			},
		},
		OutputSerialization: &types.OutputSerialization{
			CSV: &types.CSVOutput{},
		},
	}

	resp, err := s3Client.SelectObjectContent(ctx, params)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object failed %v", err), err).Fatal()
		return
	}
	defer resp.GetStream().Close()

	payload := ""
	for event := range resp.GetStream().Events() {
		switch v := event.(type) {
		case *types.SelectObjectContentEventStreamMemberRecords:
			// Records event contains the payload
			payload = string(v.Value.Payload)
		}
	}

	if err := resp.GetStream().Err(); err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object failed %v", err), err).Fatal()
		return
	}

	if payload != outputCSV1 {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object output mismatch %v", payload), errors.New("AWS S3 select object mismatch")).Fatal()
		return
	}

	// Test unicode field separator
	inputCsv2 := `"year"╦"gender"╦"ethnicity"╦"firstname"╦"count"╦"rank"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"SOPHIA"╦"119"╦"1"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"CHLOE"╦"106"╦"2"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"EMILY"╦"93"╦"3"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"OLIVIA"╦"89"╦"4"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"EMMA"╦"75"╦"5"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"ISABELLA"╦"67"╦"6"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"TIFFANY"╦"54"╦"7"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"ASHLEY"╦"52"╦"8"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"FIONA"╦"48"╦"9"
"2011"╦"FEMALE"╦"ASIAN AND PACIFIC ISLANDER"╦"ANGELA"╦"47"╦"10"
`

	outputCSV2 := `2011
2011
2011
2011
2011
2011
2011
2011
2011
2011
`

	putInput2 := &s3.PutObjectInput{
		Body:   strings.NewReader(inputCsv2),
		Bucket: aws.String(bucket),
		Key:    aws.String(object2),
	}
	_, err = s3Client.PutObject(ctx, putInput2)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object upload failed: %v", err), err).Fatal()
		return
	}

	defer cleanup(ctx, s3Client, bucket, object2, function, args, startTime, false)

	params2 := &s3.SelectObjectContentInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(object2),
		ExpressionType:  types.ExpressionTypeSql,
		Expression:      aws.String("SELECT s._1 FROM S3Object s"),
		RequestProgress: &types.RequestProgress{},
		InputSerialization: &types.InputSerialization{
			CompressionType: types.CompressionTypeNone,
			CSV: &types.CSVInput{
				FileHeaderInfo:  types.FileHeaderInfoIgnore,
				FieldDelimiter:  aws.String("╦"),
				RecordDelimiter: aws.String("\n"),
			},
		},
		OutputSerialization: &types.OutputSerialization{
			CSV: &types.CSVOutput{},
		},
	}

	resp, err = s3Client.SelectObjectContent(ctx, params2)
	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object failed for unicode separator %v", err), err).Fatal()
		return
	}
	defer resp.GetStream().Close()

	for event := range resp.GetStream().Events() {
		switch v := event.(type) {
		case *types.SelectObjectContentEventStreamMemberRecords:
			// Records event contains the payload
			payload = string(v.Value.Payload)
		}
	}

	if err := resp.GetStream().Err(); err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object failed for unicode separator %v", err), err).Fatal()
		return
	}

	if payload != outputCSV2 {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Select object output mismatch %v", payload), errors.New("AWS S3 select object mismatch")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testObjectTagging(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testObjectTagging"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	taginput := "Tag1=Value1"
	tagInputSet := []types.Tag{
		{
			Key:   aws.String("Tag1"),
			Value: aws.String("Value1"),
		},
	}
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:    strings.NewReader("testfile"),
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Tagging: aws.String(taginput),
	})

	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to success but got %v", err), err).Fatal()
		return
	}

	tagop, err := s3Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUTObjectTagging expected to success but got %v", apiErr.ErrorCode()), err).Fatal()
			return
		}
	}
	if !reflect.DeepEqual(tagop.TagSet, tagInputSet) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUTObject Tag input did not match with GetObjectTagging output %v", nil), nil).Fatal()
		return
	}

	taginputSet1 := []types.Tag{
		{
			Key:   aws.String("Key4"),
			Value: aws.String("Value4"),
		},
	}
	_, err = s3Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Tagging: &types.Tagging{
			TagSet: taginputSet1,
		},
	})
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUTObjectTagging expected to success but got %v", apiErr.ErrorCode()), err).Fatal()
			return
		}
	}

	tagop, err = s3Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUTObjectTagging expected to success but got %v", apiErr.ErrorCode()), err).Fatal()
			return
		}
	}
	if !reflect.DeepEqual(tagop.TagSet, taginputSet1) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUTObjectTagging input did not match with GetObjectTagging output %v", nil), nil).Fatal()
		return
	}
	successLogger(function, args, startTime).Info()
}

func testObjectTaggingErrors(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testObjectTaggingErrors"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:   strings.NewReader("testfile"),
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})

	if err != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to success but got %v", err), err).Fatal()
		return
	}

	// case 1 : Too many tags > 10
	input := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{Key: aws.String("Key1"), Value: aws.String("Value3")},
				{Key: aws.String("Key2"), Value: aws.String("Value4")},
				{Key: aws.String("Key3"), Value: aws.String("Value3")},
				{Key: aws.String("Key4"), Value: aws.String("Value3")},
				{Key: aws.String("Key5"), Value: aws.String("Value3")},
				{Key: aws.String("Key6"), Value: aws.String("Value3")},
				{Key: aws.String("Key7"), Value: aws.String("Value3")},
				{Key: aws.String("Key8"), Value: aws.String("Value3")},
				{Key: aws.String("Key9"), Value: aws.String("Value3")},
				{Key: aws.String("Key10"), Value: aws.String("Value3")},
				{Key: aws.String("Key11"), Value: aws.String("Value3")},
			},
		},
	}

	_, err = s3Client.PutObjectTagging(ctx, input)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PUT expected to fail but succeeded", err).Fatal()
		return
	}

	var apiErr interface{ ErrorCode() string }
	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() != "BadRequest" {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to fail but got %v", err), err).Fatal()
			return
		}
	}

	// case 2 : Duplicate Tag Keys
	input = &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{Key: aws.String("Key1"), Value: aws.String("Value3")},
				{Key: aws.String("Key1"), Value: aws.String("Value4")},
			},
		},
	}

	_, err = s3Client.PutObjectTagging(ctx, input)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PUT expected to fail but succeeded", err).Fatal()
		return
	}

	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() != "InvalidTag" {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to fail but got %v", err), err).Fatal()
			return
		}
	}

	// case 3 : Too long Tag Key
	input = &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{
					Key:   aws.String("Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1"),
					Value: aws.String("Value3"),
				},
				{Key: aws.String("Key1"), Value: aws.String("Value4")},
			},
		},
	}

	_, err = s3Client.PutObjectTagging(ctx, input)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PUT expected to fail but succeeded", err).Fatal()
		return
	}

	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() != "InvalidTag" {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to fail but got %v", err), err).Fatal()
			return
		}
	}

	// case 4 : Too long Tag value
	input = &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{
					Key:   aws.String("Key1"),
					Value: aws.String("Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1Key1"),
				},
				{Key: aws.String("Key1"), Value: aws.String("Value4")},
			},
		},
	}

	_, err = s3Client.PutObjectTagging(ctx, input)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PUT expected to fail but succeeded", err).Fatal()
		return
	}

	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() != "InvalidTag" {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to fail but got %v", err), err).Fatal()
			return
		}
	}

	successLogger(function, args, startTime).Info()
}

// Tests bucket re-create errors.
func testCreateBucketError(ctx context.Context, s3Client *s3.Client, origRegion string) {
	// initialize logging params
	startTime := time.Now()
	function := "testMakeBucketError"
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	args := map[string]interface{}{
		"bucketName": bucketName,
	}

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint("us-west-1"),
		},
	})
	if err != nil {
		// InvalidRegion is a valid error if the endpoint doesn't support
		// different 'regions', we simply skip this test in such scenarios.
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidRegion" {
			successLogger(function, args, startTime).Info()
			return
		}
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucketName, "", function, args, startTime, true)

	_, errCreating := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if errCreating == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Should Return Error for Existing bucket", err).Fatal()
		return
	}
	// Verify valid error response from server.
	var apiErr interface{ ErrorCode() string }
	if errors.As(errCreating, &apiErr) {
		if apiErr.ErrorCode() != "BucketAlreadyExists" && apiErr.ErrorCode() != "BucketAlreadyOwnedByYou" {
			failureLog(function, args, startTime, "", "Invalid error returned by server", err).Fatal()
			return
		}
	}

	successLogger(function, args, startTime).Info()
}

func testListMultipartUploads(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testListMultipartUploads"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}
	_, errCreating := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if errCreating != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", errCreating).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	multipartUpload, err := s3Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 createMultipartupload API failed", err).Fatal()
		return
	}
	parts := make([]*string, 5)
	for i := 0; i < 5; i++ {
		result, errUpload := s3Client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(object),
			UploadId:   multipartUpload.UploadId,
			PartNumber: aws.Int32(int32(i + 1)),
			Body:       strings.NewReader("fileToUpload"),
		})
		if errUpload != nil {
			_, _ = s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(object),
				UploadId: multipartUpload.UploadId,
			})
			failureLog(function, args, startTime, "", "AWS SDK Go V2 uploadPart API failed for", errUpload).Fatal()
			return
		}
		parts[i] = result.ETag
	}

	listParts, errParts := s3Client.ListParts(ctx, &s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadId: multipartUpload.UploadId,
	})
	if errParts != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 ListPartsInput API failed for", err).Fatal()
		return
	}

	if len(parts) != len(listParts.Parts) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 ListParts.Parts len mismatch want: %d got: %d", len(parts), len(listParts.Parts)), err).Fatal()
		return
	}

	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range listParts.Parts {
		tag := parts[i]
		if *tag != *part.ETag {
			failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 ListParts.Parts output mismatch want: %#v got: %#v", tag, part.ETag), err).Fatal()
			return
		}
		completedParts[i] = types.CompletedPart{
			ETag:       part.ETag,
			PartNumber: part.PartNumber,
		}
	}

	_, err = s3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
		UploadId: multipartUpload.UploadId,
	})
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CompleteMultipartUpload is expected to fail but succeeded", errors.New("expected nil")).Fatal()
		return
	}

	var apiErr interface{ ErrorCode() string }
	if errors.As(err, &apiErr) && apiErr.ErrorCode() != "EntityTooSmall" {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CompleteMultipartUpload is expected to fail with EntityTooSmall", err).Fatal()
		return
	}

	// Error cases

	// MaxParts < 0
	lpInput := &s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadId: multipartUpload.UploadId,
		MaxParts: aws.Int32(-1),
	}
	_, err = s3Client.ListParts(ctx, lpInput)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 ListPartsInput API (MaxParts < 0) failed for", err).Fatal()
		return
	}

	// PartNumberMarker < 0
	lpInput.PartNumberMarker = aws.String("-1")
	_, err = s3Client.ListParts(ctx, lpInput)
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 ListPartsInput API (PartNumberMarker < 0) failed for", err).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testSSECopyObject(ctx context.Context, s3Client *s3.Client) {
	// initialize logging params
	startTime := time.Now()
	function := "testSSECopyObjectSourceEncrypted"
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	args := map[string]interface{}{
		"bucketName": bucketName,
		"objectName": object,
	}
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucketName, object+"-enc", function, args, startTime, true)
	defer cleanup(ctx, s3Client, bucketName, object+"-noenc", function, args, startTime, false)
	wrongSuccess := errors.New("Succeeded instead of failing. ")

	// create encrypted object
	sseCustomerKey := "32byteslongsecretkeymustbegiven2"
	_, errPutEnc := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:                 strings.NewReader("fileToUpload"),
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(object + "-enc"),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(sseCustomerKey),
	})
	if errPutEnc != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to succeed but got %v", errPutEnc), errPutEnc).Fatal()
		return
	}

	// copy the encrypted object
	_, errCopyEnc := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(sseCustomerKey),
		CopySource:           aws.String(bucketName + "/" + object + "-enc"),
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(object + "-copy"),
	})
	if errCopyEnc == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CopyObject expected to fail, but it succeeds ", wrongSuccess).Fatal()
		return
	}
	invalidSSECustomerAlgorithm := "Requests specifying Server Side Encryption with Customer provided keys must provide a valid encryption algorithm"
	if !strings.Contains(errCopyEnc.Error(), invalidSSECustomerAlgorithm) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 CopyObject expected error %v got %v", invalidSSECustomerAlgorithm, errCopyEnc), errCopyEnc).Fatal()
		return
	}

	// create non-encrypted object
	_, errPut := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:   strings.NewReader("fileToUpload"),
		Bucket: aws.String(bucketName),
		Key:    aws.String(object + "-noenc"),
	})
	if errPut != nil {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 PUT expected to succeed but got %v", errPut), errPut).Fatal()
		return
	}

	// copy the non-encrypted object
	_, errCopy := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		CopySourceSSECustomerAlgorithm: aws.String("AES256"),
		CopySourceSSECustomerKey:       aws.String(sseCustomerKey),
		SSECustomerAlgorithm:           aws.String("AES256"),
		SSECustomerKey:                 aws.String(sseCustomerKey),
		CopySource:                     aws.String(bucketName + "/" + object + "-noenc"),
		Bucket:                         aws.String(bucketName),
		Key:                            aws.String(object + "-copy"),
	})
	if errCopy == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CopyObject expected to fail, but it succeeds ", wrongSuccess).Fatal()
		return
	}
	invalidEncryptionParameters := "The encryption parameters are not applicable to this object."
	if !strings.Contains(errCopy.Error(), invalidEncryptionParameters) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 CopyObject expected error %v got %v", invalidEncryptionParameters, errCopy), errCopy).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testBasicObjectOperations(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testBasicObjectOperations"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := "test-object.txt"
	content := "Hello, MinIO with AWS SDK Go v2!"
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}

	// Create bucket
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	// PUT Object
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        strings.NewReader(content),
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PutObject Failed", err).Fatal()
		return
	}

	// GET Object
	getResult, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject Failed", err).Fatal()
		return
	}
	defer getResult.Body.Close()

	// Verify content
	body := make([]byte, len(content))
	_, err = getResult.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject read body failed", err).Fatal()
		return
	}
	if string(body) != content {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 GetObject content mismatch: expected '%s', got '%s'", content, string(body)), errors.New("content mismatch")).Fatal()
		return
	}

	// HEAD Object
	headResult, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 HeadObject Failed", err).Fatal()
		return
	}
	if *headResult.ContentType != "text/plain" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 HeadObject ContentType mismatch: expected 'text/plain', got '%s'", *headResult.ContentType), errors.New("content type mismatch")).Fatal()
		return
	}
	if *headResult.ContentLength != int64(len(content)) {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 HeadObject ContentLength mismatch: expected %d, got %d", len(content), *headResult.ContentLength), errors.New("content length mismatch")).Fatal()
		return
	}

	// DELETE Object
	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 DeleteObject Failed", err).Fatal()
		return
	}

	// Verify object is deleted
	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err == nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 Object should not exist after DELETE", errors.New("object still exists")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testGetObjectRange(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testGetObjectRange"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := "range-test-object.txt"
	content := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}

	// Create bucket
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	// PUT Object
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   strings.NewReader(content),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PutObject Failed", err).Fatal()
		return
	}

	// Test 1: Get first 10 bytes (bytes=0-9)
	getResult1, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Range:  aws.String("bytes=0-9"),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject with Range Failed", err).Fatal()
		return
	}
	defer getResult1.Body.Close()

	body1 := make([]byte, 10)
	_, err = getResult1.Body.Read(body1)
	if err != nil && err.Error() != "EOF" {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject range read failed", err).Fatal()
		return
	}
	if string(body1) != "0123456789" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 GetObject range content mismatch: expected '0123456789', got '%s'", string(body1)), errors.New("range content mismatch")).Fatal()
		return
	}

	// Test 2: Get middle 10 bytes (bytes=10-19)
	getResult2, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Range:  aws.String("bytes=10-19"),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject with Range Failed", err).Fatal()
		return
	}
	defer getResult2.Body.Close()

	body2 := make([]byte, 10)
	_, err = getResult2.Body.Read(body2)
	if err != nil && err.Error() != "EOF" {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject range read failed", err).Fatal()
		return
	}
	if string(body2) != "ABCDEFGHIJ" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 GetObject range content mismatch: expected 'ABCDEFGHIJ', got '%s'", string(body2)), errors.New("range content mismatch")).Fatal()
		return
	}

	// Test 3: Get last 10 bytes (bytes=-10)
	getResult3, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Range:  aws.String("bytes=-10"),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject with suffix Range Failed", err).Fatal()
		return
	}
	defer getResult3.Body.Close()

	body3 := make([]byte, 10)
	_, err = getResult3.Body.Read(body3)
	if err != nil && err.Error() != "EOF" {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject suffix range read failed", err).Fatal()
		return
	}
	if string(body3) != "qrstuvwxyz" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 GetObject suffix range content mismatch: expected 'qrstuvwxyz', got '%s'", string(body3)), errors.New("suffix range content mismatch")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func testObjectMetadata(ctx context.Context, s3Client *s3.Client) {
	startTime := time.Now()
	function := "testObjectMetadata"
	bucket := randString(60, rand.NewSource(time.Now().UnixNano()), "aws-sdk-go-test-")
	object := "metadata-test-object.txt"
	content := "Object with custom metadata"
	args := map[string]interface{}{
		"bucketName": bucket,
		"objectName": object,
	}

	// Create bucket
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 CreateBucket Failed", err).Fatal()
		return
	}
	defer cleanup(ctx, s3Client, bucket, object, function, args, startTime, true)

	// PUT Object with custom metadata
	metadata := map[string]string{
		"author":      "MinIO Test Suite",
		"environment": "testing",
		"version":     "1.0",
	}
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        strings.NewReader(content),
		ContentType: aws.String("text/plain"),
		Metadata:    metadata,
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 PutObject with metadata Failed", err).Fatal()
		return
	}

	// HEAD Object to retrieve metadata
	headResult, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 HeadObject Failed", err).Fatal()
		return
	}

	// Verify metadata
	if headResult.Metadata["author"] != "MinIO Test Suite" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Metadata 'author' mismatch: expected 'MinIO Test Suite', got '%s'", headResult.Metadata["author"]), errors.New("metadata mismatch")).Fatal()
		return
	}
	if headResult.Metadata["environment"] != "testing" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Metadata 'environment' mismatch: expected 'testing', got '%s'", headResult.Metadata["environment"]), errors.New("metadata mismatch")).Fatal()
		return
	}
	if headResult.Metadata["version"] != "1.0" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 Metadata 'version' mismatch: expected '1.0', got '%s'", headResult.Metadata["version"]), errors.New("metadata mismatch")).Fatal()
		return
	}

	// GET Object and verify metadata is also returned
	getResult, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		failureLog(function, args, startTime, "", "AWS SDK Go V2 GetObject Failed", err).Fatal()
		return
	}
	defer getResult.Body.Close()

	// Verify metadata from GET
	if getResult.Metadata["author"] != "MinIO Test Suite" {
		failureLog(function, args, startTime, "", fmt.Sprintf("AWS SDK Go V2 GET Metadata 'author' mismatch: expected 'MinIO Test Suite', got '%s'", getResult.Metadata["author"]), errors.New("metadata mismatch")).Fatal()
		return
	}

	successLogger(function, args, startTime).Info()
}

func main() {
	ctx := context.Background()
	endpoint := os.Getenv("SERVER_ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	secure := os.Getenv("ENABLE_HTTPS")
	if strings.HasSuffix(endpoint, ":443") {
		endpoint = strings.ReplaceAll(endpoint, ":443", "")
	}
	if strings.HasSuffix(endpoint, ":80") {
		endpoint = strings.ReplaceAll(endpoint, ":80", "")
	}
	sdkEndpoint := "http://" + endpoint
	if secure == "1" {
		sdkEndpoint = "https://" + endpoint
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               sdkEndpoint,
			HostnameImmutable: true,
			Source:            aws.EndpointSourceCustom,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an S3 service client
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Create presign client
	presignClient := s3.NewPresignClient(s3Client)

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// create custom formatter
	mintFormatter := mintJSONFormatter{}
	// set custom formatter
	log.SetFormatter(&mintFormatter)
	// log Info or above -- success cases are Info level, failures are Fatal level
	log.SetLevel(log.InfoLevel)
	// execute tests
	testBasicObjectOperations(ctx, s3Client)
	testGetObjectRange(ctx, s3Client)
	testObjectMetadata(ctx, s3Client)
	testPresignedPutInvalidHash(ctx, s3Client, presignClient)
	testListObjects(ctx, s3Client)
	testSelectObject(ctx, s3Client)
	testCreateBucketError(ctx, s3Client, "us-east-1")
	testListMultipartUploads(ctx, s3Client)
	if secure == "1" {
		testSSECopyObject(ctx, s3Client)
	}
	if isObjectTaggingImplemented(ctx, s3Client) {
		testObjectTagging(ctx, s3Client)
		testObjectTaggingErrors(ctx, s3Client)
	}
}
