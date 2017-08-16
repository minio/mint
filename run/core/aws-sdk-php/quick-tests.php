<?php
#
#  Mint, (C) 2017 Minio, Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software

#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

require 'vendor/autoload.php';

use Aws\S3\S3Client;
use Aws\Credentials;
use Aws\Exception\AwsException;
use GuzzleHttp\Psr7;

// Constants
const FILE_10KB = "datafile-10-kB";
const FILE_5_MB = "datafile-5-MB";
const HTTP_OK = "200";
const HTTP_NOCONTENT = "204";
const TEST_METADATA = ['Param-1' => 'val-1'];

/**
 * ClientConfig abstracts configuration details to connect to a
 * S3-like service
 */
class ClientConfig {
    public $creds;
    public $endpoint;

    function __construct(string $access_key, string $secret_key, string $host, string $secure) {
        $this->creds = new Aws\Credentials\Credentials($access_key, $secret_key);

        if ($secure == "1")  {
            $this->endpoint = "https://" . $host;
        } else {
            $this->endpoint = "http://" . $host;
        }
    }
}

 /**
  * randomName returns a string of random characters picked from 0-9,a-z.
  *
  * By default it returns a random name of length 5.
  *
  * @param int $length - length of random name.
  *
  * @return void
  */
function randomName(int $length=5):string {
    $alphabet = array_rand(str_split('0123456789abcdefghijklmnopqrstuvwxyz'), 26);
    $alphaLen = count($alphabet);
    $rounds = floor($length / $alphaLen);
    $remaining = $length % $alphaLen;
    $bigalphabet = [];
    for ($i = 0; $i < $rounds; $i++) {
        $bigalphabet = array_merge($bigalphabet, $alphabet);
    }
    $bigalphabet = array_merge($bigalphabet, array_slice($alphabet, 0, $remaining));
    $alphabet_soup = array_flip($bigalphabet);
    shuffle($alphabet_soup);
    return 'aws-sdk-php-bucket-' . join($alphabet_soup);
}

 /**
  * getStatusCode returns HTTP status code of the given result.
  *
  * @param $result - AWS\S3 result object
  *
  * @return string - HTTP status code. E.g, "400" for Bad Request.
  */
function getStatusCode($result):string {
    return $result->toArray()['@metadata']['statusCode'];
}

 /**
  * runExceptionalTests executes a collection of tests that will throw
  * a known exception.
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $apiCall Name of the S3Client API method to call
  *
  * @param $exceptionMatcher Name of Aws\S3\Exception\S3Exception
  * method to fetch exception details
  *
  * @param $exceptionParamMap Associative array of exception names to
  * API parameters. E.g,
  * $apiCall = 'headBucket'
  * $exceptionMatcher = 'getStatusCode'
  * $exceptionParamMap = [
  * // Invalid bucket name
  *     '400' => ['Bucket' => $bucket['Name'] . '--'],
  *
  *      // Non existent bucket
  *      '404' => ['Bucket' => $bucket['Name'] . '-non-existent'],
  * ];
  *
  * @return string - HTTP status code. E.g, "400" for Bad Request.
  */
function runExceptionalTests($s3Client, $apiCall, $exceptionMatcher, $exceptionParamMap) {
    foreach($exceptionParamMap as $exn => $params) {
        $exceptionCaught = false;
        try {
            $result = $s3Client->$apiCall($params);
        } catch(Aws\S3\Exception\S3Exception $e) {
            $exceptionCaught = true;
            switch ($e->$exceptionMatcher()) {
            case $exn:
                // This is expected
                continue;
            default:
                throw $e;
            }
        }
        finally {
            if (!$exceptionCaught) {
                $message = sprintf("Expected %s to fail with %s", $apiCall, $exn);
                throw new Exception($message);
            }
        }
    }
}

 /**
  * testListBuckets tests ListBuckets S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @return void
  */
function testListBuckets(S3Client $s3Client) {
    $buckets = $s3Client->listBuckets();
    foreach ($buckets['Buckets'] as $bucket){
        echo $bucket['Name']."\n";
    }
}

 /**
  * testBucketExists tests HEAD Bucket S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @return void
  */
