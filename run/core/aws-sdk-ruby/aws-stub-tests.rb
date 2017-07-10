#!/usr/bin/env ruby
#
#  Mint (C) 2017 Minio, Inc.
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

require 'aws-sdk'
require 'securerandom'
require 'colorize'

class AWS_SDK_Ruby_Test
  #
  # Helper methods
  #
  def print_title(title)
    # Prints the title of the test
    puts '=================================================='
    msg = "\n*** " + title + "\n"
    print msg.blue
  end

  def print_log(log_msg, arg='')
    # Prints a progress log message for
    # the on-going test WITHOUT a new line.
    # It accepts an arg to print out at the end
    # of the progress message
    msg = "\t" + log_msg + "%s"
    printf(msg.light_black, arg)
  end

  def print_logn(log_msg, arg='')
    # Prints a progress log message for
    # the on-going test WITH a new line.
    # It accepts an arg to print out at the end
    # of the progress message
    msg = "\t" + log_msg + "%s" + "\n"
    printf(msg.light_black, arg)
  end

  def print_status(result, e='')
    # Prints result/status of the test, as "PASS" or "FAIL".
    # It adds the captured error message
    # if the result/status is a "FAIL".
    e = e.nil? ? nil.to_s : "ERROR: " + e.to_s + "\n"
    msg = "*** " + result + "\n" + e
    if result == "PASS"
      printf(msg.green)
    else
      printf(msg.red)
    end
  end

  def cleanUp(s3, buckets)
    # Removes objects and the bucket
    # if bucket exists
    print_log "Clean-up"
    buckets.each do |b|
      if bucketExists?(s3, b)
        removeObjects(s3, b)
        removeBucket(s3, b)
      end
    end
    print_logn("- Success!")
  rescue => e
    puts "\nERROR: Failed to clean-up buckets " + b, e
  end

  #
  # API command methods
  #
  def makeBucket(s3, bucket_name)
    # Creates a bucket, "bucket_name"
    # on S3 client , "s3".
    bucket = s3.create_bucket(bucket: bucket_name)
    return bucket if bucket.exists?
  rescue => e
    puts "\nERROR: Failed to create bucket: " + bucket_name, e
  end

  def removeBucket(s3, bucket_name)
    # Deletes/removes bucket, "bucket_name" on S3 client, "s3"
    s3.bucket(bucket_name).delete
  rescue => e
    puts "\nERROR: Failed to delete bucket: " + bucket_name, e
  end

  def putObject(s3, bucket_name, file)
    # Creates "file" (full path) in bucket, "bucket_name",
    # on S3 client, "s3"
    file_name = File.basename(file)
    s3.bucket(bucket_name).object(file_name).upload_file(file)
  rescue => e
    puts "\nERROR: Failed to create file: " + file_name, e
  end

  def getObject(s3, bucket_name, file, destination)
    # Gets/Downloads file, "file",
    # from bucket, "bucket_name", of S3 client, "s3"
    file_name = File.basename(file)
    dest = File.join(destination, file_name)
    s3.bucket(bucket_name).object(file_name).get(response_target: dest)
  rescue => e
    puts "\nERROR: Failed to download file: " + file_name, e
  end

  def copyObject(s3, source_bucket_name, target_bucket_name, source_file_name, target_file_name="")
    # Copies file, "file_name", from source bucket,
    # "source_bucket_name", to target bucket,
    # "target_bucket_name", on S3 client, "s3"
    target_file_name = source_file_name if target_file_name.empty?
    source = s3.bucket(source_bucket_name)
    target = s3.bucket(target_bucket_name)
    source_obj = source.object(source_file_name)
    target_obj = target.object(target_file_name)
    source_obj.copy_to(target_obj)
  rescue => e
    puts "\nERROR: Failed to copy file: " + file_name, e
  end

  def removeObject(s3, bucket_name, file_name)
    # Deletes file, "file_name", in bucket,
    # "bucket_name", on S3 client, "s3".
    # If file, "file_name" does not exist,
    # it quitely returns without any error message
    s3.bucket(bucket_name).object(file_name).delete
  rescue => e
    puts "\nERROR: Failed to delete file: " + file_name, e
  end

  def removeObjects(s3, bucket_name)
    # Deletes all files in bucket, "bucket_name"
    # on S3 client, "s3"
    file_name = ""
    s3.bucket(bucket_name).objects.each do |obj|
      file_name = obj.key
      obj.delete obj.key
    end
  rescue => e
    puts "\nERROR: Failed to clean-up bucket: " +
        bucket_name + ", file: " + file_name, e
  end

  def listBuckets(s3)
    # Returns an array of bucket names on S3 client, "s3"
    bucket_list = []
    s3.buckets.each do |b|
      bucket_list.push(b.name)
    end
    return bucket_list
  rescue => e
    puts "\nERROR: Failed to get the list of buckets", e
  end

  def listObjects(s3, bucket_name)
    # Returns an array of object/file names
    # in bucket, "bucket_name", on S3 client, "s3"
    object_list = []
    s3.bucket(bucket_name).objects.each do |obj|
      object_list.push(obj.key)
    end
    return object_list
  rescue => e
    puts "\nERROR: Failed to get the list of files in bucket " +
           bucket_name, e
  end

  def statObject(s3, bucket_name, file_name)
    # Returns true/false if file, "file_name" is found
    # in bucket, "bucket_name", on S3 client, "s3"
    s3.bucket(bucket_name).objects.each do |obj|
      return true if file_name == obj.key
    end
    return false
  rescue => e
    puts "\nERROR: Failed to get stat for " +
          file_name + " in " + bucket_name, e
  end

  def bucketExists?(s3, bucket_name)
    # Returns true if bucket, "bucket_name", exists,
    # false otherwise
    return s3.bucket(bucket_name).exists?
  rescue => e
    puts "\nERROR: Failed to check if bucket, " +
           bucket_name + ", exists", e
  end

  #
  # Test case methods
  #
  def listBucketsTest(s3, bucket_list)
    # Tests listBuckets api command and prints out
    # the total number of buckets found
    print_title "List Buckets Test"
    begin
      prev_total_buckets = listBuckets(s3).length
      print_log("Buckets found:", prev_total_buckets.to_s)
      print_logn("- Success!")

      new_buckets = bucket_list.length
      print_log("Making " + new_buckets.to_s + " new buckets")
      bucket_list.each do |b|
        makeBucket(s3, b)
        print_logn("- Success!")
      end
      new_total_buckets = prev_total_buckets + new_buckets
      print_log("Total buckets found now:", new_total_buckets.to_s)
 
      if new_total_buckets == prev_total_buckets + new_buckets
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Expected total number of buckets and actual number do not match"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, bucket_list)
    print_status(state, e)
  end

  def makeBucketTest(s3, bucket_name)
    # Tests makeBucket api commands.
    print_title "Make Bucket Test"
    begin
      print_log("Making a bucket")
      makeBucket(s3, bucket_name)

      if bucketExists?(s3, bucket_name)
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Bucket expected to be created does not exist"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def bucketExistsNegativeTest(s3, bucket_name)
    # Tests bucketExists api commands.
    print_title "Bucket Exists Test"
    begin
      print_log("Checking a non-existing bucket")
      if !bucketExists?(s3, bucket_name)
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "bucketExists? api command failed
          to return 'false' for non-existing bucket"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def removeBucketTest(s3, bucket_name)
    # Tests removeBucket api commands.
    print_title "Remove Bucket Test"
    begin
      print_log("Making a bucket")
      makeBucket(s3, bucket_name)
      print_logn("- Success!")

      print_log("Deleting the bucket")
      removeBucket(s3, bucket_name)

      if !bucketExists?(s3, bucket_name)
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Bucket expected to be removed still exists"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def putObjectTest(s3, bucket_name, file)
    # Tests putObject api command
    # by uploading a file
    print_title "Upload and Remove Object Test"
    begin
      print_log("Making a bucket")
      makeBucket(s3, bucket_name)
      print_logn("- Success!")

      print_log("Uploading file")
      putObject(s3, bucket_name, file)

      if statObject(s3, bucket_name, File.basename(file))
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Status for the created object returned 'false'"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def removeObjectTest(s3, bucket_name, file)
    # Tests removeObject api command
    # by uploading and removing a file
    print_title "Upload and Remove Object Test"
    begin
      print_log("Making a bucket")
      makeBucket(s3, bucket_name)
      print_logn("- Success!")

      print_log("Uploading file")
      putObject(s3, bucket_name, file)
      print_logn("- Success!")

      print_log("Removing file")
      removeObject(s3, bucket_name, File.basename(file))

      if !statObject(s3, bucket_name, File.basename(file))
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Status for the removed object returned 'true'"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def getObjectTest(s3, bucket_name, file, destination)
    # Tests getObject api command
    print_title "Download Object Test"
    begin
      file_name = File.basename(file)

      print_log("Making a bucket")
      makeBucket(s3, bucket_name)
      print_logn("- Success!")

      print_log("Uploading file: ", file_name)
      putObject(s3, bucket_name, file)
      print_logn("- Success!")

      print_log("Downloading it into destination directory: ", destination)
      getObject(s3, bucket_name, file, destination)

      if system("ls -l #{destination} > /dev/null")
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Downloaded object does not exist at " + destination
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def listObjectsTest(s3, bucket_name, file_list)
    # Tests listObjects api command and prints out
    # the total number of files/objects found
    print_title "List Objects Test"
    begin
      print_log("Making a bucket")
      makeBucket(s3, bucket_name)
      print_logn("- Success!")

      # Put all objects into the bucket
      file_list.each do |f|
        putObject(s3, bucket_name, f)
      end

      # Total number of files uploaded
      expected_no = file_list.length
      # Actual number is what api returns
      actual_no = listObjects(s3, bucket_name).length
      print_logn("Files/Objects expected: ", expected_no)
      print_log("Files/Objects found: ", actual_no)

      # Compare expected and actual values
      if expected_no == actual_no
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Expected and actual number of listed files/objects do not match!"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [bucket_name])
    print_status(state, e)
  end

  def copyObjectTest(s3, source_bucket_name, target_bucket_name, data_dir, source_file_name, target_file_name="")
    # Tests copyObject api command
    # Target file name parameter, "target_file_name", is optional.
    # It is assumed to be the source file name if not provided
    print_title "Copy Object Test"
    begin
      target_file_name = source_file_name if target_file_name.empty?
      print_log("Making source bucket: ", source_bucket_name)
      bucket = makeBucket(s3, source_bucket_name)
      print_logn("- Success!")

      print_log("Making target bucket: ", target_bucket_name)
      bucket = makeBucket(s3, target_bucket_name)
      print_logn("- Success!")

      print_logn("Uploading file: ", source_file_name)
      print_log("... into source bucket: ", source_bucket_name)
      putObject(s3, source_bucket_name, File.join(data_dir, source_file_name))
      print_logn("- Success!")

      print_logn("Copying file: ", source_file_name)
      print_logn("... from source bucket: ", source_bucket_name)
      print_logn("... as file: ", target_file_name)
      print_log("... into target bucket: ", target_bucket_name)
      copyObject(s3, source_bucket_name, target_bucket_name, source_file_name, target_file_name)

      # Check if copy worked fine
      if statObject(s3, target_bucket_name, target_file_name)
        print_logn("- Success!")
        state = "PASS"
      else
        print_logn("- Failure!")
        e = "Copied file could not be found in the expected location"
        state = "FAIL"
      end
    rescue => e
      state = "FAIL"
    end
    cleanUp(s3, [source_bucket_name, target_bucket_name])
    print_status(state, e)
  end

