/*
 * Mint (C) 2017 Minio, Inc.
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

import java.security.*;
import java.math.BigInteger;
import java.util.*;

import javax.crypto.KeyGenerator;
import javax.crypto.SecretKey;

import java.io.*;

import static java.nio.file.StandardOpenOption.*;
import java.nio.file.*;

import org.joda.time.DateTime;
/*
import okhttp3.OkHttpClient;
import okhttp3.HttpUrl;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.MultipartBody;
import okhttp3.Response;
*/
import com.google.common.io.ByteStreams;

import io.minio.*;
import io.minio.messages.*;
import io.minio.errors.*;


public class FunctionalTest {
  private static final int MB = 1024 * 1024;
  private static final Random random = new Random(new SecureRandom().nextLong());
  private static final String bucketName = getRandomName();
  private static final String customContentType = "application/javascript";
  private static String endpoint;
  private static String accessKey;
  private static String secretKey;
  private static String region;
  private static String mode;
  private static MinioClient client = null;
  private static String dataDir=System.getenv("DATA_DIR");
  private static String FileOfSize1b=dataDir + "/datafile-1-b";

  private static String FileOfSize6mb=dataDir + "/datafile-6-MB";

  private static String smallFile=dataDir + "/datafile-1-MB";
  private static String largeFile=dataDir + "/datafile-11-MB";

  /**
   * Do no-op.
   */
  public static void ignore(Object ...args) {
  }

  /**
   * Create given sized file and returns its name.
   */
  public static String createFile(int size) throws IOException {
    String filename = getRandomName();

    OutputStream os = Files.newOutputStream(Paths.get(filename), CREATE, APPEND);
    int totalBytesWritten = 0;
    int bytesToWrite = 0;
    byte[] buf = new byte[1 * MB];
    while (totalBytesWritten < size) {
      random.nextBytes(buf);
      bytesToWrite = size - totalBytesWritten;
      if (bytesToWrite > buf.length) {
        bytesToWrite = buf.length;
      }

      os.write(buf, 0, bytesToWrite);
      totalBytesWritten += bytesToWrite;
    }
    os.close();

    return filename;
  }

  /**
   * Generate random name.
   */
  public static String getRandomName() {
    return "minio-java-test-" + new BigInteger(32, random).toString(32);
  }

  /**
   * Test: makeBucket(String bucketName).
   */
  public static void makeBucket_test1() throws Exception {
    System.out.println("Test: makeBucket(String bucketName)");
    String name = getRandomName();
    client.makeBucket(name);
    client.removeBucket(name);
  }

  /**
   * Test: makeBucket(String bucketName, String region).
   */
  public static void makeBucket_test2() throws Exception {
    System.out.println("Test: makeBucket(String bucketName, String region)");
    String name = getRandomName();
    client.makeBucket(name, "eu-west-1");
    client.removeBucket(name);
  }

  /**
   * Test: makeBucket(String bucketName, String region) where bucketName has
   * periods in its name.
   */
  public static void makeBucket_test3() throws Exception {
    System.out.println("Test: makeBucket(String bucketName, String region)");
    String name = getRandomName() + ".withperiod";
    client.makeBucket(name, "eu-central-1");
    client.removeBucket(name);
  }

  /**
   * Test: listBuckets().
   */
  public static void listBuckets_test() throws Exception {
    System.out.println("Test: listBuckets()");
    for (Bucket bucket : client.listBuckets()) {
      ignore(bucket);
    }
  }

  /**
   * Test: bucketExists(String bucketName).
   */
  public static void bucketExists_test() throws Exception {
    System.out.println("Test: bucketExists(String bucketName)");
    String name = getRandomName();
    client.makeBucket(name);
    if (!client.bucketExists(name)) {
      throw new Exception("[FAILED] Test: bucketExists(String bucketName)");
    }
    client.removeBucket(name);
  }

  /**
   * Test: removeBucket(String bucketName).
   */
  public static void removeBucket_test() throws Exception {
    System.out.println("Test: removeBucket(String bucketName)");
    String name = getRandomName();
    client.makeBucket(name);
    client.removeBucket(name);
  }

  /**
   * Tear down test setup.
   */
  public static void setup() throws Exception {
    client.makeBucket(bucketName);
  }

  /**
   * Tear down test setup.
   */
  public static void teardown() throws Exception {
    client.removeBucket(bucketName);
  }

