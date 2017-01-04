/*
 * Minio Cloud Storage, (C) 2015, 2016 Minio, Inc.
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

package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

// Signature and API related constants.
const (
	signV2Algorithm = "AWS"
)

// AWS Signature Version '4' constants.
const (
	signV4Algorithm = "AWS4-HMAC-SHA256"
	iso8601Format   = "20060102T150405Z"
	yyyymmdd        = "20060102"
)

var ignoredHeaders = map[string]bool{
	"Authorization":  true,
	"Content-Type":   true,
	"Content-Length": true,
	"User-Agent":     true,
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

// Whitelist resource list that will be used in query string for signature-V2 calculation.
var resourceList = []string{
	"acl",
	"delete",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"requestPayment",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}
var src = rand.NewSource(time.Now().UTC().UnixNano())

// Function to generate random string for bucket/object names.
func randString(n int) string {
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
	return string(b)
}

// generate random object name.
func getRandomObjectName() string {
	return randString(16)
}

// generate random bucket name.
func getRandomBucketName() string {
	return randString(60)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func percentEncodeSlash(s string) string {
	return strings.Replace(s, "/", "%2F", -1)
}

// APIErrorResponse - error response format
type APIErrorResponse struct {
	XMLName    xml.Name `xml:"Error" json:"-"`
	Code       string
	Message    string
	Key        string
	BucketName string
	Resource   string
	RequestID  string `xml:"RequestId"`
	HostID     string `xml:"HostId"`
}

var (
	// The maximum allowed difference between the request generation time and the server processing time
	globalMaxSkewTime = 15 * time.Minute

	// Keeps the connection active by waiting for following amount of time.
	// Primarily used in ListenBucketNotification.
	globalSNSConnAlive = 5 * time.Second
)

// ObjectIdentifier carries key name for the object to delete.
type ObjectIdentifier struct {
	ObjectName string `xml:"Key"`
}

// DeleteObjectsResponse container for multiple object deletes.
type DeleteObjectsResponse struct {
	XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ DeleteResult" json:"-"`

	// Collection of all deleted objects
	DeletedObjects []ObjectIdentifier `xml:"Deleted,omitempty"`

	// Collection of errors deleting certain objects.
	Errors []DeleteError `xml:"Error,omitempty"`
}

// DeleteError structure.
type DeleteError struct {
	Code    string
	Message string
	Key     string
}

// concurreny level for certain parallel tests.
const (
	testConcurrencyLevel = 10
)

// partInfo - represents individual part metadata.
type partInfo struct {
	// Part number that identifies the part. This is a positive integer between
	// 1 and 10,000.
	PartNumber int

	// Date and time at which the part was uploaded.
	LastModified time.Time

	// Entity tag returned when the part was initially uploaded.
	ETag string

	// Size in bytes of the part.
	Size int64
}

// uploadMetadata - represents metadata in progress multipart upload.
type uploadMetadata struct {
	// Object name for which the multipart upload was initiated.
	Object string

	// Unique identifier for this multipart upload.
	UploadID string

	// Date and time at which the multipart upload was initiated.
	Initiated time.Time

	StorageClass string // Not supported yet.
}

// completePart - completed part container.
type completePart struct {
	// Part number identifying the part. This is a positive integer between 1 and
	// 10,000
	PartNumber int

	// Entity tag returned when the part was uploaded.
	ETag string
}

// completedParts - is a collection satisfying sort.Interface.
type completedParts []completePart

func (a completedParts) Len() int           { return len(a) }
func (a completedParts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a completedParts) Less(i, j int) bool { return a[i].PartNumber < a[j].PartNumber }

// completeMultipartUpload - represents input fields for completing multipart upload.
type completeMultipartUpload struct {
	Parts []completePart `xml:"Part"`
}

// ListBucketsResponse - format for list buckets response
type ListBucketsResponse struct {
	XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListAllMyBucketsResult" json:"-"`

	Owner Owner

	// Container for one or more buckets.
	Buckets struct {
		Buckets []Bucket `xml:"Bucket"`
	} // Buckets are nested
}
type Bucket struct {
	Name         string
	CreationDate string // time string of format "2006-01-02T15:04:05.000Z"
}

// Object container for object metadata
type Object struct {
	Key          string
	LastModified string // time string of format "2006-01-02T15:04:05.000Z"
	ETag         string
	Size         int64

	// Owner of the object.
	Owner Owner

	// The class of storage used to store the object.
	StorageClass string
}

// InitiateMultipartUploadResponse container for InitiateMultiPartUpload response, provides uploadID to start MultiPart upload
type InitiateMultipartUploadResponse struct {
	XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ InitiateMultipartUploadResult" json:"-"`

	Bucket   string
	Key      string
	UploadID string `xml:"UploadId"`
}

// Owner - bucket owner/principal
type Owner struct {
	ID          string
	DisplayName string
}

// Initiator inherit from Owner struct, fields are same
type Initiator Owner

// DeleteObjectsRequest - xml carrying the object key names which needs to be deleted.
type DeleteObjectsRequest struct {
	// Element to enable quiet mode for the request
	Quiet bool
	// List of objects to be deleted
	Objects []ObjectIdentifier `xml:"Object"`
}

// CommonPrefix container for prefix response in ListObjectsResponse
type CommonPrefix struct {
	Prefix string
}

func verifyError(t *testing.T, response *http.Response, code, description string, statusCode int) {
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	errorResponse := APIErrorResponse{}
	err = xml.Unmarshal(data, &errorResponse)
	if err != nil {
		t.Fatalf("%v, %s", err, description)
	}
	if errorResponse.Code != code {
		t.Errorf("Expected response code to be %v, got %v.", code, errorResponse.Code)
	}
	if errorResponse.Message != description {
		t.Errorf("Expected response Message to be %v, got %v.", description, errorResponse.Message)
	}
	if response.StatusCode != statusCode {
		t.Errorf("Expected response status code to be %v, got %v.", statusCode, response.StatusCode)
	}
}

// queryEncode - encodes query values in their URL encoded form. In
// addition to the percent encoding performed by getURLEncodedName()
// used here, it also percent encodes '/' (forward slash)
func queryEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := percentEncodeSlash(getURLEncodedName(k)) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(percentEncodeSlash(getURLEncodedName(v)))
		}
	}
	return buf.String()
}

// construct URL for http requests for bucket operations.
func makeTestTargetURL(endPoint, bucketName, objectName string, queryValues url.Values) string {
	urlStr := endPoint + "/"
	if bucketName != "" {
		urlStr = urlStr + bucketName + "/"
	}
	if objectName != "" {
		urlStr = urlStr + getURLEncodedName(objectName)
	}
	if len(queryValues) > 0 {
		urlStr = urlStr + "?" + queryEncode(queryValues)
	}
	return urlStr
}

// return URL for uploading object into the bucket.
func getPutObjectURL(endPoint, bucketName, objectName string) string {
	return makeTestTargetURL(endPoint, bucketName, objectName, url.Values{})
}

func getPutObjectPartURL(endPoint, bucketName, objectName, uploadID, partNumber string) string {
	queryValues := url.Values{}
	queryValues.Set("uploadId", uploadID)
	queryValues.Set("partNumber", partNumber)
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValues)
}

// return URL for fetching object from the bucket.
func getGetObjectURL(endPoint, bucketName, objectName string) string {
	return makeTestTargetURL(endPoint, bucketName, objectName, url.Values{})
}

// return URL for deleting the object from the bucket.
func getDeleteObjectURL(endPoint, bucketName, objectName string) string {
	return makeTestTargetURL(endPoint, bucketName, objectName, url.Values{})
}

// return URL for deleting multiple objects from a bucket.
func getMultiDeleteObjectURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("delete", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)

}

// return URL for HEAD on the object.
func getHeadObjectURL(endPoint, bucketName, objectName string) string {
	return makeTestTargetURL(endPoint, bucketName, objectName, url.Values{})
}

// return url to be used while copying the object.
func getCopyObjectURL(endPoint, bucketName, objectName string) string {
	return makeTestTargetURL(endPoint, bucketName, objectName, url.Values{})
}

// return URL for inserting bucket notification.
func getPutNotificationURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("notification", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for fetching bucket notification.
func getGetNotificationURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("notification", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for inserting bucket policy.
func getPutPolicyURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("policy", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for fetching bucket policy.
func getGetPolicyURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("policy", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for deleting bucket policy.
func getDeletePolicyURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("policy", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for creating the bucket.
func getMakeBucketURL(endPoint, bucketName string) string {
	return makeTestTargetURL(endPoint, bucketName, "", url.Values{})
}

// return URL for listing buckets.
func getListBucketURL(endPoint string) string {
	return makeTestTargetURL(endPoint, "", "", url.Values{})
}

// return URL for HEAD on the bucket.
func getHEADBucketURL(endPoint, bucketName string) string {
	return makeTestTargetURL(endPoint, bucketName, "", url.Values{})
}

// return URL for deleting the bucket.
func getDeleteBucketURL(endPoint, bucketName string) string {
	return makeTestTargetURL(endPoint, bucketName, "", url.Values{})
}

// return URL For fetching location of the bucket.
func getBucketLocationURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("location", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for listing objects in the bucket with V1 legacy API.
func getListObjectsV1URL(endPoint, bucketName string, maxKeys string) string {
	queryValue := url.Values{}
	if maxKeys != "" {
		queryValue.Set("max-keys", maxKeys)
	}
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for listing objects in the bucket with V2 API.
func getListObjectsV2URL(endPoint, bucketName string, maxKeys string, fetchOwner string) string {
	queryValue := url.Values{}
	queryValue.Set("list-type", "2") // Enables list objects V2 URL.
	if maxKeys != "" {
		queryValue.Set("max-keys", maxKeys)
	}
	if fetchOwner != "" {
		queryValue.Set("fetch-owner", fetchOwner)
	}
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for a new multipart upload.
func getNewMultipartURL(endPoint, bucketName, objectName string) string {
	queryValue := url.Values{}
	queryValue.Set("uploads", "")
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValue)
}

// return URL for a new multipart upload.
func getPartUploadURL(endPoint, bucketName, objectName, uploadID, partNumber string) string {
	queryValues := url.Values{}
	queryValues.Set("uploadId", uploadID)
	queryValues.Set("partNumber", partNumber)
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValues)
}

// return URL for aborting multipart upload.
func getAbortMultipartUploadURL(endPoint, bucketName, objectName, uploadID string) string {
	queryValue := url.Values{}
	queryValue.Set("uploadId", uploadID)
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValue)
}

// return URL for a listing pending multipart uploads.
func getListMultipartURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("uploads", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for listing pending multipart uploads with parameters.
func getListMultipartUploadsURLWithParams(endPoint, bucketName, prefix, keyMarker, uploadIDMarker, delimiter, maxUploads string) string {
	queryValue := url.Values{}
	queryValue.Set("uploads", "")
	queryValue.Set("prefix", prefix)
	queryValue.Set("delimiter", delimiter)
	queryValue.Set("key-marker", keyMarker)
	queryValue.Set("upload-id-marker", uploadIDMarker)
	queryValue.Set("max-uploads", maxUploads)
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for a listing parts on a given upload id.
func getListMultipartURLWithParams(endPoint, bucketName, objectName, uploadID, maxParts, partNumberMarker, encoding string) string {
	queryValues := url.Values{}
	queryValues.Set("uploadId", uploadID)
	queryValues.Set("max-parts", maxParts)
	if partNumberMarker != "" {
		queryValues.Set("part-number-marker", partNumberMarker)
	}
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValues)
}

// return URL for completing multipart upload.
// complete multipart upload request is sent after all parts are uploaded.
func getCompleteMultipartUploadURL(endPoint, bucketName, objectName, uploadID string) string {
	queryValue := url.Values{}
	queryValue.Set("uploadId", uploadID)
	return makeTestTargetURL(endPoint, bucketName, objectName, queryValue)
}

// return URL for put bucket notification.
func getPutBucketNotificationURL(endPoint, bucketName string) string {
	return getGetBucketNotificationURL(endPoint, bucketName)
}

// return URL for get bucket notification.
func getGetBucketNotificationURL(endPoint, bucketName string) string {
	queryValue := url.Values{}
	queryValue.Set("notification", "")
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// return URL for listen bucket notification.
func getListenBucketNotificationURL(endPoint, bucketName string, prefixes, suffixes, events []string) string {
	queryValue := url.Values{}

	queryValue["prefix"] = prefixes
	queryValue["suffix"] = suffixes
	queryValue["events"] = events
	return makeTestTargetURL(endPoint, bucketName, "", queryValue)
}

// Return canonical resource string.
func canonicalizedResourceV2(encodedPath string, encodedQuery string) string {
	queries := strings.Split(encodedQuery, "&")
	keyval := make(map[string]string)
	for _, query := range queries {
		key := query
		val := ""
		index := strings.Index(query, "=")
		if index != -1 {
			key = query[:index]
			val = query[index+1:]
		}
		keyval[key] = val
	}
	var canonicalQueries []string
	for _, key := range resourceList {
		val, ok := keyval[key]
		if !ok {
			continue
		}
		if val == "" {
			canonicalQueries = append(canonicalQueries, key)
			continue
		}
		canonicalQueries = append(canonicalQueries, key+"="+val)
	}
	if len(canonicalQueries) == 0 {
		return encodedPath
	}
	// the queries will be already sorted as resourceList is sorted.
	return encodedPath + "?" + strings.Join(canonicalQueries, "&")
}

// Return canonical headers.
func canonicalizedAmzHeadersV2(headers http.Header) string {
	var keys []string
	keyval := make(map[string]string)
	for key := range headers {
		lkey := strings.ToLower(key)
		if !strings.HasPrefix(lkey, "x-amz-") {
			continue
		}
		keys = append(keys, lkey)
		keyval[lkey] = strings.Join(headers[key], ",")
	}
	sort.Strings(keys)
	var canonicalHeaders []string
	for _, key := range keys {
		canonicalHeaders = append(canonicalHeaders, key+":"+keyval[key])
	}
	return strings.Join(canonicalHeaders, "\n")
}

// Return string to sign for authz header calculation.
func signV2STS(method string, encodedResource string, encodedQuery string, headers http.Header) string {
	canonicalHeaders := canonicalizedAmzHeadersV2(headers)
	if len(canonicalHeaders) > 0 {
		canonicalHeaders += "\n"
	}

	// From the Amazon docs:
	//
	// StringToSign = HTTP-Verb + "\n" +
	//       Content-Md5 + "\n" +
	//       Content-Type + "\n" +
	//       Date + "\n" +
	//       CanonicalizedProtocolHeaders +
	//       CanonicalizedResource;
	stringToSign := strings.Join([]string{
		method,
		headers.Get("Content-MD5"),
		headers.Get("Content-Type"),
		headers.Get("Date"),
		canonicalHeaders,
	}, "\n") + canonicalizedResourceV2(encodedResource, encodedQuery)

	return stringToSign
}

// sum256 calculate sha256 sum for an input byte array
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// sumMD5 calculate md5 sum for an input byte array
func sumMD5(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func newTestRequest(method, urlStr string, contentLength int64, body io.ReadSeeker) (*http.Request, error) {
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Add Content-Length
	req.ContentLength = contentLength

	// Save for subsequent use
	var hashedPayload string
	switch {
	case body == nil:
		hashedPayload = hex.EncodeToString(sum256([]byte{}))
	default:
		payloadBytes, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		hashedPayload = hex.EncodeToString(sum256(payloadBytes))
		md5Base64 := base64.StdEncoding.EncodeToString(sumMD5(payloadBytes))
		req.Header.Set("Content-Md5", md5Base64)
	}
	req.Header.Set("x-amz-content-sha256", hashedPayload)
	// Seek back to beginning.
	if body != nil {
		body.Seek(0, 0)
		// Add body
		req.Body = ioutil.NopCloser(body)
	} else {
		// this is added to avoid panic during ioutil.ReadAll(req.Body).
		// th stack trace can be found here  https://github.com/minio/minio/pull/2074 .
		// This is very similar to https://github.com/golang/go/issues/7527.
		req.Body = ioutil.NopCloser(bytes.NewReader([]byte("")))
	}

	return req, nil
}

// Sign given request using Signature V2.
func signRequestV2(req *http.Request, accessKey, secretKey string) error {
	// Initial time.
	d := time.Now().UTC()

	// Add date if not present.
	if date := req.Header.Get("Date"); date == "" {
		req.Header.Set("Date", d.Format(http.TimeFormat))
	}

	// url.RawPath will be valid if path has any encoded characters, if not it will
	// be empty - in which case we need to consider url.Path (bug in net/http?)
	encodedResource := req.URL.RawPath
	encodedQuery := req.URL.RawQuery
	if encodedResource == "" {
		splits := strings.Split(req.URL.Path, "?")
		if len(splits) > 0 {
			encodedResource = splits[0]
		}
	}

	// Calculate HMAC for secretAccessKey.
	stringToSign := signV2STS(req.Method, encodedResource, encodedQuery, req.Header)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write([]byte(stringToSign))

	// Prepare auth header.
	authHeader := new(bytes.Buffer)
	authHeader.WriteString(fmt.Sprintf("%s %s:", signV2Algorithm, accessKey))
	encoder := base64.NewEncoder(base64.StdEncoding, authHeader)
	encoder.Write(hm.Sum(nil))
	encoder.Close()

	// Set Authorization header.
	req.Header.Set("Authorization", authHeader.String())
	return nil
}

// Reserved string regexp.
var reservedNames = regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")

// getURLEncodedName encode the strings from UTF-8 byte representations to HTML hex escape sequences
//
// This is necessary since regular url.Parse() and url.Encode() functions do not support UTF-8
// non english characters cannot be parsed due to the nature in which url.Encode() is written
//
// This function on the other hand is a direct replacement for url.Encode() technique to support
// pretty much every UTF-8 character.
func getURLEncodedName(name string) string {
	// if object matches reserved string, no need to encode them
	if reservedNames.MatchString(name) {
		return name
	}
	var encodedName string
	for _, s := range name {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedName = encodedName + string(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/': // §2.3 Unreserved characters (mark)
			encodedName = encodedName + string(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				return name
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedName = encodedName + "%" + strings.ToUpper(hex)
			}
		}
	}
	return encodedName
}

// Sign given request using Signature V4.
func signRequestV4(req *http.Request, accessKey, secretKey string) error {
	// Get hashed payload.
	hashedPayload := req.Header.Get("x-amz-content-sha256")
	if hashedPayload == "" {
		return fmt.Errorf("Invalid hashed payload.")
	}

	currTime := time.Now().UTC()

	// Set x-amz-date.
	req.Header.Set("x-amz-date", currTime.Format(iso8601Format))

	// Get header map.
	headerMap := make(map[string][]string)
	for k, vv := range req.Header {
		// If request header key is not in ignored headers, then add it.
		if _, ok := ignoredHeaders[http.CanonicalHeaderKey(k)]; !ok {
			headerMap[strings.ToLower(k)] = vv
		}
	}

	// Get header keys.
	headers := []string{"host"}
	for k := range headerMap {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	region := "us-east-1"

	// Get canonical headers.
	var buf bytes.Buffer
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteByte(':')
		switch {
		case k == "host":
			buf.WriteString(req.URL.Host)
			fallthrough
		default:
			for idx, v := range headerMap[k] {
				if idx > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(v)
			}
			buf.WriteByte('\n')
		}
	}
	canonicalHeaders := buf.String()

	// Get signed headers.
	signedHeaders := strings.Join(headers, ";")

	// Get canonical query string.
	req.URL.RawQuery = strings.Replace(req.URL.Query().Encode(), "+", "%20", -1)

	// Get canonical URI.
	canonicalURI := getURLEncodedName(req.URL.Path)

	// Get canonical request.
	// canonicalRequest =
	//  <HTTPMethod>\n
	//  <CanonicalURI>\n
	//  <CanonicalQueryString>\n
	//  <CanonicalHeaders>\n
	//  <SignedHeaders>\n
	//  <HashedPayload>
	//
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	}, "\n")

	// Get scope.
	scope := strings.Join([]string{
		currTime.Format(yyyymmdd),
		region,
		"s3",
		"aws4_request",
	}, "/")

	stringToSign := "AWS4-HMAC-SHA256" + "\n" + currTime.Format(iso8601Format) + "\n"
	stringToSign = stringToSign + scope + "\n"
	stringToSign = stringToSign + hex.EncodeToString(sum256([]byte(canonicalRequest)))

	date := sumHMAC([]byte("AWS4"+secretKey), []byte(currTime.Format(yyyymmdd)))
	regionHMAC := sumHMAC(date, []byte(region))
	service := sumHMAC(regionHMAC, []byte("s3"))
	signingKey := sumHMAC(service, []byte("aws4_request"))

	signature := hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))

	// final Authorization header
	parts := []string{
		"AWS4-HMAC-SHA256" + " Credential=" + accessKey + "/" + scope,
		"SignedHeaders=" + signedHeaders,
		"Signature=" + signature,
	}
	auth := strings.Join(parts, ", ")
	req.Header.Set("Authorization", auth)

	return nil
}

// Various signature types we are supporting, currently
// two main signature types.
type signerType int

const (
	signerV2 signerType = iota
	signerV4
)

func newTestSignedRequest(method, urlStr string, contentLength int64, body io.ReadSeeker, accessKey, secretKey string, signer signerType) (*http.Request, error) {
	if signer == signerV2 {
		return newTestSignedRequestV2(method, urlStr, contentLength, body, accessKey, secretKey)
	}
	return newTestSignedRequestV4(method, urlStr, contentLength, body, accessKey, secretKey)
}

// Returns new HTTP request object signed with signature v2.
func newTestSignedRequestV2(method, urlStr string, contentLength int64, body io.ReadSeeker, accessKey, secretKey string) (*http.Request, error) {
	req, err := newTestRequest(method, urlStr, contentLength, body)
	if err != nil {
		return nil, err
	}
	req.Header.Del("x-amz-content-sha256")

	// Anonymous request return quickly.
	if accessKey == "" || secretKey == "" {
		return req, nil
	}

	err = signRequestV2(req, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Returns new HTTP request object signed with signature v4.
func newTestSignedRequestV4(method, urlStr string, contentLength int64, body io.ReadSeeker, accessKey, secretKey string) (*http.Request, error) {
	req, err := newTestRequest(method, urlStr, contentLength, body)
	if err != nil {
		return nil, err
	}

	// Anonymous request return quickly.
	if accessKey == "" || secretKey == "" {
		return req, nil
	}

	err = signRequestV4(req, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// variables to store Endpoint/URL of the server to be tested,
// its access key and secret key.
var endPoint, accessKey, secretKey string
var signer signerType

// TestMain - Test execution starts here
func TestMain(m *testing.M) {
	// Get the endpoint to be tested from the environment.
	// Should have set `export S3_POINT=<IP>:<PORT>`.
	endPoint = os.Getenv("S3_ENDPOINT")

	accessKey = os.Getenv("ACCESS_KEY")
	secretKey = os.Getenv("SECRET_KEY")

	signer = signerV4

	// pasrse the env variables.
	// Run all the tests and exit.
	os.Exit(m.Run())
}

func TestBucketSQSNotification(t *testing.T) {
	// Sample bucket notification.
	bucketNotificationBuf := `<NotificationConfiguration><QueueConfiguration><Event>s3:ObjectCreated:Put</Event><Filter><S3Key><FilterRule><Name>prefix</Name><Value>images/</Value></FilterRule></S3Key></Filter><Id>1</Id><Queue>arn:minio:sqs:us-east-1:444455556666:amqp</Queue></QueueConfiguration></NotificationConfiguration>`
	// generate a random bucket Name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	request, err = newTestSignedRequest("PUT", getPutNotificationURL(endPoint, bucketName),
		int64(len(bucketNotificationBuf)), bytes.NewReader([]byte(bucketNotificationBuf)), accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)

	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "InvalidArgument", "A specified destination ARN does not exist or is not well-formed. Verify the destination ARN.", http.StatusBadRequest)
}

// TestBucketPolicy - Inserts the bucket policy and verifies it by fetching the policy back.
// Deletes the policy and verifies the deletion by fetching it back.
func TestBucketPolicy(t *testing.T) {

	// Sample bucket policy.
	bucketPolicyBuf := `{"Version":"2012-10-17","Statement":[{"Action":["s3:GetBucketLocation","s3:ListBucket"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s"],"Sid":""},{"Action":["s3:GetObject"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s/this*"],"Sid":""}]}`

	// generate a random bucket Name.
	bucketName := getRandomBucketName()
	// create the policy statement string with the randomly generated bucket name.
	bucketPolicyStr := fmt.Sprintf(bucketPolicyBuf, bucketName, bucketName)
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	/// Put a new bucket policy.
	request, err = newTestSignedRequest("PUT", getPutPolicyURL(endPoint, bucketName),
		int64(len(bucketPolicyStr)), bytes.NewReader([]byte(bucketPolicyStr)), accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to create bucket.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}

	// Fetch the uploaded policy.
	request, err = newTestSignedRequest("GET", getGetPolicyURL(endPoint, bucketName), 0, nil,
		accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	bucketPolicyReadBuf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Verify if downloaded policy matches with previousy uploaded.
	if !bytes.Equal([]byte(bucketPolicyStr), bucketPolicyReadBuf) {
		t.Fatalf("The downloaded policy doesn't match with the upload one.\nExpected:\n %s, Got: \n %s", bucketPolicyStr, string(bucketPolicyReadBuf))
	}

	// Delete policy.
	request, err = newTestSignedRequest("DELETE", getDeletePolicyURL(endPoint, bucketName), 0, nil,
		accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}
	// Verify if the policy was indeed deleted.
	request, err = newTestSignedRequest("GET", getGetPolicyURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Expected response status %v, got %v", http.StatusNotFound, response.StatusCode)
	}
}

// TestDeleteBucket - validates DELETE bucket operation.
func TestDeleteBucket(t *testing.T) {
	bucketName := getRandomBucketName()

	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// construct request to delete the bucket.
	request, err = newTestSignedRequest("DELETE", getDeleteBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the response status code.
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}
}

// TestDeleteBucketNotEmpty - Validates the operation during an attempt to delete a non-empty bucket.
func TestDeleteBucketNotEmpty(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()

	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// generate http request for an object upload.
	// "test-object" is the object name.
	objectName := "test-object"
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the request to complete object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the status code of the response.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// constructing http request to delete the bucket.
	// making an attempt to delete an non-empty bucket.
	// expected to fail.
	request, err = newTestSignedRequest("DELETE", getDeleteBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Errorf("Expected response status %s, got %s", http.StatusConflict, response.StatusCode)
	}

}

func TestListenBucketNotificationHandler(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	req, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	invalidBucket := "Invalid\\Bucket"
	tooByte := bytes.Repeat([]byte("a"), 1025)
	tooBigPrefix := string(tooByte)
	validEvents := []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"}
	invalidEvents := []string{"invalidEvent"}

	req, err = newTestSignedRequest("GET",
		getListenBucketNotificationURL(endPoint, invalidBucket, []string{}, []string{}, []string{}),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the request.
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

	req, err = newTestSignedRequest("GET",
		getListenBucketNotificationURL(endPoint, bucketName, []string{}, []string{}, invalidEvents),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the request.
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "InvalidArgument", "A specified event is not supported for notifications.", http.StatusBadRequest)

	req, err = newTestSignedRequest("GET",
		getListenBucketNotificationURL(endPoint, bucketName, []string{tooBigPrefix}, []string{}, validEvents),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the request.
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "InvalidArgument", "Size of filter rule value cannot exceed 1024 bytes in UTF-8 representation", http.StatusBadRequest)

	req, err = newTestSignedRequest("GET",
		getListenBucketNotificationURL(endPoint, bucketName, []string{}, []string{}, validEvents),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	req.Header.Set("x-amz-content-sha256", "somethingElse")
	client = &http.Client{}
	// execute the request.
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if signerV4 == signerV4 {
		verifyError(t, response, "XAmzContentSHA256Mismatch", "The provided 'x-amz-content-sha256' header does not match what was computed.", http.StatusBadRequest)
	}

	// Change global value from 5 second to 100millisecond.
	globalSNSConnAlive = 100 * time.Millisecond
	req, err = newTestSignedRequest("GET",
		getListenBucketNotificationURL(endPoint, bucketName,
			[]string{}, []string{}, validEvents), 0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the request.
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// FIXME: uncomment this in future when we have a code to read notifications from.
	// go func() {
	// 	buf := bytes.NewReader(tooByte)
	// 	rreq, rerr := newTestSignedRequest("GET",
	// 		getPutObjectURL(endPoint, bucketName, "myobject/1"),
	// 		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	// 	c.Assert(rerr, IsNil)
	// 	client = &http.Client{}
	// 	// execute the request.
	// 	resp, rerr := client.Do(rreq)
	// 	c.Assert(rerr, IsNil)
	// 	c.Assert(resp.StatusCode, Equals, http.StatusOK)
	// }()
	response.Body.Close() // FIXME. Find a way to read from the returned body.
}

// Test deletes multple objects and verifies server resonse.
func TestDeleteMultipleObjects(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "prefix/myobject"
	delObjReq := DeleteObjectsRequest{
		Quiet: false,
	}
	for i := 0; i < 10; i++ {
		// Obtain http request to upload object.
		// object Name contains a prefix.
		objName := fmt.Sprintf("%d/%s", i, objectName)
		request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objName),
			0, nil, accessKey, secretKey, signerV4)
		if err != nil {
			t.Fatalf("%v", err)
		}

		client = &http.Client{}
		// execute the http request.
		response, err = client.Do(request)
		if err != nil {
			t.Fatalf("%v", err)
		}
		// assert the status of http response.
		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
		}
		// Append all objects.
		delObjReq.Objects = append(delObjReq.Objects, ObjectIdentifier{
			ObjectName: objName,
		})
	}
	// Marshal delete request.
	deleteReqBytes, err := xml.Marshal(delObjReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Delete list of objects.
	request, err = newTestSignedRequest("POST", getMultiDeleteObjectURL(endPoint, bucketName),
		int64(len(deleteReqBytes)), bytes.NewReader(deleteReqBytes), accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var deleteResp = DeleteObjectsResponse{}
	delRespBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = xml.Unmarshal(delRespBytes, &deleteResp)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i := 0; i < 10; i++ {
		// All the objects should be under deleted list (including non-existent object)
		if !reflect.DeepEqual(deleteResp.DeletedObjects[i], delObjReq.Objects[i]) {
			t.Errorf("The objects in delete response didn't match with the ones in the response.")
		}
	}
	if len(deleteResp.Errors) != 0 {
		t.Fatalf("Expected the length of the errors to be 0, got %d", len(deleteResp.Errors))
	}

	// Attempt second time results should be same, NoSuchKey for objects not found
	// shouldn't be set.
	request, err = newTestSignedRequest("POST", getMultiDeleteObjectURL(endPoint, bucketName),
		int64(len(deleteReqBytes)), bytes.NewReader(deleteReqBytes), accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	deleteResp = DeleteObjectsResponse{}
	delRespBytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = xml.Unmarshal(delRespBytes, &deleteResp)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i := 0; i < 10; i++ {
		if !reflect.DeepEqual(deleteResp.DeletedObjects[i], delObjReq.Objects[i]) {
			t.Errorf("The objects in delete response didn't match with the ones in the response.")
		}
	}
	if len(deleteResp.Errors) != 0 {
		t.Fatalf("Expected the length of the errors to be 0, got %d", len(deleteResp.Errors))
	}
}

// Tests delete object responses and success.
func TestDeleteObject(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "prefix/myobject"
	// obtain http request to upload object.
	// object Name contains a prefix.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the http request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the status of http response.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// object name was "prefix/myobject", an attempt to delelte "prefix"
	// Should not delete "prefix/myobject"
	request, err = newTestSignedRequest("DELETE", getDeleteObjectURL(endPoint, bucketName, "prefix"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}
	// create http request to HEAD on the object.
	// this helps to validate the existence of the bucket.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the HTTP response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// create HTTP request to delete the object.
	request, err = newTestSignedRequest("DELETE", getDeleteObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the http request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}
	// Delete of non-existent data should return success.
	request, err = newTestSignedRequest("DELETE", getDeleteObjectURL(endPoint, bucketName, "prefix/myobject1"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the http request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status.
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response.StatusCode)
	}
}

// TestNonExistentBucket - Asserts response for HEAD on non-existent bucket.
func TestNonExistentBucket(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// create request to HEAD on the bucket.
	// HEAD on an bucket helps validate the existence of the bucket.
	request, err := newTestSignedRequest("HEAD", getHEADBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the http request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the response.
	// assert the http response status code.
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Expected response status %s, got %s", http.StatusNotFound, response.StatusCode)
	}
}

// TestEmptyObject - Asserts the response for operation on a 0 byte object.
func TestEmptyObject(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the http request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-object"
	// construct http request for uploading the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the upload request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// make HTTP request to fetch the object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the http request to fetch object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	// extract the body of the response.
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response body content.
	if !bytes.Equal(responseBody, buffer.Bytes()) {
		t.Errorf("Response Body doesn't match with the expected value.")
	}
}

func TestBucket(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()

	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	request, err = newTestSignedRequest("HEAD", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
}

// Tests get anonymous object.
func TestObjectGetAnonymous(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	buffer := bytes.NewReader([]byte("hello world"))
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the make bucket http request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response http status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "testObject"
	// create HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer.Len()), buffer, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the HTTP response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// initiate anonymous HTTP request to fetch the object which does not exist. We need to return AccessDenied.
	response, err = client.Get(getGetObjectURL(endPoint, bucketName, objectName+".1"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	verifyError(t, response, "AccessDenied", "Access Denied.", http.StatusForbidden)

	// initiate anonymous HTTP request to fetch the object which does exist. We need to return AccessDenied.
	response, err = client.Get(getGetObjectURL(endPoint, bucketName, objectName))
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the http response status code.
	verifyError(t, response, "AccessDenied", "Access Denied.", http.StatusForbidden)
}

// TestGetObject - Tests fetching of a small object after its insertion into the bucket.
func TestObjectGet(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	buffer := bytes.NewReader([]byte("hello world"))
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the make bucket http request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response http status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "testObject"
	// create HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer.Len()), buffer, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the HTTP response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// concurrently reading the object, safety check for races.
	var wg sync.WaitGroup
	for i := 0; i < testConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// HTTP request to create the bucket.
			// create HTTP request to fetch the object.
			getRequest, err := newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
				0, nil, accessKey, secretKey, signerV4)
			if err != nil {
				t.Fatalf("%v", err)
			}

			reqclient := &http.Client{}
			// execute the http request to fetch the object.
			getResponse, err := reqclient.Do(getRequest)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer getResponse.Body.Close()
			// assert the http response status code.
			if getResponse.StatusCode != http.StatusOK {
				t.Errorf("Expected response status to be %d, got %d.", http.StatusOK, getResponse.StatusCode)
			}

			// extract response body content.
			responseBody, err := ioutil.ReadAll(getResponse.Body)
			if err != nil {
				t.Fatalf("%v", err)
			}
			// assert the HTTP response body content with the expected content.
			if !bytes.Equal(responseBody, []byte("hello world")) {
				t.Errorf("The responseBody doesn't match the expected value.")
			}
		}()

	}
	wg.Wait()
}

// TestMultipleObjects - Validates upload and fetching of multiple object into the bucket.
func TestMultipleObjects(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create the bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// constructing HTTP request to fetch a non-existent object.
	// expected to fail, error response asserted for expected error values later.
	objectName := "testObject"
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Asserting the error response with the expected values.
	verifyError(t, response, "NoSuchKey", "The specified key does not exist.", http.StatusNotFound)

	objectName = "testObject1"
	// content for the object to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello one"))
	// create HTTP request for the object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the returned values.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// create HTTP request to fetch the object which was uploaded above.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert whether 200 OK response status is obtained.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// extract the response body.
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the content body for the expected object data.
	if !bytes.Equal(responseBody, []byte("hello one")) {
		t.Fatalf("The expected response content doesn't match with the actual one.")
	}

	// data for new object to be uploaded.
	buffer2 := bytes.NewReader([]byte("hello two"))
	objectName = "testObject2"
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response status code for expected value 200 OK.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// fetch the object which was uploaded above.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to fetch the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// assert the response status code for expected value 200 OK.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// verify response data
	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(responseBody, []byte("hello two")) {
		t.Fatalf("The expected response content doesn't match with the actual one.")
	}

	// data for new object to be uploaded.
	buffer3 := bytes.NewReader([]byte("hello three"))
	objectName = "testObject3"
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer3.Len()), buffer3, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// verify the response code with the expected value of 200 OK.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// fetch the object which was uploaded above.
	request, err = newTestSignedRequest("GET", getPutObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// verify object.
	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(responseBody, []byte("hello three")) {
		t.Fatalf("The expected response content doesn't match with the actual one.")
	}
}

// TestNotImplemented - validates if object policy is implemented, should return 'NotImplemented'.
func TestNotImplemented(t *testing.T) {
	// Generate a random bucket name.
	bucketName := getRandomBucketName()
	request, err := newTestSignedRequest("GET", endPoint+"/"+bucketName+"/object?policy",
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected response status %s, got %s", http.StatusNotImplemented, response.StatusCode)
	}
}

// TestHeader - Validates the error response for an attempt to fetch non-existent object.
func TestHeader(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// obtain HTTP request to fetch an object from non-existent bucket/object.
	request, err := newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, "testObject"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// asserting for the expected error response.
	verifyError(t, response, "NoSuchBucket", "The specified bucket does not exist", http.StatusNotFound)
}

func TestPutBucket(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// Block 1: Testing for racey access
	// The assertion is removed from this block since the purpose of this block is to find races
	// The purpose this block is not to check for correctness of functionality
	// Run the test with -race flag to utilize this
	var wg sync.WaitGroup
	for i := 0; i < testConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// HTTP request to create the bucket.
			request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
				0, nil, accessKey, secretKey, signerV4)
			if err != nil {
				t.Fatalf("%v", err)
			}

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				t.Fatalf("Put bucket Failed: <ERROR> %s", err)
			}
			defer response.Body.Close()
		}()
	}
	wg.Wait()

	bucketName = getRandomBucketName()
	//Block 2: testing for correctness of the functionality
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	response.Body.Close()
}

// TestCopyObject - Validates copy object.
// The following is the test flow.
// 1. Create bucket.
// 2. Insert Object.
// 3. Use "X-Amz-Copy-Source" header to copy the previously created object.
// 4. Validate the content of copied object.
func TestCopyObject(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// content for the object to be created.
	buffer1 := bytes.NewReader([]byte("hello world"))
	objectName := "testObject"
	// create HTTP request for object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	request.Header.Set("Content-Type", "application/json")
	if signerV4 == signerV2 {
		if err != nil {
			t.Fatalf("%v", err)
		}
		err = signRequestV2(request, accessKey, secretKey)
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName2 := "testObject2"
	// Unlike the actual PUT object request, the request to Copy Object doesn't contain request body,
	// empty body with the "X-Amz-Copy-Source" header pointing to the object to copies it in the backend.
	request, err = newTestRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName2), 0, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// setting the "X-Amz-Copy-Source" to allow copying the content of previously uploaded object.
	request.Header.Set("X-Amz-Copy-Source", url.QueryEscape("/"+bucketName+"/"+objectName))
	if signerV4 == signerV4 {
		err = signRequestV4(request, accessKey, secretKey)
	} else {
		err = signRequestV2(request, accessKey, secretKey)
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	// the content is expected to have the content of previous disk.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// creating HTTP request to fetch the previously uploaded object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName2),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// executing the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// validating the response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// reading the response body.
	// response body is expected to have the copied content of the first uploaded object.
	object, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if string(object) != "hello world" {
		t.Errorf("Expected response body doesn't match with actual one.")
	}

}

// TestPutObject -  Tests successful put object request.
func TestPutObject(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// content for new object upload.
	buffer1 := bytes.NewReader([]byte("hello world"))
	objectName := "testObject"
	// creating HTTP request for object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// fetch the object back and verify its contents.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to fetch the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	if response.ContentLength != int64(len([]byte("hello world"))) {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	var buffer2 bytes.Buffer
	// retrive the contents of response body.
	n, err := io.Copy(&buffer2, response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if n != int64(len([]byte("hello world"))) {
		t.Errorf("Expected length of the response body to be %v, got %v.", len([]byte("hello world")), n)
	}
	// asserted the contents of the fetched object with the expected result.
	if !bytes.Equal(buffer2.Bytes(), []byte("hello world")) {
		t.Errorf("contents of the fetched object doesn't match with the expected result.")
	}
}

// TestListBuckets - Make request for listing of all buckets.
// XML response is parsed.
// Its success verifies the format of the response.
func TestListBuckets(t *testing.T) {
	// create HTTP request for listing buckets.
	request, err := newTestSignedRequest("GET", getListBucketURL(endPoint),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to list buckets.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var results ListBucketsResponse
	// parse the list bucket response.
	decoder := xml.NewDecoder(response.Body)
	err = decoder.Decode(&results)
	// validating that the xml-decoding/parsing was successful.
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// This tests validate if PUT handler can successfully detect signature mismatch.
func TestValidateSignature(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// Execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objName := "test-object"

	// Body is on purpose set to nil so that we get payload generated for empty bytes.

	// Create new HTTP request with incorrect secretKey to generate an incorrect signature.
	secretKey := secretKey + "a"
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objName), 0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "SignatureDoesNotMatch", "The request signature we calculated does not match the signature you provided. Check your key and signing method.", http.StatusForbidden)
}

// This tests validate if PUT handler can successfully detect SHA256 mismatch.
func TestSHA256Mismatch(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// Execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objName := "test-object"

	// Body is on purpose set to nil so that we get payload generated for empty bytes.

	// Create new HTTP request with incorrect secretKey to generate an incorrect signature.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objName), 0, nil, accessKey, secretKey, signerV4)
	if signer == signerV4 {
		if request.Header.Get("x-amz-content-sha256") != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			t.Errorf("x-amz-content-sha256 header doesn't match with the expected one.")
		}
	}
	// Set the body to generate signature mismatch.
	request.Body = ioutil.NopCloser(bytes.NewReader([]byte("Hello, World")))
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if signer == signerV4 {
		verifyError(t, response, "XAmzContentSHA256Mismatch", "The provided 'x-amz-content-sha256' header does not match what was computed.", http.StatusBadRequest)
	}
}

// TestNotBeAbleToCreateObjectInNonexistentBucket - Validates the error response
// on an attempt to upload an object into a non-existent bucket.
func TestPutObjectLongName(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// Execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// Content for the object to be uploaded.
	buffer := bytes.NewReader([]byte("hello world"))
	// make long object name.
	longObjName := fmt.Sprintf("%0255d/%0255d/%0255d", 1, 1, 1)
	// create new HTTP request to insert the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, longObjName),
		int64(buffer.Len()), buffer, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// make long object name.
	longObjName = fmt.Sprintf("%0256d", 1)
	buffer = bytes.NewReader([]byte("hello world"))
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, longObjName),
		int64(buffer.Len()), buffer, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "XMinioInvalidObjectName", "Object name contains unsupported characters. Unsupported characters are `^*|\\\"", http.StatusBadRequest)
}

// TestNotBeAbleToCreateObjectInNonexistentBucket - Validates the error response
// on an attempt to upload an object into a non-existent bucket.
func TestNotBeAbleToCreateObjectInNonexistentBucket(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// content of the object to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))

	// preparing for upload by generating the upload URL.
	objectName := "test-object"
	request, err := newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// Execute the HTTP request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the response error message.
	verifyError(t, response, "NoSuchBucket", "The specified bucket does not exist", http.StatusNotFound)
}

// TestHeadOnObjectLastModified - Asserts response for HEAD on an object.
// HEAD requests on an object validates the existence of the object.
// The responses for fetching the object when If-Modified-Since
// and If-Unmodified-Since headers set are validated.
// If-Modified-Since - Return the object only if it has been modified since the specified time, else return a 304 (not modified).
// If-Unmodified-Since - Return the object only if it has not been modified since the specified time, else return a 412 (precondition failed).
func TestHeadOnObjectLastModified(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// preparing for object upload.
	objectName := "test-object"
	// content for the object to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))
	// obtaining URL for uploading the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// executing the HTTP request to download the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// make HTTP request to obtain object info.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// verify the status of the HTTP response.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// retrive the info of last modification time of the object from the response header.
	lastModified := response.Header.Get("Last-Modified")
	// Parse it into time.Time structure.
	lastTime, err := time.Parse(http.TimeFormat, lastModified)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// make HTTP request to obtain object info.
	// But this time set the "If-Modified-Since" header to be 10 minute more than the actual
	// last modified time of the object.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	request.Header.Set("If-Modified-Since", lastTime.Add(10*time.Minute).UTC().Format(http.TimeFormat))
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Since the "If-Modified-Since" header was ahead in time compared to the actual
	// modified time of the object expecting the response status to be http.StatusNotModified.
	if response.StatusCode != http.StatusNotModified {
		t.Errorf("Expected response status %s, got %s", http.StatusNotModified, response.StatusCode)
	}

	// Again, obtain the object info.
	// This time setting "If-Unmodified-Since" to a time after the object is modified.
	// As documented above, expecting http.StatusPreconditionFailed.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	request.Header.Set("If-Unmodified-Since", lastTime.Add(-10*time.Minute).UTC().Format(http.TimeFormat))
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("Expected response status %s, got %s", http.StatusPreconditionFailed, response.StatusCode)
	}
}

// TestHeadOnBucket - Validates response for HEAD on the bucket.
// HEAD request on the bucket validates the existence of the bucket.
func TestHeadOnBucket(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getHEADBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// make HEAD request on the bucket.
	request, err = newTestSignedRequest("HEAD", getHEADBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Asserting the response status for expected value of http.StatusOK.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
}

// TestContentTypePersists - Object upload with different Content-type is first done.
// And then a HEAD and GET request on these objects are done to validate if the same Content-Type set during upload persists.
func TestContentTypePersists(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// Uploading a new object with Content-Type "image/png".
	// content for the object to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))
	objectName := "test-object.png"
	// constructing HTTP request for object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	request.Header.Set("Content-Type", "image/png")
	if signerV4 == signerV2 {
		err = signRequestV2(request, accessKey, secretKey)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	client = &http.Client{}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// Fetching the object info using HEAD request for the object which was uploaded above.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Verify if the Content-Type header is set during the object persists.
	respContentType := response.Header.Get("Content-Type")
	expectedContentType := "image/png"

	if respContentType != expectedContentType {
		t.Errorf("Expected the response Content-Type to be `%s`, got `%s`", expectedContentType, respContentType)
	}

	// Fetching the object itself and then verify the Content-Type header.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// Execute the HTTP to fetch the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// Verify if the Content-Type header is set during the object persists.
	if respContentType != expectedContentType {
		t.Errorf("Expected the response Content-Type to be `%s`, got `%s`", expectedContentType, respContentType)
	}

	// Uploading a new object with Content-Type  "application/json".
	objectName = "test-object.json"
	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// setting the request header to be application/json.
	request.Header.Set("Content-Type", "application/json")
	if signerV4 == signerV2 {
		err = signRequestV2(request, accessKey, secretKey)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	// Execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// Obtain the info of the object which was uploaded above using HEAD request.
	request, err = newTestSignedRequest("HEAD", getHeadObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}

	respContentType = response.Header.Get("Content-Type")
	expectedContentType = "application/json"
	// Verify if the Content-Type header is set during the object persists.
	if respContentType != expectedContentType {
		t.Errorf("Expected the response Content-Type to be `%s`, got `%s`", expectedContentType, respContentType)
	}

	// Fetch the object and assert whether the Content-Type header persists.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	respContentType = response.Header.Get("Content-Type")
	// Verify if the Content-Type header is set during the object persists.
	if respContentType != expectedContentType {
		t.Errorf("Expected the response Content-Type to be `%s`, got `%s`", expectedContentType, respContentType)
	}

}

// TestPartialContent - Validating for GetObject with partial content request.
// By setting the Range header, A request to send specific bytes range of data from an
// already uploaded object can be done.
func TestPartialContent(t *testing.T) {
	bucketName := getRandomBucketName()

	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	buffer1 := bytes.NewReader([]byte("Hello World"))
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, "bar"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// Prepare request
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, "bar"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	request.Header.Add("Range", "bytes=6-7")

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusPartialContent {
		t.Errorf("Expected response status %s, got %s", http.StatusPartialContent, response.StatusCode)
	}
	partialObject, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if string(partialObject) != "Wo" {
		t.Errorf("Expected partial object content differs from the expected one.")
	}
}

// TestListObjectsHandler - Setting valid parameters to List Objects
// and then asserting the response with the expected one.
func TestListObjectsHandler(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	buffer1 := bytes.NewReader([]byte("Hello World"))
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, "bar"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// create listObjectsV1 request with valid parameters
	request, err = newTestSignedRequest("GET", getListObjectsV1URL(endPoint, bucketName, "1000"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	getContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(string(getContent), "<Key>bar</Key>") {
		t.Errorf("Invalid Get content.")
	}

	// create listObjectsV2 request with valid parameters
	request, err = newTestSignedRequest("GET", getListObjectsV2URL(endPoint, bucketName, "1000", ""),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	getContent, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !strings.Contains(string(getContent), "<Key>bar</Key>") {
		t.Errorf("Invalid content obtained from response body.")
	}

	if !strings.Contains(string(getContent), "<Owner><ID></ID><DisplayName></DisplayName></Owner>") {
		t.Errorf("Invalid content obtained from response body.")
	}

	// create listObjectsV2 request with valid parameters and fetch-owner activated
	request, err = newTestSignedRequest("GET", getListObjectsV2URL(endPoint, bucketName, "1000", "true"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	getContent, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !strings.Contains(string(getContent), "<Key>bar</Key>") {
		t.Errorf("Invalid content obtained from response body.")
	}

	if !strings.Contains(string(getContent), "<Owner><ID>minio</ID><DisplayName>minio</DisplayName></Owner>") {
		t.Errorf("Invalid content obtained from response body.")
	}
}

// TestListObjectsHandlerErrors - Setting invalid parameters to List Objects
// and then asserting the error response with the expected one.
func TestListObjectsHandlerErrors(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// create listObjectsV1 request with invalid value of max-keys parameter. max-keys is set to -2.
	request, err = newTestSignedRequest("GET", getListObjectsV1URL(endPoint, bucketName, "-2"),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// validating the error response.
	verifyError(t, response, "InvalidArgument", "Argument maxKeys must be an integer between 0 and 2147483647", http.StatusBadRequest)

	// create listObjectsV2 request with invalid value of max-keys parameter. max-keys is set to -2.
	request, err = newTestSignedRequest("GET", getListObjectsV2URL(endPoint, bucketName, "-2", ""),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// validating the error response.
	verifyError(t, response, "InvalidArgument", "Argument maxKeys must be an integer between 0 and 2147483647", http.StatusBadRequest)

}

// TestPutBucketErrors - request for non valid bucket operation
// and validate it with expected error result.
func TestPutBucketErrors(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// generating a HTTP request to create bucket.
	// using invalid bucket name.
	request, err := newTestSignedRequest("PUT", endPoint+"/putbucket-.",
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expected to fail with error message "InvalidBucketName".
	verifyError(t, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)
	// HTTP request to create the bucket.
	request, err = newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to create bucket.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// make HTTP request to create the same bucket again.
	// expected to fail with error message "BucketAlreadyOwnedByYou".
	request, err = newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "BucketAlreadyOwnedByYou", "Your previous request to create the named bucket succeeded and you already own it.",
		http.StatusConflict)

	// request for ACL.
	// Since Minio server doesn't support ACL's the request is expected to fail with  "NotImplemented" error message.
	request, err = newTestSignedRequest("PUT", endPoint+"/"+bucketName+"?acl",
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "NotImplemented", "A header you provided implies functionality that is not implemented", http.StatusNotImplemented)
}

func TestGetObjectLarge10MiB(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// form HTTP reqest to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create the bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,123"`
	// Create 10MiB content where each line contains 1024 characters.
	for i := 0; i < 10*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	putContent := buffer.String()

	buf := bytes.NewReader([]byte(putContent))

	objectName := "test-big-object"
	// create HTTP request for object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the status code to verify successful upload.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// prepare HTTP requests to download the object.
	request, err = newTestSignedRequest("GET", getPutObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to download the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// extract the content from response body.
	getContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Compare putContent and getContent.
	if string(getContent) != putContent {
		t.Errorf("Put and get content differ.")
	}
}

// TestGetObjectLarge11MiB - Tests validate fetching of an object of size 11MB.
func TestGetObjectLarge11MiB(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
	// Create 11MiB content where each line contains 1024 characters.
	for i := 0; i < 11*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	putMD5 := sumMD5(buffer.Bytes())

	objectName := "test-11Mb-object"
	// Put object
	buf := bytes.NewReader(buffer.Bytes())
	// create HTTP request foe object upload.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request for object upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// create HTTP request to download the object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// fetch the content from response body.
	getContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Get md5Sum of the response content.
	getMD5 := sumMD5(getContent)

	// Compare putContent and getContent.
	if hex.EncodeToString(putMD5) != hex.EncodeToString(getMD5) {
		t.Errorf("Get and Put content differ.")
	}
}

// TestGetPartialObjectMisAligned - tests get object partially mis-aligned.
// create a large buffer of mis-aligned data and upload it.
// then make partial range requests to while fetching it back and assert the response content.
func TestGetPartialObjectMisAligned(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create the bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`

	rand.Seed(time.Now().UTC().UnixNano())
	// Create a misalgined data.
	for i := 0; i < 13*rand.Intn(1<<16); i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line[:rand.Intn(1<<8)]))
	}
	putContent := buffer.String()
	buf := bytes.NewReader([]byte(putContent))

	objectName := "test-big-file"
	// HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// test Cases containing data to make partial range requests.
	// also has expected response data.
	var testCases = []struct {
		byteRange      string
		expectedString string
	}{
		// request for byte range 10-11.
		// expecting the result to contain only putContent[10:12] bytes.
		{"10-11", putContent[10:12]},
		// request for object data after the first byte.
		{"1-", putContent[1:]},
		// request for object data after the first byte.
		{"6-", putContent[6:]},
		// request for last 2 bytes of th object.
		{"-2", putContent[len(putContent)-2:]},
		// request for last 7 bytes of the object.
		{"-7", putContent[len(putContent)-7:]},
	}
	for _, testCase := range testCases {
		// HTTP request to download the object.
		request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
			0, nil, accessKey, secretKey, signerV4)
		if err != nil {
			t.Fatalf("%v", err)
		}
		// Get partial content based on the byte range set.
		request.Header.Add("Range", "bytes="+testCase.byteRange)

		client = &http.Client{}
		// execute the HTTP request.
		response, err = client.Do(request)
		if err != nil {
			t.Fatalf("%v", err)
		}
		// Since only part of the object is requested, expecting response status to be http.StatusPartialContent .
		// Assert the status code to verify successful upload.
		if response.StatusCode != http.StatusPartialContent {
			t.Errorf("Expected response status %s, got %s", http.StatusPartialContent, response.StatusCode)
		}
		// parse the HTTP response body.
		getContent, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("%v", err)
		}

		// Compare putContent and getContent.
		if string(getContent) != testCase.expectedString {
			t.Errorf("Get and Put content differ.")
		}
	}
}

// TestGetPartialObjectLarge11MiB - Test validates partial content request for a 11MiB object.
func TestGetPartialObjectLarge11MiB(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create the bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	line := `234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
	// Create 11MiB content where each line contains 1024
	// characters.
	for i := 0; i < 11*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	putContent := buffer.String()

	objectName := "test-large-11Mb-object"

	buf := bytes.NewReader([]byte(putContent))
	// HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// HTTP request to download the object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// This range spans into first two blocks.
	request.Header.Add("Range", "bytes=10485750-10485769")

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Since only part of the object is requested, expecting response status to be http.StatusPartialContent .
	if response.StatusCode != http.StatusPartialContent {
		t.Errorf("Expected response status %s, got %s", http.StatusPartialContent, response.StatusCode)
	}
	// read the downloaded content from the response body.
	getContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Compare putContent and getContent.
	if string(getContent) != putContent[10485750:10485770] {
		t.Errorf("Put and Get content doesn't match.")
	}
}

// TestGetPartialObjectLarge11MiB - Test validates partial content request for a 10MiB object.
func TestGetPartialObjectLarge10MiB(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	// expecting the error to be nil.
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expecting the HTTP response status code to 200 OK.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
	// Create 10MiB content where each line contains 1024 characters.
	for i := 0; i < 10*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	putContent := buffer.String()
	buf := bytes.NewReader([]byte(putContent))

	objectName := "test-big-10Mb-file"
	// HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buf.Len()), buf, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// verify whether upload was successful.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// HTTP request to download the object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Get partial content based on the byte range set.
	request.Header.Add("Range", "bytes=2048-2058")

	client = &http.Client{}
	// execute the HTTP request to download the partila content.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Since only part of the object is requested, expecting response status to be http.StatusPartialContent .
	// verify whether upload was successful.
	if response.StatusCode != http.StatusPartialContent {
		t.Errorf("Expected response status %s, got %s", http.StatusPartialContent, response.StatusCode)
	}

	// read the downloaded content from the response body.
	getContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Compare putContent and getContent.
	if string(getContent) != putContent[2048:2059] {
		t.Errorf("Get content doesn't match with the put content.")
	}
}

// TestGetObjectErrors - Tests validate error response for invalid object operations.
func TestGetObjectErrors(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()

	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-non-exitent-object"
	// HTTP request to download the object.
	// Since the specified object doesn't exist in the given bucket,
	// expected to fail with error message "NoSuchKey"
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	verifyError(t, response, "NoSuchKey", "The specified key does not exist.", http.StatusNotFound)

	// request to download an object, but an invalid bucket name is set.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, "getobjecterrors-.", objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expected to fail with "InvalidBucketName".
	verifyError(t, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)
}

// TestGetObjectRangeErrors - Validate error response when object is fetched with incorrect byte range value.
func TestGetObjectRangeErrors(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// content for the object to be uploaded.
	buffer1 := bytes.NewReader([]byte("Hello World"))

	objectName := "test-object"
	// HTTP request to upload the object.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the object.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// verify whether upload was successful.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// HTTP request to download the object.
	request, err = newTestSignedRequest("GET", getGetObjectURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	// Invalid byte range set.
	request.Header.Add("Range", "bytes=-0")
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expected to fail with "InvalidRange" error message.
	verifyError(t, response, "InvalidRange", "The requested range is not satisfiable", http.StatusRequestedRangeNotSatisfiable)
}

// TestObjectMultipartAbort - Test validates abortion of a multipart upload after uploading 2 parts.
func TestObjectMultipartAbort(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-multipart-object"

	// 1. Initiate 2 uploads for the same object
	// 2. Upload 2 parts for the second upload
	// 3. Abort the second upload.
	// 4. Abort the first upload.
	// This will test abort upload when there are more than one upload IDs
	// and the case where there is only one upload ID.

	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// parse the response body and obtain the new upload ID.
	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	}
	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// parse the response body and obtain the new upload ID.
	decoder = xml.NewDecoder(response.Body)
	newResponse = &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	}
	// uploadID to be used for rest of the multipart operations on the object.
	uploadID := newResponse.UploadID

	// content for the part to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "1"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to upload the first part.
	response1, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response1.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response1.StatusCode)
	}
	// content for the second part to be uploaded.
	buffer2 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the second part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "2"),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to upload the second part.
	response2, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response2.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response2.StatusCode)
	}
	// HTTP request for aborting the multipart upload.
	request, err = newTestSignedRequest("DELETE", getAbortMultipartUploadURL(endPoint, bucketName, objectName, uploadID),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to abort the multipart upload.
	response3, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expecting the response status code to be http.StatusNoContent.
	// The assertion validates the success of Abort Multipart operation.
	if response3.StatusCode != http.StatusNoContent {
		t.Errorf("Expected response status %s, got %s", http.StatusNoContent, response3.StatusCode)
	}
}

// TestBucketMultipartList - Initiates a NewMultipart upload, uploads parts and validates listing of the parts.
func TestBucketMultipartList(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName), 0,
		nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-multipart-object"
	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expecting the response status code to be http.StatusOK(200 OK) .
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// parse the response body and obtain the new upload ID.
	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	}
	// uploadID to be used for rest of the multipart operations on the object.
	uploadID := newResponse.UploadID

	// content for the part to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "1"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to upload the first part.
	response1, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response1.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response1.StatusCode)
	}

	// content for the second part to be uploaded.
	buffer2 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the second part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "2"),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to upload the second part.
	response2, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response2.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response2.StatusCode)
	}

	// HTTP request to ListMultipart Uploads.
	request, err = newTestSignedRequest("GET", getListMultipartURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response3, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response3.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response3.StatusCode)
	}

	// The reason to duplicate this structure here is to verify if the
	// unmarshalling works from a client perspective, specifically
	// while unmarshalling time.Time type for 'Initiated' field.
	// time.Time does not honor xml marshaler, it means that we need
	// to encode/format it before giving it to xml marshalling.

	// This below check adds client side verification to see if its
	// truly parseable.

	// listMultipartUploadsResponse - format for list multipart uploads response.
	type listMultipartUploadsResponse struct {
		XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListMultipartUploadsResult" json:"-"`

		Bucket             string
		KeyMarker          string
		UploadIDMarker     string `xml:"UploadIdMarker"`
		NextKeyMarker      string
		NextUploadIDMarker string `xml:"NextUploadIdMarker"`
		EncodingType       string
		MaxUploads         int
		IsTruncated        bool
		// All the in progress multipart uploads.
		Uploads []struct {
			Key          string
			UploadID     string `xml:"UploadId"`
			Initiator    Initiator
			Owner        Owner
			StorageClass string
			Initiated    time.Time // Keep this native to be able to parse properly.
		}
		Prefix         string
		Delimiter      string
		CommonPrefixes []CommonPrefix
	}

	// parse the response body.
	decoder = xml.NewDecoder(response3.Body)
	newResponse3 := &listMultipartUploadsResponse{}
	err = decoder.Decode(newResponse3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Assert the bucket name in the response with the expected bucketName.
	if newResponse3.Bucket != bucketName {
		t.Errorf("The bucket name is response doesn't match with expected bucket name.")
	}
	// Assert the IsTruncated field in the response with the expected bucketName.
	if newResponse3.IsTruncated != false {
		t.Errorf("IsTruncated field in the response doesn't match with the expected bucketName.")
	}
}

// TestValidateObjectMultipartUploadID - Test Initiates a new multipart upload and validates the uploadID.
func TestValidateObjectMultipartUploadID(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "directory1/directory2/object"
	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	// parse the response body and obtain the new upload ID.
	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}
	err = decoder.Decode(newResponse)
	// expecting the decoding error to be nil.
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Verifying for Upload ID value to be greater than 0.
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	}
}

// TestObjectMultipartListError - Initiates a NewMultipart upload, uploads parts and validates
// error response for an incorrect max-parts parameter .
func TestObjectMultipartListError(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-multipart-object"
	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// parse the response body and obtain the new upload ID.
	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	}
	// uploadID to be used for rest of the multipart operations on the object.
	uploadID := newResponse.UploadID

	// content for the part to be uploaded.
	buffer1 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "1"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request to upload the first part.
	response1, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response1.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response1.StatusCode)
	}

	// content for the second part to be uploaded.
	buffer2 := bytes.NewReader([]byte("hello world"))
	// HTTP request for the second part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "2"),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// execute the HTTP request to upload the second part.
	response2, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response2.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response2.StatusCode)
	}

	// HTTP request to ListMultipart Uploads.
	// max-keys is set to valid value of 1
	request, err = newTestSignedRequest("GET", getListMultipartURLWithParams(endPoint, bucketName, objectName, uploadID, "1", "", ""),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response3, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response3.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response3.StatusCode)
	}

	// HTTP request to ListMultipart Uploads.
	// max-keys is set to invalid value of -2.
	request, err = newTestSignedRequest("GET", getListMultipartURLWithParams(endPoint, bucketName, objectName, uploadID, "-2", "", ""),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// execute the HTTP request.
	response4, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Since max-keys parameter in the ListMultipart request set to invalid value of -2,
	// its expected to fail with error message "InvalidArgument".
	verifyError(t, response4, "InvalidArgument", "Argument max-parts must be an integer between 0 and 2147483647", http.StatusBadRequest)
}

// TestObjectValidMD5 - First uploads an object with a valid Content-Md5 header and verifies the status,
// then upload an object in a wrong Content-Md5 and validate the error response.
func TestObjectValidMD5(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// Create a byte array of 5MB.
	// content for the object to be uploaded.
	data := bytes.Repeat([]byte("0123456789abcdef"), 5*1024*1024/16)
	// calculate md5Sum of the data.
	hasher := md5.New()
	hasher.Write(data)
	md5Sum := hasher.Sum(nil)

	buffer1 := bytes.NewReader(data)
	objectName := "test-1-object"
	// HTTP request for the object to be uploaded.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// set the Content-Md5 to be the hash to content.
	request.Header.Set("Content-Md5", base64.StdEncoding.EncodeToString(md5Sum))
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expecting a successful upload.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	objectName = "test-2-object"
	buffer1 = bytes.NewReader(data)
	// HTTP request for the object to be uploaded.
	request, err = newTestSignedRequest("PUT", getPutObjectURL(endPoint, bucketName, objectName),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// set Content-Md5 to invalid value.
	request.Header.Set("Content-Md5", "kvLTlMrX9NpYDQlEIFlnDA==")
	// expecting a failure during upload.
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Since Content-Md5 header was wrong, expecting to fail with "SignatureDoesNotMatch" error.
	verifyError(t, response, "SignatureDoesNotMatch", "The request signature we calculated does not match the signature you provided. Check your key and signing method.", http.StatusForbidden)
}

// TestObjectMultipart - Initiates a NewMultipart upload, uploads 2 parts,
// completes the multipart upload and validates the status of the operation.
func TestObjectMultipart(t *testing.T) {
	// generate a random bucket name.
	bucketName := getRandomBucketName()
	// HTTP request to create the bucket.
	request, err := newTestSignedRequest("PUT", getMakeBucketURL(endPoint, bucketName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &http.Client{}
	// execute the HTTP request to create bucket.
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

	objectName := "test-multipart-object"
	// construct HTTP request to initiate a NewMultipart upload.
	request, err = newTestSignedRequest("POST", getNewMultipartURL(endPoint, bucketName, objectName),
		0, nil, accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request initiating the new multipart upload.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// expecting the response status code to be http.StatusOK(200 OK).
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}
	// parse the response body and obtain the new upload ID.
	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(newResponse.UploadID) <= 0 {
		t.Fatalf("Expected the length of the UploadID to be greater than 0.")
	} // uploadID to be used for rest of the multipart operations on the object.
	uploadID := newResponse.UploadID

	// content for the part to be uploaded.
	// Create a byte array of 5MB.
	data := bytes.Repeat([]byte("0123456789abcdef"), 5*1024*1024/16)
	// calculate md5Sum of the data.
	hasher := md5.New()
	hasher.Write(data)
	md5Sum := hasher.Sum(nil)

	buffer1 := bytes.NewReader(data)
	// HTTP request for the part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "1"),
		int64(buffer1.Len()), buffer1, accessKey, secretKey, signerV4)
	// set the Content-Md5 header to the base64 encoding the md5Sum of the content.
	request.Header.Set("Content-Md5", base64.StdEncoding.EncodeToString(md5Sum))
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the first part.
	response1, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response1.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response1.StatusCode)
	}

	// content for the second part to be uploaded.
	// Create a byte array of 1 byte.
	data = []byte("0")

	hasher = md5.New()
	hasher.Write(data)
	// calculate md5Sum of the data.
	md5Sum = hasher.Sum(nil)

	buffer2 := bytes.NewReader(data)
	// HTTP request for the second part to be uploaded.
	request, err = newTestSignedRequest("PUT", getPartUploadURL(endPoint, bucketName, objectName, uploadID, "2"),
		int64(buffer2.Len()), buffer2, accessKey, secretKey, signerV4)
	// set the Content-Md5 header to the base64 encoding the md5Sum of the content.
	request.Header.Set("Content-Md5", base64.StdEncoding.EncodeToString(md5Sum))
	if err != nil {
		t.Fatalf("%v", err)
	}

	client = &http.Client{}
	// execute the HTTP request to upload the second part.
	response2, err := client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if response2.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response2.StatusCode)
	}

	// Complete multipart upload
	completeUploads := &completeMultipartUpload{
		Parts: []completePart{
			{
				PartNumber: 1,
				ETag:       response1.Header.Get("ETag"),
			},
			{
				PartNumber: 2,
				ETag:       response2.Header.Get("ETag"),
			},
		},
	}

	completeBytes, err := xml.Marshal(completeUploads)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Indicating that all parts are uploaded and initiating completeMultipartUpload.
	request, err = newTestSignedRequest("POST", getCompleteMultipartUploadURL(endPoint, bucketName, objectName, uploadID),
		int64(len(completeBytes)), bytes.NewReader(completeBytes), accessKey, secretKey, signerV4)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Execute the complete multipart request.
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// verify whether complete multipart was successful.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response status %s, got %s", http.StatusOK, response.StatusCode)
	}

}