end

# MAIN CODE
# Set variables necessary to create an s3 client instance.
# Get them from the environment variables

# Region information, eg. "us-east-1"
region = ENV['SERVER_REGION'] ||= 'SERVER_REGION is not set'
# Minio server, eg. "play.minio.io:9000"
endpoint =  ENV['SERVER_ENDPOINT'] ||= 'SERVER_ENDPOINT is not set'
access_key_id = ENV['ACCESS_KEY'] ||= 'ACESS_KEY is not set'
secret_access_key = ENV['SECRET_KEY'] ||= 'SECRET_KEY is not set'
# The location where the bucket and file
# objects are going to be created.
data_dir = ENV['MINT_DATA_DIR'] ||= 'MINT_DATA_DIR is not set'
# "1/0" value to decide if "HTTPS"
# needs to be used on or not.
enable_https = ENV['ENABLE_HTTPS']
# Add "https://" to "endpoint" if environment
# variable "ENABLE_HTTPS" is set to 1
endpoint = enable_https == "1" ? 'https://' + endpoint : 'http://' + endpoint
# Create s3 client instances, "s3Resource" and "s3Client"
s3Resource = Aws::S3::Resource.new(region: region, endpoint: endpoint, access_key_id: access_key_id,
secret_access_key: secret_access_key, force_path_style: true)