  /**
   * Test: putObject(String bucketName, String objectName, String filename).
   */
  public static void putObject_test1() throws Exception {
    System.out.println("Test: putObject(String bucketName, String objectName, String filename)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: multipart: putObject(String bucketName, String objectName, String filename).
   */
  public static void putObject_test2() throws Exception {
    System.out.println("Test: multipart: putObject(String bucketName, String objectName, String filename)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, largeFile);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: multipart resume: putObject(String bucketName, String objectName, String filename).
   */
  public static void putObject_test3() throws Exception {
    System.out.println("Test: multipart resume: putObject(String bucketName, String objectName, String filename)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(largeFile));
    try {
      client.putObject(bucketName, objectName, is, 20 * 1024 * 1024, null);
    } catch (InsufficientDataException e) {
      ignore();
    }
    is.close();

    client.putObject(bucketName, objectName, largeFile);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: putObject(String bucketName, String objectName, String contentType, long size, InputStream body).
   */
  public static void putObject_test4() throws Exception {
    System.out.println("Test: putObject(String bucketName, String objectName, String contentType, long size, "
                       + "InputStream body)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(smallFile));
    client.putObject(bucketName, objectName, is, 1024 * 1024, customContentType);
    is.close();
    ObjectStat objectStat = client.statObject(bucketName, objectName);
    if (!customContentType.equals(objectStat.contentType())) {
      throw new Exception("[FAILED] Test: putObject(String bucketName, String objectName, String contentType, "
                          + "long size, InputStream body)");
    }
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: With content-type: putObject(String bucketName, String objectName, String filename, String contentType).
   */
  public static void putObject_test5() throws Exception {
    System.out.println("Test: putObject(String bucketName, String objectName, String filename,"
                       + " String contentType)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, largeFile, customContentType);
    ObjectStat objectStat = client.statObject(bucketName, objectName);
    if (!customContentType.equals(objectStat.contentType())) {
      throw new Exception("[FAILED] Test: putObject(String bucketName, String objectName, String filename,"
                          + " String contentType)");
    }
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: putObject(String bucketName, String objectName, String filename).
   * where objectName has multiple path segments.
   */
  public static void putObject_test6() throws Exception {
    System.out.println("Test: objectName with path segments: "
                       + "putObject(String bucketName, String objectName, String filename)");
    String objectName = "path/to/" + getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: client side encryption: putObject(String bucketName, String objectName, InputStream stream, long * size,
   * String contentType, SecretKey key).
   */
  /*
  public static void putObject_test7() throws Exception {
    System.out.println("Test: putObject(String bucketName, String objectName, InputStream body, "
                       + "String contentType)");
    String filename = createFile(3 * MB);
    InputStream is = Files.newInputStream(Paths.get(filename));
    client.putObject(bucketName, filename, is, customContentType);
    is.close();
    Files.delete(Paths.get(filename));
    ObjectStat objectStat = client.statObject(bucketName, filename);
    if (!customContentType.equals(objectStat.contentType())) {
      throw new Exception("[FAILED] Test: putObject(String bucketName, String objectName, String contentType, "
                          + "long size, InputStream body)");
    }
    client.removeObject(bucketName, filename);
  }
  */

  /**
   * Test: multipart: putObject(String bucketName, String objectName, InputStream body, String contentType).
   */
  /*
  public static void putObject_test8() throws Exception {
    System.out.println("Test: multipart: putObject(String bucketName, String objectName, InputStream body, "
                       + "String contentType)");
    String filename = createFile(537 * MB);
    InputStream is = Files.newInputStream(Paths.get(filename));
    client.putObject(bucketName, filename, is, customContentType);
    is.close();
    Files.delete(Paths.get(filename));
    ObjectStat objectStat = client.statObject(bucketName, filename);
    if (!customContentType.equals(objectStat.contentType())) {
      throw new Exception("[FAILED] Test: putObject(String bucketName, String objectName, String contentType, "
                          + "long size, InputStream body)");
    }
    client.removeObject(bucketName, filename);
  }
  */

  /**
   * Test: client side encryption: putObject(String bucketName, String objectName, InputStream stream, long size, String
   * contentType, SecretKey key).
   */
  /*
  public static void putObject_test9() throws Exception {

    System.out.println(
        "Test: encryption (AES): putObject(String bucketName, String objectName, InputStream stream,long size, "
            + "String contentType, SecretKey key).");
    String fileName = createFile(13 * MB);
    InputStream is = Files.newInputStream(Paths.get(fileName));

    // Generate key with 128 bit key.
    KeyGenerator symKeyGenerator = KeyGenerator.getInstance("AES");
    symKeyGenerator.init(128);
    SecretKey symKey = symKeyGenerator.generateKey();

    try {
      client.putObject(bucketName, fileName, is, 13 * 1024 * 1024, null, symKey);
    } catch (InsufficientDataException e) {
      throw new MinioException("Insufficient data received");
    }
    is.close();

    Files.delete(Paths.get(fileName));
    client.removeObject(bucketName, fileName);

  }
 */
  /**
   * Test: client side encryption: putObject(String bucketName, String objectName, InputStream stream, long size, String
   * contentType, KeyPair keypair).
   */
  /*
  public static void putObject_test10() throws Exception {

    System.out.println(
        "Test: encryption (RSA): putObject(String bucketName, String objectName, InputStream stream, "
            + "long size, String contentType, KeyPair keypair).");
    String fileName = createFile(13 * MB);
    InputStream is = Files.newInputStream(Paths.get(fileName));

    // Generate RSA key pair
    KeyPairGenerator keyGenerator = KeyPairGenerator.getInstance("RSA");
    keyGenerator.initialize(1024, new SecureRandom());
    KeyPair keypair = keyGenerator.generateKeyPair();

    try {
      client.putObject(bucketName, fileName, is, 13 * 1024 * 1024, null, keypair);
    } catch (InsufficientDataException e) {
      throw new MinioException("Insufficient data received");
    }
    is.close();

    Files.delete(Paths.get(fileName));
    client.removeObject(bucketName, fileName);
  }
  */
  /**
   * Test: statObject(String bucketName, String objectName).
   */
  public static void statObject_test() throws Exception {
    System.out.println("Test: statObject(String bucketName, String objectName)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.statObject(bucketName, objectName);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: getObject(String bucketName, String objectName).
   */
  public static void getObject_test1() throws Exception {
    System.out.println("Test: getObject(String bucketName, String objectName)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    InputStream is = client.getObject(bucketName, objectName);
    is.close();
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: getObject(String bucketName, String objectName, long offset).
   */
  public static void getObject_test2() throws Exception {
    System.out.println("Test: getObject(String bucketName, String objectName, long offset)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    InputStream is = client.getObject(bucketName, objectName, 1000L);
    is.close();
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: getObject(String bucketName, String objectName, long offset, Long length).
   */
  public static void getObject_test3() throws Exception {
    System.out.println("Test: getObject(String bucketName, String objectName, long offset, Long length)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    InputStream is = client.getObject(bucketName, objectName, 1000L, 1024 * 1024L);
    is.close();
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: getObject(String bucketName, String objectName, String filename).
   */
  public static void getObject_test4() throws Exception {
    System.out.println("Test: getObject(String bucketName, String objectName, String filename)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.getObject(bucketName, objectName, smallFile + ".downloaded");
    Files.delete(Paths.get(smallFile + ".downloaded"));
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: getObject(String bucketName, String objectName, String filename).
   * where objectName has multiple path segments.
   */
  public static void getObject_test5() throws Exception {
    System.out.println("Test: objectName with multiple path segments: "
                       + "getObject(String bucketName, String objectName, String filename)");
    String objectName = "path/to/" + getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.getObject(bucketName, objectName, smallFile + ".downloaded");
    Files.delete(Paths.get(smallFile + ".downloaded"));
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test client side encryption (AES): getObject(String bucketName, String objectName, SecretKey key).
   */
  /*
  public static void getObject_test6() throws Exception {
    System.out.println(
        "Test: encryption (AES): getObject(String bucketName, String objectName, "
            + "SecretKey key).");

    String fileName = createFile(13 * MB);
    InputStream is = Files.newInputStream(Paths.get(fileName));

    // Generate key with 128 bit keys.
    KeyGenerator symKeyGenerator = KeyGenerator.getInstance("AES");
    symKeyGenerator.init(128);
    SecretKey symKey = symKeyGenerator.generateKey();

    try {
      // Put an encrypted object
      client.putObject(bucketName, fileName, is, 13 * 1024 * 1024, null, symKey);
    } catch (InsufficientDataException e) {
      throw new MinioException("Insufficient data received");
    }
    is.close();

    // Get the object without decryption
    InputStream ois = client.getObject(bucketName, fileName);
    String plainFileName = getRandomName();
    OutputStream os = Files.newOutputStream(Paths.get(plainFileName));

    // Read in the decrypted bytes and write to fileNameOut
    int numRead = 0;
    byte[] buf = new byte[8192];
    while ((numRead = ois.read(buf)) >= 0) {
      os.write(buf, 0, numRead);
    }
    os.close();
    ois.close();

    if (Arrays.equals(Files.readAllBytes(Paths.get(plainFileName)), Files.readAllBytes(Paths.get(fileName)))) {
      throw new MinioException("Files should not be equal");
    }

    // Get the object with decryption
    ois = client.getObject(bucketName, fileName, symKey);
    String decFileName = getRandomName();
    os = Files.newOutputStream(Paths.get(decFileName));

    numRead = 0;
    buf = new byte[8192];
    while ((numRead = ois.read(buf)) >= 0) {
      os.write(buf, 0, numRead);
    }
    os.close();
    ois.close();

    // Check if two files are not equal
    if (!Arrays.equals(Files.readAllBytes(Paths.get(decFileName)), Files.readAllBytes(Paths.get(fileName)))) {
      throw new MinioException("Files should be equal");
    }

    Files.delete(Paths.get(fileName));
    Files.delete(Paths.get(plainFileName));
    Files.delete(Paths.get(decFileName));
    client.removeObject(bucketName, fileName);
  }
  /**
   * Test: client side encryption (RSA): getObject(String bucketName, String objectName, KeyPair keyPair).
   */
  /*
  public static void getObject_test7() throws Exception {
    System.out.println(
        "Test: encryption (RSA): getObject(String bucketName, String objectName, "
            + "KeyPair keyPair).");

    String fileName = createFile(13 * MB);
    InputStream is = Files.newInputStream(Paths.get(fileName));

    // Generate RSA key pair
    KeyPairGenerator keyGenerator = KeyPairGenerator.getInstance("RSA");
    keyGenerator.initialize(1024, new SecureRandom());
    KeyPair keypair = keyGenerator.generateKeyPair();

    try {
      // Put an encrypted object
      client.putObject(bucketName, fileName, is, 13 * 1024 * 1024, null, keypair);
    } catch (InsufficientDataException e) {
      throw new MinioException("Insufficient data received");
    }
    is.close();

    // Get the object without decryption
    String plainFileName = getRandomName();
    InputStream ois = client.getObject(bucketName, fileName);
    OutputStream os = Files.newOutputStream(Paths.get(plainFileName));

    // Read in the decrypted bytes and write to fileNameOut
    int numRead = 0;
    byte[] buf = new byte[8192];
    while ((numRead = ois.read(buf)) >= 0) {
      os.write(buf, 0, numRead);
    }
    os.close();
    ois.close();

    // Check if two files are not equal
    if (Arrays.equals(Files.readAllBytes(Paths.get(plainFileName)), Files.readAllBytes(Paths.get(fileName)))) {
      throw new MinioException("Files should not be equal");
    }


    // Get the object with decryption
    String decFileName = createFile(0);
    ois = client.getObject(bucketName, fileName, keypair);
    os = Files.newOutputStream(Paths.get(decFileName));

    numRead = 0;
    buf = new byte[8192];
    while ((numRead = ois.read(buf)) >= 0) {
      os.write(buf, 0, numRead);
    }
    os.close();
    ois.close();

    // Check if two files are equal
    if (!Arrays.equals(Files.readAllBytes(Paths.get(decFileName)), Files.readAllBytes(Paths.get(fileName)))) {
      throw new MinioException("Files should be equal");
    }

    Files.delete(Paths.get(fileName));
    Files.delete(Paths.get(plainFileName));
    Files.delete(Paths.get(decFileName));
    client.removeObject(bucketName, fileName);
  }
*/
  /**
   * Test: listObjects(final String bucketName).
   */
  public static void listObject_test1() throws Exception {
    int i;
    System.out.println("Test: listObjects(final String bucketName)");
    String[] objectNames = new String[3];
    for (i = 0; i < 3; i++) {
      String objectName = getRandomName();
      client.putObject(bucketName, objectName,smallFile);
      objectNames[i] = objectName;
    }

    i = 0;
    for (Result<?> r : client.listObjects(bucketName)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    for (i = 0; i < 3; i++) {
      client.removeObject(bucketName, objectNames[i]);
    }
  }

  /**
   * Test: listObjects(bucketName, final String prefix).
   */
  public static void listObject_test2() throws Exception {
    int i;
    System.out.println("Test: listObjects(final String bucketName, final String prefix)");
    String[] objectNames = new String[3];
    for (i = 0; i < 3; i++) {
      String objectName = getRandomName();
      client.putObject(bucketName, objectName, smallFile);
      objectNames[i] = objectName;
    }

    i = 0;
    for (Result<?> r : client.listObjects(bucketName, "minio")) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    for (i = 0; i < 3; i++) {
      client.removeObject(bucketName, objectNames[i]);
    }
  }

  /**
   * Test: listObjects(bucketName, final String prefix, final boolean recursive).
   */
  public static void listObject_test3() throws Exception {
    int i;
    System.out.println("Test: listObjects(final String bucketName, final String prefix, final boolean recursive)");
    String[] objectNames = new String[3];
    for (i = 0; i < 3; i++) {
      String objectName = getRandomName();
      client.putObject(bucketName, objectName, smallFile);
      objectNames[i] = objectName;
    }

    i = 0;
    for (Result<?> r : client.listObjects(bucketName, "minio", true)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    for (i = 0; i < 3; i++) {
      client.removeObject(bucketName, objectNames[i]);
    }
  }

  /**
   * Test: listObjects(final string bucketName).
   */
  public static void listObject_test4() throws Exception {
    int i;
    System.out.println("Test: empty bucket: listObjects(final String bucketName)");

    i = 0;
    for (Result<?> r : client.listObjects(bucketName, "minio", true)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }
  }

  /**
   * Test: recursive: listObjects(bucketName, final String prefix, final boolean recursive).
   */
  public static void listObject_test5() throws Exception {
    int i;
    int objCount = 1050;

    System.out.println("Test: recursive: listObjects(final String bucketName, final String prefix"
           + ", final boolean recursive)");
    String[] objectNames = new String[objCount];

    String baseObjectName = getRandomName();
    for (i = 0; i < objCount; i++) {
      objectNames[i] = baseObjectName + "-" + i;
      client.putObject(bucketName, objectNames[i],FileOfSize1b);
    }

    i = 0;
    for (Result<?> r : client.listObjects(bucketName, "minio", true)) {
      ignore(i++, r.get());
    }

    // Check the number of uploaded objects
    if (i != objCount) {
      throw new Exception("[FAILED] Test: recursive: listObject_test5(), number of items, expected: "
           + objCount + ", found: " + i);
    }

    for (i = 0; i < objCount; i++) {
      client.removeObject(bucketName, objectNames[i]);
    }
  }

  /**
   * Test: listObjects(bucketName, final String prefix, final boolean recursive, final boolean useVersion1).
   */
  public static void listObject_test6() throws Exception {
    int i;
    System.out.println("Test: listObjects(final String bucketName, final String prefix, final boolean recursive,"
                       + " final boolean useVersion1)");
    String[] objectNames = new String[3];
    for (i = 0; i < 3; i++) {
      String objectName = getRandomName();
      client.putObject(bucketName, objectName, smallFile);
      objectNames[i] = objectName;
    }

    i = 0;
    for (Result<?> r : client.listObjects(bucketName, "minio", true, true)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    for (i = 0; i < 3; i++) {
      client.removeObject(bucketName, objectNames[i]);
    }
  }

  /**
   * Test: removeObject(String bucketName, String objectName).
   */
  public static void removeObject_test1() throws Exception {
    System.out.println("Test: removeObject(String bucketName, String objectName)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);
    client.removeObject(bucketName, objectName);
  }

  /**
   * Test: removeObject(final String bucketName, final Iterable&lt;String&gt; objectNames).
   */
  public static void removeObject_test2() throws Exception {
    System.out.println("Test: removeObject(final String bucketName, final Iterable<String> objectNames)");

    String[] objectNames = new String[4];
    for (int i = 0; i < 3; i++) {
      String objectName = getRandomName();
      client.putObject(bucketName, objectName, smallFile);
      objectNames[i] = objectName;
    }
    objectNames[3] = "nonexistent-object";

    for (Result<?> r : client.removeObject(bucketName, Arrays.asList(objectNames))) {
      ignore(r.get());
    }
  }

  /**
   * Test: listIncompleteUploads(String bucketName).
   */
  public static void listIncompleteUploads_test1() throws Exception {
    System.out.println("Test: listIncompleteUploads(String bucketName)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(FileOfSize6mb));
    try {
      client.putObject(bucketName, objectName, is, 9 * 1024 * 1024, null);
    } catch (InsufficientDataException e) {
      ignore("Exception occurred as excepted");
    }
    is.close();

    int i = 0;
    for (Result<Upload> r : client.listIncompleteUploads(bucketName)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    client.removeIncompleteUpload(bucketName, objectName);
  }

  /**
   * Test: listIncompleteUploads(String bucketName, String prefix).
   */
  public static void listIncompleteUploads_test2() throws Exception {
    System.out.println("Test: listIncompleteUploads(String bucketName, String prefix)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(FileOfSize6mb));
    try {
      client.putObject(bucketName, objectName, is, 9 * 1024 * 1024, null);
    } catch (InsufficientDataException e) {
      ignore("Exception occurred as excepted");
    }
    is.close();

    int i = 0;
    for (Result<Upload> r : client.listIncompleteUploads(bucketName, "minio")) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    client.removeIncompleteUpload(bucketName, objectName);
  }

  /**
   * Test: listIncompleteUploads(final String bucketName, final String prefix, final boolean recursive).
   */
  public static void listIncompleteUploads_test3() throws Exception {
    System.out.println("Test: listIncompleteUploads(final String bucketName, final String prefix, "
                       + "final boolean recursive)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(FileOfSize6mb));
    try {
      client.putObject(bucketName, objectName, is, 9 * 1024 * 1024, null);
    } catch (InsufficientDataException e) {
      ignore("Exception occurred as excepted");
    }
    is.close();

    int i = 0;
    for (Result<Upload> r : client.listIncompleteUploads(bucketName, "minio", true)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    client.removeIncompleteUpload(bucketName, objectName);
  }

  /**
   * Test: removeIncompleteUpload(String bucketName, String objectName).
   */
  public static void removeIncompleteUploads_test() throws Exception {
    System.out.println("Test: removeIncompleteUpload(String bucketName, String objectName)");
    String objectName = getRandomName();
    InputStream is = Files.newInputStream(Paths.get(FileOfSize6mb));
    try {
      client.putObject(bucketName, objectName, is, 9 * 1024 * 1024, null);
    } catch (InsufficientDataException e) {
      ignore("Exception occurred as excepted");
    }
    is.close();

    int i = 0;
    for (Result<Upload> r : client.listIncompleteUploads(bucketName)) {
      ignore(i++, r.get());
      if (i == 10) {
        break;
      }
    }

    client.removeIncompleteUpload(bucketName, objectName);
  }

  /**
   * public String presignedGetObject(String bucketName, String objectName).
   */
  /*
  public static void presignedGetObject_test1() throws Exception {
    System.out.println("Test: presignedGetObject(String bucketName, String objectName)");
    String filename = createFile(3 * MB);
    client.putObject(bucketName, filename, filename);

    String urlString = client.presignedGetObject(bucketName, filename);
    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder
        .url(HttpUrl.parse(urlString))
        .method("GET", null)
        .build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (response.isSuccessful()) {
        OutputStream os = Files.newOutputStream(Paths.get(filename + ".downloaded"), StandardOpenOption.CREATE);
        ByteStreams.copy(response.body().byteStream(), os);
        response.body().close();
        os.close();
      } else {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName)"
                          + ", Error: <No response from server>");
    }

    if (!Arrays.equals(Files.readAllBytes(Paths.get(filename)),
                       Files.readAllBytes(Paths.get(filename + ".downloaded")))) {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName)"
                          + ", Error: <Content differs>");
    }

    Files.delete(Paths.get(filename));
    Files.delete(Paths.get(filename + ".downloaded"));
    client.removeObject(bucketName, filename);
  }
  */
  /**
   * Test: presignedGetObject(String bucketName, String objectName, Integer expires).
   */
  /*
  public static void presignedGetObject_test2() throws Exception {
    System.out.println("Test: presignedGetObject(String bucketName, String objectName, Integer expires)");
    String filename = createFile(3 * MB);
    client.putObject(bucketName, filename, filename);

    String urlString = client.presignedGetObject(bucketName, filename, 3600);
    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder
        .url(HttpUrl.parse(urlString))
        .method("GET", null)
        .build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (response.isSuccessful()) {
        OutputStream os = Files.newOutputStream(Paths.get(filename + ".downloaded"), StandardOpenOption.CREATE);
        ByteStreams.copy(response.body().byteStream(), os);
        response.body().close();
        os.close();
      } else {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName, Integer expires)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName, Integer expires)"
                          + ", Error: <No response from server>");
    }

    if (!Arrays.equals(Files.readAllBytes(Paths.get(filename)),
                       Files.readAllBytes(Paths.get(filename + ".downloaded")))) {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName, Integer expires)"
                          + ", Error: <Content differs>");
    }

    Files.delete(Paths.get(filename));
    Files.delete(Paths.get(filename + ".downloaded"));
    client.removeObject(bucketName, filename);
  }
  */

  /**
   * public String presignedGetObject(String bucketName, String objectName, Integer expires, Map reqParams).
   */
  /*
  public static void presignedGetObject_test3() throws Exception {
    System.out.println("Test: presignedGetObject(String bucketName, String objectName, Integer expires, "
                       + "Map<String, String> reqParams)");
    String filename = createFile(3 * MB);
    client.putObject(bucketName, filename, filename);

    Map<String, String> reqParams = new HashMap<>();
    reqParams.put("response-content-type", "application/json");

    String urlString = client.presignedGetObject(bucketName, filename, 3600, reqParams);
    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder
        .url(HttpUrl.parse(urlString))
        .method("GET", null)
        .build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (response.isSuccessful()) {
        OutputStream os = Files.newOutputStream(Paths.get(filename + ".downloaded"), StandardOpenOption.CREATE);
        ByteStreams.copy(response.body().byteStream(), os);
        if (!response.header("Content-Type").equals("application/json")) {
          throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName,"
                              + " Integer expires, Map<String, String> reqParams)"
                              + ", Response: " + response);
        }
        response.body().close();
        os.close();
      } else {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName,"
                            + " Integer expires, Map<String, String> reqParams)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName,"
                          + " Integer expires, Map<String, String> reqParams)"
                          + ", Error: <No response from server>");
    }

    if (!Arrays.equals(Files.readAllBytes(Paths.get(filename)),
                       Files.readAllBytes(Paths.get(filename + ".downloaded")))) {
      throw new Exception("[FAILED] Test: presignedGetObject(String bucketName, String objectName,"
                          + " Integer expires, Map<String, String> reqParams)"
                          + ", Error: <Content differs>");
    }

    Files.delete(Paths.get(filename));
    Files.delete(Paths.get(filename + ".downloaded"));
    client.removeObject(bucketName, filename);
  }
*/
  /**
   * public String presignedPutObject(String bucketName, String objectName).
   */
  /*
  public static void presignedPutObject_test1() throws Exception {
    System.out.println("Test: presignedPutObject(String bucketName, String objectName)");
    String filename = createFile(3 * MB);
    String urlString = client.presignedPutObject(bucketName, filename);

    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder
        .url(HttpUrl.parse(urlString))
        .method("PUT", RequestBody.create(null, Files.readAllBytes(Paths.get(filename))))
        .build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (!response.isSuccessful()) {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedPutObject(String bucketName, String objectName)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedPutObject(String bucketName, String objectName)"
                          + ", Error: <No response from server>");
    }

    Files.delete(Paths.get(filename));
    client.removeObject(bucketName, filename);
  }
*/
  /**
   * Test: presignedPutObject(String bucketName, String objectName, Integer expires).
   */
  /*
  public static void presignedPutObject_test2() throws Exception {
    System.out.println("Test: presignedPutObject(String bucketName, String objectName, Integer expires)");
    String filename = createFile(3 * MB);
    String urlString = client.presignedPutObject(bucketName, filename, 3600);

    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder
        .url(HttpUrl.parse(urlString))
        .method("PUT", RequestBody.create(null, Files.readAllBytes(Paths.get(filename))))
        .build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (!response.isSuccessful()) {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedPutObject(String bucketName, String objectName, Integer expires)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedPutObject(String bucketName, String objectName, Integer expires)"
                          + ", Error: <No response from server>");
    }

    Files.delete(Paths.get(filename));
    client.removeObject(bucketName, filename);
  }
*/
  /**
   * Test: presignedPostPolicy(PostPolicy policy).
   */
  /*
  public static void presignedPostPolicy_test() throws Exception {
    System.out.println("Test: presignedPostPolicy(PostPolicy policy)");
    String filename = createFile(3 * MB);
    PostPolicy policy = new PostPolicy(bucketName, filename, DateTime.now().plusDays(7));
    policy.setContentRange(1 * MB, 4 * MB);
    Map<String, String> formData = client.presignedPostPolicy(policy);

    MultipartBody.Builder multipartBuilder = new MultipartBody.Builder();
    multipartBuilder.setType(MultipartBody.FORM);
    for (Map.Entry<String, String> entry : formData.entrySet()) {
      multipartBuilder.addFormDataPart(entry.getKey(), entry.getValue());
    }
    multipartBuilder.addFormDataPart("file", filename, RequestBody.create(null, new File(filename)));

    Request.Builder requestBuilder = new Request.Builder();
    Request request = requestBuilder.url(endpoint + "/" + bucketName).post(multipartBuilder.build()).build();
    OkHttpClient transport = new OkHttpClient();
    Response response = transport.newCall(request).execute();

    if (response != null) {
      if (!response.isSuccessful()) {
        String errorXml = "";

        // read entire body stream to string.
        Scanner scanner = new Scanner(response.body().charStream());
        scanner.useDelimiter("\\A");
        if (scanner.hasNext()) {
          errorXml = scanner.next();
        }
        scanner.close();

        throw new Exception("[FAILED] Test: presignedPostPolicy(PostPolicy policy)"
                            + ", Response: " + response
                            + ", Error: " + errorXml);
      }
    } else {
      throw new Exception("[FAILED] Test: presignedPostPolicy(PostPolicy policy)"
                          + ", Error: <No response from server>");
    }

    Files.delete(Paths.get(filename));
    client.removeObject(bucketName, filename);
  }
*/
  /**
   * Test: PutObject(): do put object using multi-threaded way in parallel.
   */
  public static void threadedPutObject() throws Exception {
    System.out.println("Test: threadedPutObject");
    Thread[] threads = new Thread[7];

    for (int i = 0; i < 7; i++) {
      threads[i] = new Thread(new PutObjectRunnable(client, bucketName,largeFile));
    }

    for (int i = 0; i < 7; i++) {
      threads[i].start();
    }

    // Waiting for threads to complete.
    for (int i = 0; i < 7; i++) {
      threads[i].join();
    }

    // All threads are completed.
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName).
   */
  public static void copyObject_test1() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);
    client.copyObject(bucketName, objectName, destBucketName);
    client.getObject(destBucketName, objectName, objectName + ".downloaded");
    Files.delete(Paths.get(objectName + ".downloaded"));

    client.removeObject(bucketName, objectName);
    client.removeObject(destBucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with ETag to match.
   */
  public static void copyObject_test2() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with Matching ETag (Negative Case)");

    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    CopyConditions copyConditions = new CopyConditions();
    copyConditions.setMatchETag("TestETag");

    try {
      client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    } catch (ErrorResponseException e) {
      ignore();
    }

    client.removeObject(bucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with ETag to match.
   */
  public static void copyObject_test3() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with Matching ETag (Positive Case)");

    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    ObjectStat stat = client.statObject(bucketName, objectName);
    CopyConditions copyConditions = new CopyConditions();
    copyConditions.setMatchETag(stat.etag());

    // File should be copied as ETag set in copyConditions matches object's ETag.
    client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    client.getObject(destBucketName, objectName, objectName + ".downloaded");
    Files.delete(Paths.get(objectName + ".downloaded"));

    client.removeObject(bucketName, objectName);
    client.removeObject(destBucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with ETag to not match.
   */
  public static void copyObject_test4() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with not matching ETag"
                       + " (Positive Case)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    CopyConditions copyConditions = new CopyConditions();
    copyConditions.setMatchETagNone("TestETag");

    // File should be copied as ETag set in copyConditions doesn't match object's ETag.
    client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    client.getObject(destBucketName, objectName, objectName + ".downloaded");
    Files.delete(Paths.get(objectName + ".downloaded"));

    client.removeObject(bucketName, objectName);
    client.removeObject(destBucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with ETag to not match.
   */
  public static void copyObject_test5() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with not matching ETag"
                       + " (Negative Case)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    ObjectStat stat = client.statObject(bucketName, objectName);
    CopyConditions copyConditions = new CopyConditions();
    copyConditions.setMatchETagNone(stat.etag());

    try {
      client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    } catch (ErrorResponseException e) {
      // File should not be copied as ETag set in copyConditions matches object's ETag.
      ignore();
    }

    client.removeObject(bucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with object modified after condition.
   */
  public static void copyObject_test6() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with modified after "
                       + "condition (Positive Case)");
    String objectName = createFile(3 * MB);
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    CopyConditions copyConditions = new CopyConditions();
    DateTime dateRepresentation = new DateTime(2015, Calendar.MAY, 3, 10, 10);

    copyConditions.setModified(dateRepresentation);

    // File should be copied as object was modified after the set date.
    client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    client.getObject(destBucketName, objectName, objectName + ".downloaded");
    Files.delete(Paths.get(objectName + ".downloaded"));

    client.removeObject(bucketName, objectName);
    client.removeObject(destBucketName, objectName);
    client.removeBucket(destBucketName);
  }

  /**
   * Test: copyObject(String bucketName, String objectName, String destBucketName,
   * CopyConditions copyConditions) with object modified after condition.
   */
  public static void copyObject_test7() throws Exception {
    System.out.println("Test: copyObject(String bucketName, String objectName, String destBucketName,"
                       + "CopyConditions copyConditions) with modified after"
                       + " condition (Negative Case)");
    String objectName = getRandomName();
    client.putObject(bucketName, objectName, smallFile);

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName);

    CopyConditions copyConditions = new CopyConditions();
    DateTime dateRepresentation = new DateTime(2015, Calendar.MAY, 3, 10, 10);

    copyConditions.setUnmodified(dateRepresentation);

    try {
      client.copyObject(bucketName, objectName, destBucketName, copyConditions);
    } catch (ErrorResponseException e) {
      // File should not be copied as object was modified after date set in copyConditions.
      if (!e.errorResponse().code().equals("PreconditionFailed")) {
        throw e;
      }
    }

    client.removeObject(bucketName, objectName);
    // Destination bucket is expected to be empty, otherwise it will trigger an exception.
    client.removeBucket(destBucketName);
  }

  /**
   * Test: setBucketNotification(String bucketName, NotificationConfiguration notificationConfiguration).
   */
  public static void setBucketNotification_test1() throws Exception {
    // This test requires 'MINIO_JAVA_TEST_TOPIC' and 'MINIO_JAVA_TEST_REGION' environment variables.
    String topic = System.getenv("MINIO_JAVA_TEST_TOPIC");
    String region = System.getenv("MINIO_JAVA_TEST_REGION");
    if (topic == null || topic.equals("") || region == null || region.equals("")) {
      // do not run functional test as required environment variables are missing.
      return;
    }

    System.out.println("Test: setBucketNotification(String bucketName, "
                       + "NotificationConfiguration notificationConfiguration)");

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName, region);

    NotificationConfiguration notificationConfiguration = new NotificationConfiguration();

    // Add a new topic configuration.
    List<TopicConfiguration> topicConfigurationList = notificationConfiguration.topicConfigurationList();
    TopicConfiguration topicConfiguration = new TopicConfiguration();
    topicConfiguration.setTopic(topic);

    List<EventType> eventList = new LinkedList<>();
    eventList.add(EventType.OBJECT_CREATED_PUT);
    eventList.add(EventType.OBJECT_CREATED_COPY);
    topicConfiguration.setEvents(eventList);

    Filter filter = new Filter();
    filter.setPrefixRule("images");
    filter.setSuffixRule("pg");
    topicConfiguration.setFilter(filter);

    topicConfigurationList.add(topicConfiguration);
    notificationConfiguration.setTopicConfigurationList(topicConfigurationList);

    client.setBucketNotification(destBucketName, notificationConfiguration);

    client.removeBucket(destBucketName);
  }

  /**
   * Test: getBucketNotification(String bucketName).
   */
  public static void getBucketNotification_test1() throws Exception {
    // This test requires 'MINIO_JAVA_TEST_TOPIC' and 'MINIO_JAVA_TEST_REGION' environment variables.
    String topic = System.getenv("MINIO_JAVA_TEST_TOPIC");
    String region = System.getenv("MINIO_JAVA_TEST_REGION");
    if (topic == null || topic.equals("") || region == null || region.equals("")) {
      // do not run functional test as required environment variables are missing.
      return;
    }

    System.out.println("Test: getBucketNotification(String bucketName)");

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName, region);

    NotificationConfiguration notificationConfiguration = new NotificationConfiguration();

    // Add a new topic configuration.
    List<TopicConfiguration> topicConfigurationList = notificationConfiguration.topicConfigurationList();
    TopicConfiguration topicConfiguration = new TopicConfiguration();
    topicConfiguration.setTopic(topic);

    List<EventType> eventList = new LinkedList<>();
    eventList.add(EventType.OBJECT_CREATED_PUT);
    topicConfiguration.setEvents(eventList);

    topicConfigurationList.add(topicConfiguration);
    notificationConfiguration.setTopicConfigurationList(topicConfigurationList);

    client.setBucketNotification(destBucketName, notificationConfiguration);
    String expectedResult = notificationConfiguration.toString();

    notificationConfiguration = client.getBucketNotification(destBucketName);

    topicConfigurationList = notificationConfiguration.topicConfigurationList();
    topicConfiguration = topicConfigurationList.get(0);
    topicConfiguration.setId(null);
    String result = notificationConfiguration.toString();

    if (!result.equals(expectedResult)) {
      System.out.println("FAILED. expected: " + expectedResult + ", got: " + result);
    }

    client.removeBucket(destBucketName);
  }


  /**
   * Test: removeAllBucketNotification(String bucketName).
   */
  public static void removeAllBucketNotification_test1() throws Exception {
    // This test requires 'MINIO_JAVA_TEST_TOPIC' and 'MINIO_JAVA_TEST_REGION' environment variables.
    String topic = System.getenv("MINIO_JAVA_TEST_TOPIC");
    String region = System.getenv("MINIO_JAVA_TEST_REGION");
    if (topic == null || topic.equals("") || region == null || region.equals("")) {
      // do not run functional test as required environment variables are missing.
      return;
    }

    System.out.println("Test: removeAllBucketNotification(String bucketName)");

    String destBucketName = getRandomName();
    client.makeBucket(destBucketName, region);

    NotificationConfiguration notificationConfiguration = new NotificationConfiguration();

    // Add a new topic configuration.
    List<TopicConfiguration> topicConfigurationList = notificationConfiguration.topicConfigurationList();
    TopicConfiguration topicConfiguration = new TopicConfiguration();
    topicConfiguration.setTopic(topic);

    List<EventType> eventList = new LinkedList<>();
    eventList.add(EventType.OBJECT_CREATED_PUT);
    eventList.add(EventType.OBJECT_CREATED_COPY);
    topicConfiguration.setEvents(eventList);

    Filter filter = new Filter();
    filter.setPrefixRule("images");
    filter.setSuffixRule("pg");
    topicConfiguration.setFilter(filter);

    topicConfigurationList.add(topicConfiguration);
    notificationConfiguration.setTopicConfigurationList(topicConfigurationList);

    client.setBucketNotification(destBucketName, notificationConfiguration);

    notificationConfiguration = new NotificationConfiguration();
    String expectedResult = notificationConfiguration.toString();

    client.removeAllBucketNotification(destBucketName);

    notificationConfiguration = client.getBucketNotification(destBucketName);
    String result = notificationConfiguration.toString();
    if (!result.equals(expectedResult)) {
      System.out.println("FAILED. expected: " + expectedResult + ", got: " + result);
    }

    client.removeBucket(destBucketName);
  }

  /**
   * runTests: runs as much as possible of test combinations.
   */
  public static void runTests() throws Exception {
    makeBucket_test1();
    if (endpoint.toLowerCase().contains("s3")) {
      makeBucket_test2();
      makeBucket_test3();
    }

    listBuckets_test();

    bucketExists_test();

    removeBucket_test();

    setup();

    putObject_test1();
    putObject_test2();
    putObject_test3();
    putObject_test4();
    putObject_test5();
    putObject_test6();
    //putObject_test7();
    //putObject_test8();
    //putObject_test9();
    //putObject_test10();

    statObject_test();
    getObject_test1();
    getObject_test2();
    getObject_test3();
    getObject_test4();
    getObject_test5();
    //getObject_test6();
    //getObject_test7();

    listObject_test1();
    listObject_test2();
    listObject_test3();
    listObject_test4();
    listObject_test5();
    listObject_test6();

    removeObject_test1();
    removeObject_test2();

    listIncompleteUploads_test1();
    listIncompleteUploads_test2();
    listIncompleteUploads_test3();

    removeIncompleteUploads_test();
    /*
    presignedGetObject_test1();
    presignedGetObject_test2();
    presignedGetObject_test3();

    presignedPutObject_test1();
    presignedPutObject_test2();

    presignedPostPolicy_test();
  */
    copyObject_test1();
    copyObject_test2();
    copyObject_test3();
    copyObject_test4();
    copyObject_test5();
    copyObject_test6();
    copyObject_test7();

    threadedPutObject();

    teardown();

    // notification tests requires 'MINIO_JAVA_TEST_TOPIC' and 'MINIO_JAVA_TEST_REGION' environment variables
    // to be set appropriately.
    setBucketNotification_test1();
    getBucketNotification_test1();
    removeAllBucketNotification_test1();
  }

  /**
   * runFastTests: runs a fast set of tests.
   */
  public static void runFastTests() throws Exception {
    makeBucket_test1();
    listBuckets_test();
    bucketExists_test();
    removeBucket_test();

    setup();

    putObject_test1();
    statObject_test();
    getObject_test1();
    listObject_test1();
    removeObject_test1();
    listIncompleteUploads_test1();
    removeIncompleteUploads_test();
   // presignedGetObject_test1();
   // presignedPutObject_test1();
   // presignedPostPolicy_test();
    copyObject_test1();

    teardown();
  }


  /**
   * main().
   */
  public static void main(String[] args) {
    if (args.length != 4) {
      System.out.println("usage: FunctionalTest <ENDPOINT> <ACCESSKEY> <SECRETKEY> <REGION>");
      System.exit(-1);
    }

    endpoint = args[0];
    accessKey = args[1];
    secretKey = args[2];
    region = args[3];
    try {
      client = new MinioClient(endpoint, accessKey, secretKey);
      FunctionalTest.runTests();
    } catch (Exception e) {
      e.printStackTrace();
      System.exit(-1);
    }
  }
}
