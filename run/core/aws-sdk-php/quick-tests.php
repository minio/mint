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
    $data_dir = $GLOBALS['data_dir'];
    foreach($objects as $bucket => $object) {
        $s3Client->createBucket(['Bucket' => $bucket]);
        try {
            if (!file_exists($data_dir . '/' . FILE_10KB))
                throw new Exception('File not found ' . $data_dir . '/' . FILE_10KB);

            $stream = Psr7\stream_for(fopen($data_dir . '/' . FILE_10KB, 'r'));
            $result = $s3Client->putObject([
                'Bucket' => $bucket,
                'Key' => $object,
                'Body' => $stream,
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
    $data_dir = $GLOBALS['data_dir'];
    try {
        $stream = Psr7\stream_for(fopen($data_dir . '/' . FILE_10KB, 'r'));
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
    $data_dir = $GLOBALS['data_dir'];
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
            $stream = Psr7\stream_for(fopen($data_dir . '/' . FILE_5_MB, 'r'));
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
    $data_dir = $GLOBALS['data_dir'];
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
            $stream = Psr7\stream_for(fopen($data_dir . '/' . FILE_5_MB, 'r'));
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
    $data_dir = $GLOBALS['data_dir'];
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
            throw new exception('copyobject api failed for ' .
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
        throw new exception('deleteObjects api failed for ' .
                            $bucket);
}

 /**
  * cleanpupSetup removes all buckets and objects created during the
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

    // Delete the buckets
    foreach (array_keys($objects) as $bucket) {
        // Delete the bucket
        $s3Client->deleteBucket(['Bucket' => $bucket]);

        // Wait until the bucket is removed from object store
        $s3Client->waitUntil('BucketNotExists', ['Bucket' => $bucket]);
    }
}

// Get client configuration from environment variables
$access_key = getenv("ACCESS_KEY");
$secret_key = getenv("SECRET_KEY");
$endpoint = getenv("SERVER_ENDPOINT");
$secure = getenv("ENABLE_HTTPS");

/**
 * @global string $GLOBALS['data_dir']
 * @name $data_dir
 */
$GLOBALS['data_dir'] = '/mint/data';
$GLOBALS['data_dir'] = getenv("DATA_DIR");


// Create config object
$config = new ClientConfig($access_key, $secret_key, $endpoint, $secure);


// Create a S3Client
$s3Client = new S3Client([
    'credentials' => $config->creds,
    'endpoint' => $config->endpoint,
    'use_path_style_endpoint' => true,
    'region' => 'us-east-1',
    'version' => '2006-03-01'
]);

$objects =  [
    randomName() => 'obj1',
    randomName() => 'obj2',
];

try {
    initSetup($s3Client, $objects);
    $firstBucket = array_keys($objects)[0];
    $firstObject = $objects[$firstBucket];
    testGetBucketLocation($s3Client, $firstBucket);
    testListBuckets($s3Client);
    testBucketExists($s3Client, array_keys($objects));
    testGetPutObject($s3Client, $firstBucket, $firstObject);
    testCopyObject($s3Client, $firstBucket, $firstObject);
    testDeleteObjects($s3Client, $firstBucket, $firstObject);
    testMultipartUpload($s3Client, $firstBucket, $firstObject);
    testMultipartUploadFailure($s3Client, $firstBucket, $firstObject);
    testAbortMultipartUpload($s3Client, $firstBucket, $firstObject);
}
finally {
    cleanupSetup($s3Client, $objects);
}

?>