# Create test Class instance and call the tests
aws = AWS_SDK_Ruby_Test.new
bucket_name1 = SecureRandom.hex(6)
bucket_name2 = SecureRandom.hex(6)
bucket_list = [bucket_name1, bucket_name2]
file_name1 = 'datafile-1-MB'
file_new_name = 'datafile-1-MB-copy'
file_name_list = ['datafile-1-MB', 'datafile-5-MB', 'datafile-6-MB']
# Add data_dir in front of each file name in file_name_list
file_list = file_name_list.map{ |f| File.join(data_dir, f)}
destination = '/tmp'

aws.listBucketsTest(s3Resource, bucket_list)
aws.listObjectsTest(s3Resource, bucket_name1, file_list)
aws.makeBucketTest(s3Resource, bucket_name1)
aws.bucketExistsNegativeTest(s3Resource, bucket_name1)
aws.removeBucketTest(s3Resource, bucket_name1)
aws.putObjectTest(s3Resource, bucket_name1, File.join(data_dir, file_name1))
aws.removeObjectTest(s3Resource, bucket_name1, File.join(data_dir, file_name1))
aws.getObjectTest(s3Resource, bucket_name1, File.join(data_dir, file_name1), destination)
aws.copyObjectTest(s3Resource, bucket_name1, bucket_name2, data_dir, file_name1)
aws.copyObjectTest(s3Resource, bucket_name1, bucket_name2, data_dir, file_name1, file_new_name)
