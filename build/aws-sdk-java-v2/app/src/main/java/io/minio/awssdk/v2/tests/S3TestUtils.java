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

import software.amazon.awssdk.core.ResponseInputStream;
import software.amazon.awssdk.core.async.AsyncRequestBody;
import software.amazon.awssdk.core.sync.RequestBody;
import software.amazon.awssdk.services.s3.S3AsyncClient;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.*;


import java.io.File;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.concurrent.CompletableFuture;

public class S3TestUtils {
    private final S3Client s3Client;
    private final S3AsyncClient s3AsyncClient;
    private final S3AsyncClient s3CrtAsyncClient;

    S3TestUtils(S3Client s3Client, S3AsyncClient s3AsyncClient, S3AsyncClient s3CrtAsyncClient) {
	this.s3Client = s3Client;
	this.s3AsyncClient = s3AsyncClient;
	this.s3CrtAsyncClient = s3CrtAsyncClient;
    }

    void uploadMultipartObject(String bucketName, String keyName) throws IOException {
        int mB = 1024 * 1024;

        CreateMultipartUploadRequest.Builder builder = CreateMultipartUploadRequest.builder();
        builder.bucket(bucketName).key(keyName);

        // Step 1: Initialize.
        CreateMultipartUploadResponse res = s3Client.createMultipartUpload(builder.build());
        String uploadId = res.uploadId();
        List<CompletedPart> completedParts = new ArrayList<>();

        // Step 2: Upload parts.
        for (int i = 1; i <= 10; i++) {
            UploadPartRequest req = UploadPartRequest
                    .builder()
                    .bucket(bucketName)
                    .key(keyName)
                    .uploadId(uploadId)
                    .partNumber(i)
                    .build();
            UploadPartResponse resp = s3Client.uploadPart(req, RequestBody.fromByteBuffer(getRandomByteBuffer(5 * mB)));
            String etag = resp.eTag();
            CompletedPart part = CompletedPart
                    .builder()
                    .partNumber(i)
                    .eTag(etag)
                    .build();
            completedParts.add(part);
        }

        // Step 3: Complete.
        CompletedMultipartUpload completedMultipartUpload = CompletedMultipartUpload.builder().parts(completedParts).build();
        CompleteMultipartUploadRequest completeMultipartUploadRequest =
                CompleteMultipartUploadRequest.builder()
                        .bucket(bucketName).key(keyName)
                        .uploadId(uploadId)
                        .multipartUpload(completedMultipartUpload)
                        .build();
        s3Client.completeMultipartUpload(completeMultipartUploadRequest);
    }

    void uploadMultipartObjectAsync(String bucketName, String keyName) throws Exception {
        int mB = 1024 * 1024;

        CreateMultipartUploadRequest.Builder builder = CreateMultipartUploadRequest.builder();
        builder.bucket(bucketName).key(keyName);

        // Step 1: Initialize.
        CompletableFuture<CreateMultipartUploadResponse> res = s3AsyncClient.createMultipartUpload(builder.build());
        String uploadId = res.get().uploadId();
        List<CompletedPart> completedParts = new ArrayList<>();

        // Step 2: Upload parts.
        for (int i = 1; i <= 10; i++) {
            UploadPartRequest req = UploadPartRequest
                    .builder()
                    .bucket(bucketName)
                    .key(keyName)
                    .uploadId(uploadId)
                    .partNumber(i)
                    .build();
            CompletableFuture<UploadPartResponse> resp = s3AsyncClient.uploadPart(req, AsyncRequestBody.fromByteBuffer(getRandomByteBuffer(5 * mB)));
            String etag = resp.get().eTag();
            CompletedPart part = CompletedPart
                    .builder()
                    .partNumber(i)
                    .eTag(etag)
                    .build();
            completedParts.add(part);
        }

        // Step 3: Complete.
        CompletedMultipartUpload completedMultipartUpload = CompletedMultipartUpload.builder().parts(completedParts).build();
        CompleteMultipartUploadRequest completeMultipartUploadRequest =
                CompleteMultipartUploadRequest.builder()
                        .bucket(bucketName).key(keyName)
                        .uploadId(uploadId)
                        .multipartUpload(completedMultipartUpload)
                        .build();
        s3AsyncClient.completeMultipartUpload(completeMultipartUploadRequest);
    }

    private static ByteBuffer getRandomByteBuffer(int size) throws IOException {
        byte[] b = new byte[size];
        new Random().nextBytes(b);
        return ByteBuffer.wrap(b);
    }

    void uploadObject(String bucketName, String keyName, String filePath) throws IOException {

        File f = new File(filePath);
        PutObjectRequest request = PutObjectRequest
                .builder()
                .bucket(bucketName)
                .key(keyName)
                .build();
        s3Client.putObject(request, RequestBody.fromFile(f));
    }

    void downloadObject(String bucketName, String keyName, String expectedMD5) throws Exception, IOException {
        GetObjectRequest request = GetObjectRequest
                .builder()
                .bucket(bucketName)
                .key(keyName)
                .build();
        ResponseInputStream<GetObjectResponse> response = s3Client.getObject(request);

        String calculatedMD5 = Utils.getBufferMD5(response.readAllBytes());

        if (!expectedMD5.equals("") && !calculatedMD5.equals(expectedMD5)) {
            throw new Exception("downloaded object has unexpected md5sum, expected: " + expectedMD5 + ", found: " + calculatedMD5);
        }
    }

    void copyObject(String bucketName, String keyName,
                    String targetBucketName, String targetKeyName, String newSseKey,
                    boolean replace) {
        CopyObjectRequest.Builder builder = CopyObjectRequest.builder();
        builder.sourceBucket(bucketName).sourceKey(keyName).destinationBucket(targetBucketName).destinationKey(targetKeyName);
        if (replace) {
            builder.metadataDirective(MetadataDirective.COPY);
        }
        s3Client.copyObject(builder.build());
    }

    long retrieveObjectMetadata(String bucketName, String keyName, String sseKey) {
        GetObjectRequest request = GetObjectRequest
                .builder()
                .bucket(bucketName)
                .key(keyName)
                .sseCustomerKey(sseKey)
                .build();
        ResponseInputStream<GetObjectResponse> response = s3Client.getObject(request);
        return  response.response().contentLength();
    }
}