function testBucketExists(S3Client $s3Client) {
    // List all buckets
    $buckets = $s3Client->listBuckets();
    // All HEAD on existing buckets must return success
    foreach($buckets['Buckets'] as $bucket) {
        $result = $s3Client->headBucket(['Bucket' => $bucket['Name']]);
        if (getStatusCode($result) != HTTP_OK)
            throw new Exception('headBucket API failed for ' . $bucket['Name']);
    }

    // Run failure tests
    $params = [
        // Invalid bucket name
        '400' => ['Bucket' => $bucket['Name'] . '--'],

        // Non existent bucket
        '404' => ['Bucket' => $bucket['Name'] . '-non-existent'],
    ];
    runExceptionalTests($s3Client, 'headBucket', 'getStatusCode', $params);
}


 /**
  * testHeadObject tests HeadObject S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $objects Associative array of buckets and objects
  *
  * @return void
  */
function testHeadObject($s3Client, $objects) {
    foreach($objects as $bucket => $object) {
        $result = $s3Client->headObject(['Bucket' => $bucket, 'Key' => $object]);
        if (getStatusCode($result) != HTTP_OK)
            throw new Exception('headObject API failed for ' .
                                $bucket . '/' . $object);
        if ($result['Metadata'] != TEST_METADATA) {
            throw new Exception("headObject API Metadata didn't match for " .
                                $bucket . '/' . $object);
        }
    }

    // Run failure tests
    $params = [
        '404' => ['Bucket' => $bucket, 'Key' => $object . '-non-existent']
    ];
    runExceptionalTests($s3Client, 'headObject', 'getStatusCode', $params);
}

 /**
  * testListObjects tests ListObjectsV1 and V2 S3 APIs
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $objects Associative array of buckets and objects
  *
  * @return void
  */
function testListObjects($s3Client, $bucket, $object) {
    try {
        for ($i = 0; $i < 5; $i++) {
            $copyKey = $object . '-copy-' . strval($i);
            $result = $s3Client->copyObject([
                'Bucket' => $bucket,
                'Key' => $copyKey,
                'CopySource' => $bucket . '/' . $object,
            ]);
            if (getStatusCode($result) != HTTP_OK)
                throw new Exception("copyObject API failed for " . $bucket . '/' . $object);
        }

        $paginator = $s3Client->getPaginator('ListObjects', ['Bucket' => $bucket]);
        foreach ($paginator->search('Contents[].Key') as $key) {
            echo 'key = ' . $key . "\n";
        }

        $paginator = $s3Client->getPaginator('ListObjectsV2', ['Bucket' => $bucket]);
        foreach ($paginator->search('Contents[].Key') as $key) {
            echo 'key = ' . $key . "\n";
        }

        $prefix = 'obj';
        $result = $s3Client->listObjects(['Bucket' => $bucket, 'Prefix' => $prefix]);
        if (getStatusCode($result) != HTTP_OK || $result['Prefix'] != $prefix)
            throw new Exception("listObject API failed for " . $bucket . '/' . $object);

        $maxKeys = 1;
        $result = $s3Client->listObjects(['Bucket' => $bucket, 'MaxKeys' => $maxKeys]);
        if (getStatusCode($result) != HTTP_OK || count($result['Contents']) != $maxKeys)
            throw new Exception("listObject API failed for " . $bucket . '/' . $object);

        $params = [
            'InvalidArgument' => ['Bucket' => $bucket, 'MaxKeys' => -1],
            'NoSuchBucket' => ['Bucket' => $bucket . '-non-existent']
        ];
        runExceptionalTests($s3Client, 'listObjects', 'getAwsErrorCode', $params);

    } finally {
        $s3Client->deleteObjects([
            'Bucket' => $bucket,
            'Delete' => [
                'Objects' => array_map(function($a, $b) {
                    return ['Key' =>  $a . '-copy-' . strval($b)];
                }, array_fill(0, 5, $object), range(0,4))
            ],
        ]);
    }
}

 /**
  * testListMultipartUploads tests ListMultipartUploads, ListParts and
  * UploadPartCopy S3 APIs
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket Name of bucket
  *
  * @param $object Name of object to be copied
  *
  * @return void
  */
