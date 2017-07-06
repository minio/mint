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

  def print_title(title)
    # Prints the title for the test
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


  def create_bucket(s3Resource)
    # Tests if a bucket can be created for s3
    # client instance, "s3Resource".
    bucket_name = SecureRandom.hex(6)
    bucket = s3Resource.create_bucket(bucket: bucket_name)

    return bucket if bucket.exists?
    raise "Create bucket failure"
  rescue => e
    puts "ERROR: ", e
  end


  def delete_bucket(bucket)
    # Tests if a bucket object, "bucket", can be deleted/removed
    if bucket.exists?
      bucket.objects.each do |obj|
        obj.delete obj.key
      end
      bucket.delete
    end
  rescue => e
    puts "ERROR: Failed to delete bucket:", e
  end


  def list_buckets_test(s3Resource, s3Client)
    # Tests if existing bucket objects for
    # s3 client instances (s3Resource and s3Client)
    # can be looped through (list functionality)
    # It also logs/pritns the total number of bucket objects
    # found for each client instance.
    print_title "List buckets Test"
    begin
      i = j = 0
      s3Resource.buckets.limit(1000).each do |b|
        i += 1
      end
      print_log("Buckets (Resource) found:", i)
      print_logn("- Success!")
      s3Client.list_buckets.buckets.each do |b|
        j += 1
      end
      print_log("Buckets (Client) found:", j)
      print_logn("- Success!")
      state = "PASS"
    rescue => e
      state = "FAIL"
    end
    # Clean-up
    print_status(state, e)
  end



  def make_remove_bucket_test(s3Resource)
    # Tests if a bucket can be made/created.
    # If successful, it also tests, if the
    # same created bucket can be removed/deleted.
    print_title "Make/Remove Bucket Test"
    begin
      print_log("Making a bucket")
      bucket = create_bucket(s3Resource)
      if bucket.exists?
        state = "PASS"
      else
        state = "FAIL"
      end
      print_logn("- Success!")
      print_log("Deleting the bucket")
      delete_bucket(bucket)
      if !bucket.exists?
        state = "PASS"
      else
        state = "FAIL"
      end
      print_logn("- Success!")
    rescue => e
        state = "FAIL"
    end
    print_status(state, e)
  end


  def upload_object_test(s3Resource, data_dir)
    # Tests if an file object can be uploaded
    # to s3 using s3 client, "s3Resource" at
    # the location, "data_dir"
    # It cleans up after the test is done.
    file = data_dir + '/datafile-1-MB'
    print_title "Upload Object Test"
    begin
      name = File.basename(file)
      # Create the object to upload
      bucket = create_bucket(s3Resource)
      obj = bucket.object(name)
      # Upload it
      print_log("Uploading a 1MB file")
      obj.upload_file(file)
      print_logn("- Success!")
      state = "PASS"
    rescue => e
      state = "FAIL"
    end
    # Clean-up
    print_log "Clean-up"
    delete_bucket(bucket)
    print_logn("- Success!")
    print_status(state, e)
  end

  def download_object_test(s3Resource,data_dir)
    # Tests if a file object can be uplaoded.
    # To achieve this goal, it first downloads
    #  the file object, and then uploads it.
    # It cleans up after the test is done.
    file = data_dir + '/datafile-1-MB'
    destination = '/tmp' + '/datafile-1-MB'
    print_title "Download Object Test"
    begin
      name = File.basename(file)
      bucket = create_bucket(s3Resource)
      obj = bucket.object(name)
      print_log("First uploading a 1MB file")
      obj.upload_file(file)
      print_logn("- Success!")
      # Get the item's content and save it to a file
      print_log("Downloading the same object into your local /tmp directory")
      obj.get(response_target: destination)
      print_logn("- Success!")
      state = "PASS"
    rescue => e
      state = "FAIL"
    end
    print_log("Clean-up")
    delete_bucket(bucket)
    system("rm #{destination}")
    print_logn("- Success!")
    print_status(state, e)
  end
end

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
# variable "ENABLE_HTTPS" is turned on
endpoint = enable_https == "1" ? 'https://' + endpoint : 'http://' + endpoint

# Create s3 client instances, "s3Resource" and "s3Client"
s3Resource = Aws::S3::Resource.new(region: region, endpoint: endpoint, access_key_id: access_key_id,
secret_access_key: secret_access_key, force_path_style: true)
s3Client = Aws::S3::Client.new(region: region, endpoint: endpoint, access_key_id: access_key_id,
secret_access_key: secret_access_key, force_path_style: true)

# Create the test class instance and call the tests
aws = AWS_SDK_Ruby_Test.new
aws.list_buckets_test(s3Resource, s3Client)
aws.make_remove_bucket_test(s3Resource)
aws.upload_object_test(s3Resource, data_dir)
aws.download_object_test(s3Resource, data_dir)
