/*
 *  Mint, (C) 2018-2023 Minio, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package io.minio.awssdk.v2.tests;

import software.amazon.awssdk.auth.credentials.*;
import software.amazon.awssdk.core.internal.http.loader.DefaultSdkHttpClientBuilder;
import software.amazon.awssdk.core.waiters.WaiterResponse;
import software.amazon.awssdk.http.SdkHttpClient;
import software.amazon.awssdk.http.SdkHttpConfigurationOption;
import software.amazon.awssdk.http.async.SdkAsyncHttpClient;
import software.amazon.awssdk.http.nio.netty.NettyNioAsyncHttpClient;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.S3AsyncClient;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.*;
import software.amazon.awssdk.utils.AttributeMap;

import java.io.IOException;
import java.math.BigInteger;
import java.net.URI;
import java.nio.file.Paths;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.Arrays;
import java.util.List;
import java.util.Random;

public class FunctionalTests {
    private static final String PASS = "PASS";
    private static final String FAILED = "FAIL";
    private static final String IGNORED = "NA";

    private static String accessKey;
    private static String secretKey;
    private static Region region;
    private static String endpoint;
    private static boolean enableHTTPS;

    private static final Random random = new Random(new SecureRandom().nextLong());
    private static final String bucketName = getRandomName();
    private static boolean mintEnv = false;

    private static String file1Kb;
    private static String file1Mb;
    private static String file6Mb;

    private static S3Client s3Client;
    private static S3AsyncClient s3AsyncClient;
    private static S3TestUtils s3TestUtils;

    public static String getRandomName() {
        return "aws-sdk-java-v2-test-" + new BigInteger(32, random).toString(32);
    }

    /**
     * Prints a success log entry in JSON format.
     */
    public static void mintSuccessLog(String function, String args, long startTime) {
        if (mintEnv) {
            System.out.println(
                    new MintLogger(function, args, System.currentTimeMillis() - startTime, PASS, null, null, null));
        }
    }

    /**
     * Prints a failure log entry in JSON format.
     */
    public static void mintFailedLog(String function, String args, long startTime, String message, String error) {
        if (mintEnv) {
            System.out.println(new MintLogger(function, args, System.currentTimeMillis() - startTime, FAILED, null,
                    message, error));
        }
    }

    /**
     * Prints a ignore log entry in JSON format.
     */
    public static void mintIgnoredLog(String function, String args, long startTime) {
        if (mintEnv) {
            System.out.println(
                    new MintLogger(function, args, System.currentTimeMillis() - startTime, IGNORED, null, null, null));
        }
    }

    public static void initTests() throws IOException {
        // Create bucket
        s3Client.createBucket(CreateBucketRequest
                .builder()
                .bucket(bucketName)
                .build());
        s3Client.waiter().waitUntilBucketExists(HeadBucketRequest
                .builder()
                .bucket(bucketName)
                .build());
    }

    // Run tests
    public static void runTests() throws Exception {
        createBucket_test();
        createBucketWithVersion_test();
        uploadObject_test();
        uploadMultiPart_test();
        uploadMultiPartAsync_test();
        uploadObjectVersions_test();
//        uploadSnowballObjects_test();
    }

    public static void createBucket_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: S3Client.createBucket");
        }
        if (!enableHTTPS) {
            return;
        }

        String bucket = getRandomName();
        long startTime = System.currentTimeMillis();
        try {
            s3Client.createBucket(CreateBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            s3Client.waiter().waitUntilBucketExists(HeadBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            mintSuccessLog("S3Client.createBucket", "bucket: " + bucket, startTime);
        } catch (Exception ex) {
            mintFailedLog(
                    "S3Client.createBucket",
                    "bucket: " + bucket,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void createBucketWithVersion_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: S3Client.createBucket");
        }
        if (!enableHTTPS) {
            return;
        }

        String bucket = getRandomName();
        long startTime = System.currentTimeMillis();
        try {
            s3Client.createBucket(CreateBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            s3Client.waiter().waitUntilBucketExists(HeadBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            s3Client.putBucketVersioning(PutBucketVersioningRequest
                    .builder()
                    .bucket(bucket)
                    .versioningConfiguration(VersioningConfiguration
                            .builder()
                            .status(BucketVersioningStatus.ENABLED)
                            .build())
                    .build());
            mintSuccessLog("S3Client.putBucketVersioning", "bucket: " + bucket, startTime);
        } catch (Exception ex) {
            mintFailedLog(
                    "S3Client.putBucketVersioning",
                    "bucket: " + bucket,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void uploadObject_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: S3Client.putObject");
        }
        if (!enableHTTPS) {
            return;
        }

        long startTime = System.currentTimeMillis();
        String file1KbMD5 = Utils.getFileMD5(file1Kb);
        String objectName = "testobject";
        try {
            s3TestUtils.uploadObject(bucketName, objectName, file1Kb);
            s3TestUtils.downloadObject(bucketName, objectName, file1KbMD5);
            mintSuccessLog(
                    "S3Client.putObject",
                    "bucket: " + bucketName + ", object: " + objectName + ", String: " + file1Kb,
                    startTime);
        } catch (Exception ex) {
            mintFailedLog("S3Client.putObject",
                    "bucket: " + bucketName + ", object: " + objectName + ", String: " + file1Kb,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void uploadMultiPart_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: S3Client.uploadPart");
        }
        if (!enableHTTPS) {
            return;
        }

        long startTime = System.currentTimeMillis();
        String objectName = "testobject";
        try {
            s3TestUtils.uploadMultipartObject(bucketName, objectName);
            s3TestUtils.downloadObject(bucketName, objectName, "");
            mintSuccessLog(
                    "S3Client.uploadPart",
                    "bucket: " + bucketName + ", object: " + objectName,
                    startTime);
        } catch (Exception ex) {
            mintFailedLog(
                    "S3Client.uploadPart",
                    "bucket: " + bucketName + ", object: " + objectName,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void uploadMultiPartAsync_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: Async S3Client.uploadPart");
        }
        if (!enableHTTPS) {
            return;
        }

        long startTime = System.currentTimeMillis();
        String objectName = "testobject";
        try {
            s3TestUtils.uploadMultipartObjectAsync(bucketName, objectName);
            s3TestUtils.downloadObject(bucketName, objectName, "");
            mintSuccessLog(
                    "Async S3Client.uploadPart",
                    "bucket: " + bucketName + ", object: " + objectName,
                    startTime);
        } catch (Exception ex) {
            mintFailedLog(
                    "Async S3Client.uploadPart",
                    "bucket: " + bucketName + ", object: " + objectName,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void uploadObjectVersions_test() throws Exception {
        if (!mintEnv) {
            System.out.println("Test: S3Client.putObject versions");
        }
        if (!enableHTTPS) {
            return;
        }

        String bucket = getRandomName();
        long startTime = System.currentTimeMillis();
        String objectName = "testobject";
        try {
            s3Client.createBucket(CreateBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            s3Client.waiter().waitUntilBucketExists(HeadBucketRequest
                    .builder()
                    .bucket(bucket)
                    .build());
            s3Client.putBucketVersioning(PutBucketVersioningRequest
                    .builder()
                    .bucket(bucket)
                    .versioningConfiguration(VersioningConfiguration
                            .builder()
                            .status(BucketVersioningStatus.ENABLED)
                            .build())
                    .build());
            // load same object multiple times
            s3TestUtils.uploadObject(bucketName, objectName, file1Kb);
            s3TestUtils.uploadObject(bucketName, objectName, file1Kb);
            s3TestUtils.uploadObject(bucketName, objectName, file1Kb);
            s3TestUtils.downloadObject(bucketName, objectName, "");
            mintSuccessLog("S3Client.putObject versions",
                    "bucket: " + bucket + ", object: " + objectName,
                    startTime);
        } catch (Exception ex) {
            mintFailedLog("S3Client.putObject versions",
                    "bucket: " + bucket + ", object: " + objectName,
                    startTime,
                    null,
                    ex.toString() + " >>> " + Arrays.toString(ex.getStackTrace()));
            throw ex;
        }
    }

    public static void teardown() throws IOException {
        ListBucketsResponse response = s3Client.listBuckets(ListBucketsRequest
                .builder()
                .build());
        List<Bucket> buckets = response.buckets();
        for (Bucket bucket : buckets) {
            // Remove all objects under the test bucket & the bucket itself
            ListObjectsV2Request request = ListObjectsV2Request
                    .builder()
                    .bucket(bucket.name())
                    .build();
            ListObjectsV2Response listObjectsResponse;
            do {
                listObjectsResponse = s3Client.listObjectsV2(request);
                for (S3Object obj : listObjectsResponse.contents()) {
                    s3Client.deleteObject(DeleteObjectRequest
                            .builder()
                            .bucket(bucket.name())
                            .key(obj.key())
                            .build());
                }
            } while (listObjectsResponse.isTruncated());
            // finally remove the bucket
            s3Client.deleteBucket(DeleteBucketRequest
                    .builder()
                    .bucket(bucket.name())
                    .build());
        }
    }

    public static void main(String[] args) throws Exception, IOException, NoSuchAlgorithmException {
        endpoint = System.getenv("SERVER_ENDPOINT");
        accessKey = System.getenv("ACCESS_KEY");
        secretKey = System.getenv("SECRET_KEY");
        enableHTTPS = System.getenv("ENABLE_HTTPS").equals("1");

        region = Region.US_EAST_1;

        if (enableHTTPS) {
            endpoint = "https://" + endpoint;
        } else {
            endpoint = "http://" + endpoint;
        }

        String dataDir = System.getenv("MINT_DATA_DIR");
        if (dataDir != null && !dataDir.equals("")) {
            mintEnv = true;
            file1Kb = Paths.get(dataDir, "datafile-1-kB").toString();
            file1Mb = Paths.get(dataDir, "datafile-1-MB").toString();
            file6Mb = Paths.get(dataDir, "datafile-6-MB").toString();
        }

        String mintMode = null;
        if (mintEnv) {
            mintMode = System.getenv("MINT_MODE");
        }
        AwsBasicCredentials credentials = AwsBasicCredentials.create(accessKey, secretKey);
        if (enableHTTPS) {
            SdkHttpClient sdkHttpClient = new DefaultSdkHttpClientBuilder().buildWithDefaults(
                    AttributeMap
                            .builder()
                            .put(SdkHttpConfigurationOption.TRUST_ALL_CERTIFICATES, true)
                            .build());
            s3Client = S3Client
                    .builder()
                    .endpointOverride(URI.create(endpoint))
                    .credentialsProvider(StaticCredentialsProvider.create(credentials))
                    .region(region)
                    .httpClient(sdkHttpClient)
                    .build();
            SdkAsyncHttpClient sdkAsyncHttpClient = NettyNioAsyncHttpClient
                    .builder()
                    .buildWithDefaults(AttributeMap
                            .builder()
                            .put(SdkHttpConfigurationOption.TRUST_ALL_CERTIFICATES, true)
                            .build());
            s3AsyncClient = S3AsyncClient
                    .builder()
                    .endpointOverride(URI.create(endpoint))
                    .credentialsProvider(StaticCredentialsProvider.create(credentials))
                    .region(region)
                    .httpClient(sdkAsyncHttpClient)
                    .build();
        } else {
            s3Client = S3Client
                    .builder()
                    .endpointOverride(URI.create(endpoint))
                    .credentialsProvider(StaticCredentialsProvider.create(credentials))
                    .region(region)
                    .build();
            s3AsyncClient = S3AsyncClient
                    .builder()
                    .endpointOverride(URI.create(endpoint))
                    .credentialsProvider(StaticCredentialsProvider.create(credentials))
                    .region(region)
                    .build();
        }

        s3TestUtils = new S3TestUtils(s3Client, s3AsyncClient);

        try {
            initTests();
            FunctionalTests.runTests();
        } catch (Exception e) {
            e.printStackTrace();
            System.exit(-1);
        } finally {
            teardown();
        }
    }
}