function testListMultipartUploads($s3Client, $bucket, $object) {
    $data_dir = $GLOBALS['MINT_DATA_DIR'];
    // Initiate multipart upload
    $result = $s3Client->createMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object . '-copy',
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('createMultipartupload API failed for ' .
                            $bucket . '/' . $object);

    // upload 5 parts
    $uploadId = $result['UploadId'];
    $parts = [];
    try {
        for ($i = 0; $i < 5; $i++) {
            $result = $s3Client->uploadPartCopy([
                'Bucket' => $bucket,
                'Key' => $object . '-copy',
                'UploadId' => $uploadId,
                'PartNumber' => $i+1,
                'CopySource' => $bucket . '/' . $object,
            ]);
            if (getStatusCode($result) != HTTP_OK) {
                throw new Exception('uploadPart API failed for ' .
                                    $bucket . '/' . $object);
            }
            array_push($parts, [
                'ETag' => $result['ETag'],
                'PartNumber' => $i+1,
            ]);
        }

        // ListMultipartUploads and ListParts may return empty
        // responses in the case of minio gateway gcs and minio server
        // FS mode. So, the following tests don't make assumptions on
        // result response.
        $paginator = $s3Client->getPaginator('ListMultipartUploads',
                                             ['Bucket' => $bucket]);
        foreach ($paginator->search('Uploads[].{Key: Key, UploadId: UploadId}') as $keyHash) {
            echo 'key = ' . $keyHash['Key'] . ' uploadId = ' . $keyHash['UploadId'] . "\n";
        }

        $paginator = $s3Client->getPaginator('ListParts', [
            'Bucket' => $bucket,
            'Key' => $object . '-copy',
            'UploadId' => $uploadId,
        ]);
        foreach ($paginator->search('Parts[].{PartNumber: PartNumber, ETag: ETag}') as $partsHash) {
            echo 'partNumber = ' . $partsHash['PartNumber'] . ' ETag = ' . $partsHash['ETag'] . "\n";
        }

    } finally {
        $s3Client->abortMultipartUpload([
            'Bucket' => $bucket,
            'Key' => $object . '-copy',
            'UploadId' => $uploadId
        ]);
    }
}

 /**
  * initSetup creates buckets and objects necessary for the functional
  * tests to run
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $objects Associative array of buckets and objects
  *
  * @return void
  */
