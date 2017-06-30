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

region = ENV['SERVER_REGION'] ||= 'SERVER_REGION not set'
endpoint =  ENV['SERVER_ENDPOINT'] ||= 'SERVER_ENDPOINT'
access_key_id = ENV['ACCESS_KEY'] ||= 'ACCESS_KEY is not set'
secret_access_key = ENV['SECRET_KEY'] ||= 'SECRET_KEY is not set'
data_dir = ENV['DATA_DIR'] ||= 'DATA_DIR is not set'
enable_https = ENV['ENABLE_HTTPS']


if enable_https == "1"
    endpoint = 'https://' + endpoint
else
    endpoint = 'http://' + endpoint
end

# Set up AWS Client
client = Aws::S3::Resource.new(region: region, endpoint: endpoint, access_key_id: access_key_id, 
secret_access_key: secret_access_key, force_path_style: true)

# Test list buckets #1
begin
    client.buckets.limit(1000).each do |b|
  	    puts "#{b.name}"
	end
	puts "List Buckets Test Pass"
rescue
	puts "Test List Buckets # 1 Fails."	
end
 
# Test making a bucket #1
bucket_exists = false
bucket_name = SecureRandom.hex(6)
begin 
	client.create_bucket(bucket: bucket_name)
	puts "Bucket Create Test Pass"
rescue
	puts "Bucket Create #1 Fails"
end

# Test bucket exists #1
begin
	resp = client.bucket(bucket_name).exists?
	if resp == true 
	    puts "Bucket Exists Test Pass"
	end
	rescue
	    puts "Bucket Exists Test Fails"	 
end

# Uploading an object to a bucket
file = data_dir + '/datafile-1-MB'
# Get just the file name
begin
	name = File.basename(file)
	# Create the object to upload
	obj = client.bucket(bucket_name).object(name)
	# Upload it      
	obj.upload_file(file)
	puts "Uploading Object Test Pass"
rescue
	puts "Uploading Object to Bucket Failed"
end

# Download an object
begin
	obj = client.bucket(bucket_name).object('small.file')
	# Get the item's content and save it to a file
	obj.get(response_target: '/tmp/my-small.file')
	puts "Downloading Object from Bucket Pass"
rescue
	puts "Downloading Object from Bucket Failed"
end

# Delete bucket
begin
	bucket_name_temp =   SecureRandom.hex(6)
	client.create_bucket(bucket: bucket_name_temp)
	resp = client.bucket(bucket_name).exists?
	 
	if resp == true 	 
        client.delete_bucket(bucket: bucket_name_temp)
        puts "Bucket Delete Test Pass"
	end 
rescue
    puts "Bucket Delete Test Fails"
end
