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
                return new Exception("putObject API failed for " . $bucket . '/' . $object);
        }

        finally {
            // close data file
            if (!is_null($stream))
                $stream->close();
        }
    }
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
    'bucket1' => 'obj1',
    'bucket2' => 'obj2',
];

initSetup($s3Client, $objects, $data_dir);
testListBuckets($s3Client);
testBucketExists($s3Client, array_keys($objects));
cleanupSetup($s3Client, $objects);
?>