function initSetup(S3Client $s3Client, $objects) {
    $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
    foreach($objects as $bucket => $object) {
        $s3Client->createBucket(['Bucket' => $bucket]);
        $stream = NULL;
        try {
            if (!file_exists($MINT_DATA_DIR . '/' . FILE_10KB))
                throw new Exception('File not found ' . $MINT_DATA_DIR . '/' . FILE_10KB);

            $stream = Psr7\stream_for(fopen($MINT_DATA_DIR . '/' . FILE_10KB, 'r'));
            $result = $s3Client->putObject([
                'Bucket' => $bucket,
                'Key' => $object,
                'Body' => $stream,
                'Metadata' => TEST_METADATA,
            ]);
            if (getStatusCode($result) != HTTP_OK)
                throw new Exception("putObject API failed for " . $bucket . '/' . $object);
        }

        finally {
            // close data file
            if (!is_null($stream))
                $stream->close();
        }
    }

    // Create an empty bucket for bucket policy + delete tests
    $result = $s3Client->createBucket(['Bucket' => $GLOBALS['emptyBucket']]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception("createBucket API failed for " . $bucket);

}


 /**
  * testGetPutObject tests GET/PUT object S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket on which GET/PUT operations are performed
  *
  * @param $object object to be downloaded/uploaded
  *
  * @return void
  */
function testGetPutObject($s3Client, $bucket, $object) {
    // Upload a 10KB file
    $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
    try {
        $stream = Psr7\stream_for(fopen($MINT_DATA_DIR . '/' . FILE_10KB, 'r'));
        $result = $s3Client->putObject([
            'Bucket' => $bucket,
            'Key' => $object,
            'Body' => $stream,
        ]);
    }
    finally {
        $stream->close();
    }

    if (getStatusCode($result) != HTTP_OK)
        throw new Exception("putObject API failed for " . $bucket . '/' . $object);

    // Download the same object and verify size
    $result = $s3Client->getObject([
        'Bucket' => $bucket,
        'Key' => $object,
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception("getObject API failed for " . $bucket . '/' . $object);

    $body = $result['Body'];
    $bodyLen = 0;
    while (!$body->eof()) {
        $bodyLen += strlen($body->read(4096));
    }

    if ($bodyLen != 10 * 1024) {
        throw new Exception("Object downloaded has different content length than uploaded object "
                            . $bucket . '/' . $object);
    }
}

 /**
  * testMultipartUploadFailure tests MultipartUpload failures
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket on objects are uploaded using
  * MultipartUpload API
  *
  * @param $object object to be uploaded
  *
  * @return void
  */
function testMultipartUploadFailure($s3Client, $bucket, $object) {
    $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
    // Initiate multipart upload
    $result = $s3Client->createMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object,
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('createMultipartupload API failed for ' .
                            $bucket . '/' . $object);

    // upload 2 parts
    $uploadId = $result['UploadId'];
    $parts = [];
    try {
        for ($i = 0; $i < 2; $i++) {
            $stream = Psr7\stream_for(fopen($MINT_DATA_DIR . '/' . FILE_5_MB, 'r'));
            $limitedStream = new Psr7\LimitStream($stream, 4 * 1024 * 1024, 0);
            $result = $s3Client->uploadPart([
                'Bucket' => $bucket,
                'Key' => $object,
                'UploadId' => $uploadId,
                'ContentLength' => 4 * 1024 * 1024,
                'Body' => $limitedStream,
                'PartNumber' => $i+1,
            ]);
            if (getStatusCode($result) != HTTP_OK) {
                throw new Exception('uploadPart API failed for ' .
                                    $bucket . '/' . $object);
            }
            array_push($parts, [
                'ETag' => $result['ETag'],
                'PartNumber' => $i+1,
            ]);

            $limitedStream->close();
            $limitedStream = NULL;
        }
    }
    finally {
        if (!is_null($limitedStream))
            $limitedStream->close();
    }

    $params = [
        'EntityTooSmall' => [
            'Bucket' => $bucket,
            'Key' => $object,
            'UploadId' => $uploadId,
            'MultipartUpload' => [
                'Parts' => $parts,
            ],
        ],
        'NoSuchUpload' => [
            'Bucket' => $bucket,
            'Key' => $object,
            'UploadId' => 'non-existent',
            'MultipartUpload' => [
                'Parts' => $parts,
            ],
        ],
    ];
    runExceptionalTests($s3Client, 'completeMultipartUpload', 'getAwsErrorCode', $params);
}

 /**
  * testMultipartUpload tests MultipartUpload S3 APIs
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket on objects are uploaded using
  * MultipartUpload API
  *
  * @param $object object to be uploaded
  *
  * @return void
  */
function testMultipartUpload($s3Client, $bucket, $object) {
    $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
    // Initiate multipart upload
    $result = $s3Client->createMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object,
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('createMultipartupload API failed for ' .
                            $bucket . '/' . $object);

    // upload 2 parts
    $uploadId = $result['UploadId'];
    $parts = [];
    try {
        for ($i = 0; $i < 2; $i++) {
            $stream = Psr7\stream_for(fopen($MINT_DATA_DIR . '/' . FILE_5_MB, 'r'));
            $result = $s3Client->uploadPart([
                'Bucket' => $bucket,
                'Key' => $object,
                'UploadId' => $uploadId,
                'ContentLength' => 5 * 1024 * 1024,
                'Body' => $stream,
                'PartNumber' => $i+1,
            ]);
            if (getStatusCode($result) != HTTP_OK) {
                throw new Exception('uploadPart API failed for ' .
                                    $bucket . '/' . $object);
            }
            array_push($parts, [
                'ETag' => $result['ETag'],
                'PartNumber' => $i+1,
            ]);

            $stream->close();
            $stream = NULL;
        }
    }
    finally {
        if (!is_null($stream))
            $stream->close();
    }

    // complete multipart upload
    $result = $s3Client->completeMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object,
        'UploadId' => $uploadId,
        'MultipartUpload' => [
            'Parts' => $parts,
        ],
    ]);
    if (getStatusCode($result) != HTTP_OK) {
        throw new Exception('completeMultipartupload API failed for ' .
                            $bucket . '/' . $object);
    }
}

 /**
  * testAbortMultipartUpload tests aborting of a multipart upload
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket on objects are uploaded using
  * MultipartUpload API
  *
  * @param $object object to be uploaded
  *
  * @return void
  */
function testAbortMultipartUpload($s3Client, $bucket, $object) {
    $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
    // Initiate multipart upload
    $result = $s3Client->createMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object,
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('createMultipartupload API failed for ' .
                            $bucket . '/' . $object);

    // Abort multipart upload
    $uploadId = $result['UploadId'];
    $result = $s3Client->abortMultipartUpload([
        'Bucket' => $bucket,
        'Key' => $object,
        'UploadId' => $uploadId,
    ]);
    if (getStatusCode($result) != HTTP_NOCONTENT)
        throw new Exception('abortMultipartupload API failed for ' .
                            $bucket . '/' . $object);

    //Run failure tests
    $params = [
        // Upload doesn't exist
        'NoSuchUpload' => [
            'Bucket' => $bucket,
            'Key' => $object,
            'UploadId' => 'non-existent',
        ],
    ];
    runExceptionalTests($s3Client, 'abortMultipartUpload', 'getAwsErrorCode', $params);
}

 /**
  * testGetBucketLocation tests GET bucket location S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket whose location is to be determined
  *
  * @return void
  */
function testGetBucketLocation($s3Client, $bucket) {
    // Valid test
    $result = $s3Client->getBucketLocation(['Bucket' => $bucket]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('getBucketLocation API failed for ' .
                            $bucket);

    // Run failure tests.
    $params = [
        // InvalidBucketName test
        'InvalidBucketName' => ['Bucket' => $bucket . '--'],

        // Bucket not found
        'NoSuchBucket' => ['Bucket' => $bucket . '-non-existent'],
    ];
    runExceptionalTests($s3Client, 'getBucketLocation', 'getAwsErrorCode', $params);
}

 /**
  * testCopyObject tests copy object S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket from where object is copied from and copied
  * into
  *
  * @param $object object to be copied
  *
  * @return void
  */
function testCopyObject($s3Client, $bucket, $object) {
    $result = $s3Client->copyObject([
        'Bucket' => $bucket,
        'Key' => $object . '-copy',
        'CopySource' => $bucket . '/' . $object,
    ]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('copyObject API failed for ' .
                            $bucket);

    $s3Client->deleteObject([
        'Bucket' => $bucket,
        'Key' => $object . '-copy',
    ]);

    // Run failure tests
    $params = [
        // Invalid copy source format
        'InvalidArgument' => [
            'Bucket' => $bucket,
            'Key' => $object . '-copy',
            'CopySource' => $bucket . $object
        ],

        // Missing source object
        'NoSuchKey' => [
            'Bucket' => $bucket,
            'Key' => $object . '-copy',
            'CopySource' => $bucket . '/' . $object . '-non-existent'
        ],
    ];
    runExceptionalTests($s3Client, 'copyObject', 'getAwsErrorCode', $params);
}

 /**
  * testDeleteObjects tests Delete Objects S3 API
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket from where objects are deleted
  * into
  *
  * @param $object whose copies are deleted in one batch
  *
  * @return void
  */
function testDeleteObjects($s3Client, $bucket, $object) {
    $copies = [];
    for ($i = 0; $i < 3; $i++) {
        $copyKey = $object . '-copy' . strval($i);
        $result = $s3Client->copyObject([
            'Bucket' => $bucket,
            'Key' => $copyKey,
            'CopySource' => $bucket . '/' . $object,
        ]);
        if (getstatuscode($result) != HTTP_OK)
            throw new Exception('copyobject API failed for ' .
                                $bucket);
        array_push($copies, ['Key' => $copyKey]);
    }

    $result = $s3Client->deleteObjects([
        'Bucket' => $bucket,
        'Delete' => [
            'Objects' => $copies,
        ],
    ]);
    if (getstatuscode($result) != HTTP_OK)
        throw new Exception('deleteObjects api failed for ' .
                            $bucket);
}

 /**
  * testAnonDeleteObjects tests Delete Objects S3 API for anonymous requests. 
  * The test case checks this scenario: 
  * http://docs.aws.amazon.com/AmazonS3/latest/API/multiobjectdeleteapi.html#multiobjectdeleteapi-examples
  *
  * @param $anonymousClient Anonymous AWS\S3\S3Client object
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket from where objects are deleted from
  *
  * @param $object whose copies are deleted in one batch
  *
  * @return void
  */
function testAnonDeleteObjects($anonymousClient, $s3Client, $bucket, $object) {
    $copies = [];
    for ($i = 0; $i < 3; $i++) {
        $copyKey = $object . '-copy' . strval($i);
        $result = $s3Client->copyObject([
            'Bucket' => $bucket,
            'Key' => $copyKey,
            'CopySource' => $bucket . '/' . $object,
        ]);
        if (getstatuscode($result) != HTTP_OK)
            throw new Exception('copyobject API failed for ' .
                                $bucket);
        array_push($copies, ['Key' => $copyKey]);
    }

    // Try anonymous delete. 
    $result = $anonymousClient->deleteObjects([
        'Bucket' => $bucket,
        'Delete' => [
            'Objects' => $copies,
        ],
    ]);
    // Response code should be 200
    if (getstatuscode($result) != HTTP_OK)
        throw new Exception('deleteObjects returned incorrect response ' .
            getStatusCode($result));

    // Each object should have error code AccessDenied
    for ($i = 0; $i < 3; $i++) {
        if ($result["Errors"][$i]["Code"] != "AccessDenied") 
            throw new Exception('Incorrect response deleteObjects anonymous 
                                call for ' .$bucket);
    }

    // Delete objects after the test passed
    $result = $s3Client->deleteObjects([
        'Bucket' => $bucket,
        'Delete' => [
            'Objects' => $copies,
        ],
    ]);

    if (getstatuscode($result) != HTTP_OK)
        throw new Exception('deleteObjects api failed for ' .
                            $bucket);

    // Each object should have empty code in case of successful delete
    for ($i = 0; $i < 3; $i++) {
        if ($result["Errors"][$i]["Code"] != "") 
            throw new Exception('Incorrect response deleteObjects anonymous 
                                call for ' .$bucket);
    }
}

 /**
  *  testBucketPolicy tests GET/PUT Bucket policy S3 APIs
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $bucket bucket on which policy is being set
  *
  * @return void
  */
function testBucketPolicy($s3Client, $bucket) {
    // Taken from policy set using `mc policy download`
    $downloadPolicy = sprintf('{"Version":"2012-10-17","Statement":[{"Action":["s3:GetBucketLocation","s3:ListBucket"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s"],"Sid":""},{"Action":["s3:GetObject"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s/*"],"Sid":""}]}', $bucket, $bucket);

    $result = $s3Client->putBucketPolicy([
        'Bucket' => $bucket,
        'Policy' => $downloadPolicy
    ]);
    if (getstatuscode($result) != HTTP_NOCONTENT)
        throw new Exception('putBucketPolicy API failed for ' .
                            $bucket);
    $result = $s3Client->getBucketPolicy(['Bucket' => $bucket]);
    if (getstatuscode($result) != HTTP_OK)
        throw new Exception('putBucketPolicy API failed for ' .
                            $bucket);

    if ($result['Policy'] != $downloadPolicy)
        throw new Exception('bucket policy we got is not we set');

    // Delete the bucket, make the bucket (again) and check if policy is none
    // Ref: https://github.com/minio/minio/issues/4714
    $result = $s3Client->deleteBucket(['Bucket' => $bucket]);
    if (getstatuscode($result) != HTTP_NOCONTENT)
        throw new Exception('deleteBucket API failed for ' .
                            $bucket);

    $result = $s3Client->createBucket(['Bucket' => $bucket]);
    if (getstatuscode($result) != HTTP_OK)
        throw new Exception('createBucket API failed for ' .
                            $bucket);

    $params = [
        '404' => ['Bucket' => $bucket]
    ];
    runExceptionalTests($s3Client, 'getBucketPolicy', 'getStatusCode', $params);

    try {
        $MINT_DATA_DIR = $GLOBALS['MINT_DATA_DIR'];
        // Create an object to test anonymous GET object
        $object = 'test-anon';
        if (!file_exists($MINT_DATA_DIR . '/' . FILE_10KB))
            throw new Exception('File not found ' . $MINT_DATA_DIR . '/' . FILE_10KB);

        $stream = Psr7\stream_for(fopen($MINT_DATA_DIR . '/' . FILE_10KB, 'r'));
        $result = $s3Client->putObject([
                'Bucket' => $bucket,
                'Key' => $object,
                'Body' => $stream,
        ]);
        if (getstatuscode($result) != HTTP_OK)
            throw new Exception('createBucket API failed for ' .
                                $bucket);

        $anonConfig = new ClientConfig("", "", $GLOBALS['endpoint'], $GLOBALS['secure']);
        $anonymousClient = new S3Client([
            'credentials' => false,
            'endpoint' => $anonConfig->endpoint,
            'use_path_style_endpoint' => true,
            'region' => 'us-east-1',
            'version' => '2006-03-01'
        ]);
        runExceptionalTests($anonymousClient, 'getObject', 'getStatusCode', [
            '403' => [
                'Bucket' => $bucket,
                'Key' => $object,
            ]
        ]);

    } finally {
        // close data file
        if (!is_null($stream))
            $stream->close();
        $s3Client->deleteObject(['Bucket' => $bucket, 'Key' => $object]);
    }

}

 /**
  * cleanupSetup removes all buckets and objects created during the
  * functional test
  *
  * @param $s3Client AWS\S3\S3Client object
  *
  * @param $objects Associative array of buckets to objects
  *
  * @return void
  */
function cleanupSetup($s3Client, $objects) {
    // Delete all objects
    foreach ($objects as $bucket => $object) {
        $s3Client->deleteObject(['Bucket' => $bucket, 'Key' => $object]);
    }

    // Delete the buckets incl. emptyBucket
    $allBuckets = array_keys($objects);
    array_push($allBuckets, $GLOBALS['emptyBucket']);
    foreach ($allBuckets as $bucket) {
        // Delete the bucket
        $s3Client->deleteBucket(['Bucket' => $bucket]);

        // Wait until the bucket is removed from object store
        $s3Client->waitUntil('BucketNotExists', ['Bucket' => $bucket]);
    }
}


 /**
  * runTestFunction helper function to wrap a test function and log
  * success or failure accordingly.
  *
  * @param myfunc name of test function to be run
  *
  * @param args parameters to be passed to test function
  *
  * @return void
  */
function runTestFunction($myfunc, ...$args) {
    static $counter = 1;

    printf("%d. Testing %s\n\n", $counter, $myfunc);
    try {
        $myfunc(...$args);
    } catch (Exception $e) {
        printf("%d. %s failed at %d %s", $counter, $myfunc, $e->getLine(), $e->getFile());
        throw $e;
    } finally {
        $counter++;
    }
    echo "\nPASSED " . $myfunc . "\n\n";
}

// Get client configuration from environment variables
$GLOBALS['access_key'] = getenv("ACCESS_KEY");
$GLOBALS['secret_key'] = getenv("SECRET_KEY");
$GLOBALS['endpoint'] = getenv("SERVER_ENDPOINT");
$GLOBALS['secure'] = getenv("ENABLE_HTTPS");

/**
 * @global string $GLOBALS['MINT_DATA_DIR']
 * @name $MINT_DATA_DIR
 */
$GLOBALS['MINT_DATA_DIR'] = '/mint/data';
$GLOBALS['MINT_DATA_DIR'] = getenv("MINT_DATA_DIR");


// Create config object
$config = new ClientConfig($GLOBALS['access_key'], $GLOBALS['secret_key'],
                           $GLOBALS['endpoint'], $GLOBALS['secure']);

// Create a S3Client
$s3Client = new S3Client([
    'credentials' => $config->creds,
    'endpoint' => $config->endpoint,
    'use_path_style_endpoint' => true,
    'region' => 'us-east-1',
    'version' => '2006-03-01'
]);

// Create anonymous config object
$anonConfig = new ClientConfig("", "", $GLOBALS['endpoint'], $GLOBALS['secure']);

// Create anonymous S3 client
$anonymousClient = new S3Client([
    'credentials' => false,
    'endpoint' => $anonConfig->endpoint,
    'use_path_style_endpoint' => true,
    'region' => 'us-east-1',
    'version' => '2006-03-01'
]);

// Used by initSetup
$emptyBucket = randomName();
$objects =  [
    randomName() => 'obj1',
    randomName() => 'obj2',
];

try {
    initSetup($s3Client, $objects);
    $firstBucket = array_keys($objects)[0];
    $firstObject = $objects[$firstBucket];
    runTestFunction('testGetBucketLocation', $s3Client, $firstBucket);
    runTestFunction('testListBuckets', $s3Client);
    runTestFunction('testListObjects', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testListMultipartUploads', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testBucketExists', $s3Client, array_keys($objects));
    runTestFunction('testHeadObject', $s3Client, $objects);
    runTestFunction('testGetPutObject', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testCopyObject', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testDeleteObjects', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testAnonDeleteObjects', $anonymousClient, $s3Client, $firstBucket, $firstObject);
    runTestFunction('testMultipartUpload', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testMultipartUploadFailure', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testAbortMultipartUpload', $s3Client, $firstBucket, $firstObject);
    runTestFunction('testBucketPolicy', $s3Client, $emptyBucket);
}
finally {
    cleanupSetup($s3Client, $objects);
}

?>
