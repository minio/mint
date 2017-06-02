import os
import io
import sys
import uuid
import urllib3
import certifi
import logging
import functools
import inspect
from faker import Factory

from random import choice
from string import ascii_uppercase
from datetime import datetime, timedelta

from minio import Minio, PostPolicy, CopyConditions
from minio.policy import Policy
from minio.error import (ResponseError, PreconditionFailed,
                         BucketAlreadyOwnedByYou, BucketAlreadyExists)
import logger as logger
from decorator import log_decorate


LOG_FILE = "logger.log"
logger = logger.create_logger(sys.argv[1] + "/" + LOG_FILE)
fake = Factory.create()
# Generate unique string
S3_ADDRESS = os.getenv('S3_ADDRESS')
ACCESS_KEY = os.getenv('ACCESS_KEY')
SECRET_KEY = os.getenv('SECRET_KEY') 
is_s3 = S3_ADDRESS.startswith("s3.amazonaws")


def generate_random_string(length=20):
    return ''.join(choice(ascii_uppercase) for i in range(length))


@log_decorate(logger)
def make_bucket_test(client, bucket_name):
    # Make a new bucket.
    try:
        is_s3 = client._endpoint_url.startswith("s3.amazonaws")
        if is_s3:
            try:
                client.make_bucket(bucket_name + '.unique',
                                   location='us-west-1')
                client.bucket_exists(bucket_name + '.unique')
            except BucketAlreadyOwnedByYou as err:
                pass
            except BucketAlreadyExists as err:
                pass
            except Exception as err:
                logger.error(err)
        else:
            client.make_bucket(bucket_name)
            found = client.bucket_exists(bucket_name)
    except Exception as err:
        logger.error(err)



@log_decorate(logger)
def list_buckets_test(client):
    try:
        # List all buckets.
        buckets = client.list_buckets()
        for bucket in buckets:
            _, _ = bucket.name, bucket.creation_date
    except Exception as err:
        logger.error(err)

@log_decorate(logger)
def remove_bucket_test(client,bucket_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        client.remove_bucket(bucket_name)
        found = client.bucket_exists(bucket_name)
        assert found == False
    except Exception as err:
        logger.error(err)

@log_decorate(logger)
def put_small_object_from_stream_test(client,bucket_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'

        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()

        object_name = uuid.uuid4().__str__()
        # Put a file
        file_stat = os.stat(testfile)
        with open(testfile, 'rb') as file_data:
            client.put_object(bucket_name, object_name, file_data,
                              file_stat.st_size)
        file_data.close()
        os.remove(testfile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)

@log_decorate(logger)
def put_large_object_from_stream_test(client,bucket_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True

        largefile = 'largefile'
        with open(largefile, 'wb') as file_data:
            for i in range(0, 1040857):
                file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        object_name = uuid.uuid4().__str__()
        # Put a file
        file_stat = os.stat(largefile)
        with open(largefile, 'rb') as file_data:
            client.put_object(bucket_name, object_name, file_data,
                              file_stat.st_size)
        file_data.close()
        os.remove(largefile)
    except Exception as err:
        logger.error(err)

@log_decorate(logger)
def put_small_object_from_file_test(client,bucket_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'

        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        object_name = uuid.uuid4().__str__()
        # Fput a file
        client.fput_object(bucket_name, object_name+'-f', testfile)
        if is_s3:
            client.fput_object(bucket_name, object_name+'-f', testfile,
                               metadata={'x-amz-storage-class': 'STANDARD_IA'})
        os.remove(testfile)

    except Exception as err:
        logger.error(err)

@log_decorate(logger)
def put_large_object_from_file_test(client,bucket_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        largefile = 'largefile'
        with open(largefile, 'wb') as file_data:
            for i in range(0, 1040857):
                file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        object_name = uuid.uuid4().__str__()
        # Put a file
        file_stat = os.stat(largefile)
        with open(largefile, 'rb') as file_data:
            client.put_object(bucket_name, object_name, file_data,
                              file_stat.st_size)
        file_data.close()
        os.remove(largefile)
    except Exception as err:
        logger.error(err)


def setup_client():
    client = Minio(S3_ADDRESS,
                   ACCESS_KEY,
                   SECRET_KEY)
    _http = urllib3.PoolManager(
        cert_reqs='CERT_REQUIRED',
        ca_certs=certifi.where()
    )
# Enable trace
# client.trace_on(sys.stderr)

    return client


def run_tests(client):
    bucket_name = generate_random_string()
    bucket_name = "bucket120"
    try:
        make_bucket_test(client, bucket_name)
        make_bucket_test(client, generate_random_string(65))
        list_buckets_test(client)
        put_small_object_from_stream_test(client, bucket_name)
        put_large_object_from_stream_test(client, bucket_name)
        put_small_object_from_file_test(client, bucket_name)
    except Exception as err:
        print("failing tests", err)
        pass
    finally:
        pass


def teardown():
    print("Ending minio-py functional tests")
    return


if __name__ == '__main__':
    print("running minio-python functional tests")
    client = setup_client()
    run_tests(client)
    teardown()
