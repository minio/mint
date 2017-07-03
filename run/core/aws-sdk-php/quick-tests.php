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

function getStatusCode($result):string {
    return $result->toArray()['@metadata']['statusCode'];
}

// Test for Listing all S3 Bucket
function testListBuckets(S3Client $s3Client) {
    $buckets = $s3Client->listBuckets();
    foreach ($buckets['Buckets'] as $bucket){
        echo $bucket['Name']."\n";
    }
}

// Test for createBucket, headBucket
function testBucketExists(S3Client $s3Client) {
    // List all buckets
    $buckets = $s3Client->listBuckets();
    // All HEAD on existing buckets must return success
    foreach($buckets['Buckets'] as $bucket) {
        $result = $s3Client->headBucket(['Bucket' => $bucket['Name']]);
        if (getStatusCode($result) != HTTP_OK)
            throw new Exception('headBucket API failed for ' . $bucket['Name']);
    }
}

// initializes setup with creating $objects using data inside $data_dir
// Also tests createBucket, putObject
function initSetup(S3Client $s3Client, $objects, $data_dir) {
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

}

function testGetBucketLocation($s3Client, $bucket) {
    $result = $s3Client->getBucketLocation(['Bucket' => $bucket]);
    if (getStatusCode($result) != HTTP_OK)
        throw new Exception('getBucketLocation API failed for ' .
                            $bucket);
}

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
}

function testDeleteObjects($s3Client, $bucket, $object) {
    $copies = [];
    for ($i = 0; $i < 3; $i++) {
        $copyKey = $object . '-copy' . strval($i);
        $result = $s3Client->copyObject([
            'Bucket' => $bucket,
            'Key' => $copyKey,
            'CopySource' => $bucket . '/' . $object,
        ]);
        if (getstatuscode($result) != http_ok)
            throw new exception('copyobject api failed for ' .
                                $bucket);
        array_push($copies, ['Key' => $copyKey]);
    }

    print_r($copies);

    $result = $s3Client->deleteObjects([
        'Bucket' => $bucket,
        'Delete' => [
            'Objects' => $copies,
        ],
    ]);
    if (getstatuscode($result) != http_ok)
        throw new exception('deleteObjects api failed for ' .
                            $bucket);
}

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

$data_dir = '/mint/data';
$data_dir = getenv("DATA_DIR");


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
    initSetup($s3Client, $objects, $data_dir);
    $firstBucket = array_keys($objects)[0];
    $firstObject = $objects[$firstBucket];
    testGetBucketLocation($s3Client, $firstBucket);
    testListBuckets($s3Client);
    testBucketExists($s3Client, array_keys($objects));
    testGetPutObject($s3Client, $firstBucket, $firstObject);
    testCopyObject($s3Client, $firstBucket, $firstObject);
    testDeleteObjects($s3Client, $firstBucket, $firstObject);
    testMultipartUpload($s3Client, $firstBucket, $firstObject);
    testAbortMultipartUpload($s3Client, $firstBucket, $firstObject);
}
finally {
    cleanupSetup($s3Client, $objects);
}

?>
