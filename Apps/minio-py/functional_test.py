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
                         BucketAlreadyOwnedByYou, BucketAlreadyExists, InvalidBucketError)
import logger as logger
from decorator import log_decorate


LOG_FILE = "output.log"
logger = logger.create_logger(sys.argv[1] + "/" + LOG_FILE)
fake = Factory.create()
# Generate unique string
S3_ADDRESS = os.getenv('S3_ADDRESS')
ACCESS_KEY = os.getenv('ACCESS_KEY')
SECRET_KEY = os.getenv('SECRET_KEY') 
S3_SECURE  = os.getenv('S3_SECURE') 
is_s3 = S3_ADDRESS.startswith("s3.amazonaws")
_http = None

def generate_random_string(length=20):
    return ''.join(choice(ascii_uppercase) for i in range(length)).lower()


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
                raise

        else:
            client.make_bucket(bucket_name)
            found = client.bucket_exists(bucket_name)
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def make_bucket_test2(client, bucket_name):

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
                raise

        else:
            client.make_bucket(bucket_name)
            found = client.bucket_exists(bucket_name)
    except InvalidBucketError as err:
            pass
    except Exception as err:
        logger.error(err)
        raise


@log_decorate(logger)
def list_buckets_test(client):
    try:
        # List all buckets.
        buckets = client.list_buckets()
        for bucket in buckets:
            _, _ = bucket.name, bucket.creation_date
    except Exception as err:
        logger.error(err)
        raise

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
        raise

@log_decorate(logger)
def put_small_object_from_stream_test(client,bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'

        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()

        # Put a file
        file_stat = os.stat(testfile)
        with open(testfile, 'rb') as file_data:
            client.put_object(bucket_name, object_name, file_data,
                              file_stat.st_size)
        file_data.close()
        os.remove(testfile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
        logger.info("OK")
    except Exception as err:
        logger.error(err)
        raise
@log_decorate(logger)
def put_large_object_from_stream_test(client,bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True

        largefile = 'largefile'
        with open(largefile, 'wb') as file_data:
            for i in range(0, 140857):
                file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        # Put a file
        file_stat = os.stat(largefile)
        with open(largefile, 'rb') as file_data:
            client.put_object(bucket_name, object_name, file_data,
                              file_stat.st_size)
        file_data.close()
        os.remove(largefile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def put_object(client, bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'

        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        file_stat = os.stat(testfile)

        # Fput a file
        client.fput_object(bucket_name, object_name, testfile)
        if is_s3:
            client.fput_object(bucket_name, object_name, testfile,
                               metadata={'x-amz-storage-class': 'STANDARD_IA'})
        os.remove(testfile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def remove_object(client, bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        client.remove_object(bucket_name, object_name)
        logger.info("??? OK " + bucket_name + " : " + object_name)
    except Exception as err:
        logger.info("removing object ::: " + bucket_name + " : " + object_name)
        logger.error(err)
        raise

@log_decorate(logger)
def put_small_object_from_file_test(client,bucket_name,object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'

        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        file_stat = os.stat(testfile)
        # Fput a file
        client.fput_object(bucket_name, object_name, testfile)
        if is_s3:
            client.fput_object(bucket_name, object_name, testfile,
                               metadata={'x-amz-storage-class': 'STANDARD_IA'})
        os.remove(testfile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def put_large_object_from_file_test(client,bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True, logger.error(bucket_name + " missing on server")
        largefile = 'largefile'
        with open(largefile, 'wb') as file_data:
            for i in range(0, 140857):
                file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        # Put a file
        file_stat = os.stat(largefile)
        if is_s3:
            client.fput_object(bucket_name, object_name, largefile,
                               metadata={'x-amz-storage-class': 'STANDARD_IA'})
        else:
            client.fput_object(bucket_name, object_name, largefile)

        os.remove(largefile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def copy_object_test(client,bucket_name,dest_object_name, src_object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'
        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        # Put a file
        file_stat = os.stat(testfile)
        client.fput_object(bucket_name, src_object_name, testfile)
        os.remove(testfile)
        stat = client.stat_object(bucket_name,src_object_name)
        assert stat.size == file_stat.st_size
        client.copy_object(bucket_name,dest_object_name,
                       '/'+bucket_name+'/'+src_object_name)
        copy_stat = client.stat_object(bucket_name,dest_object_name)
        assert copy_stat.size == stat.size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def copy_object_with_conditions_test(client,bucket_name, dest_object_name, src_object_name):
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
        client.fput_object(bucket_name, src_object_name, testfile)
        os.remove(testfile)
        stat = client.stat_object(bucket_name,src_object_name)
        assert stat.size == file_stat.st_size
        try:
            copy_conditions = CopyConditions()
            copy_conditions.set_match_etag('test-etag')
            client.copy_object(bucket_name, dest_object_name,
                               '/'+bucket_name+'/'+src_object_name,
                               copy_conditions)
            copy_stat = client.stat_object(bucket_name,dest_object_name)
            assert copy_stat.size == stat.size
        except PreconditionFailed as err:
            if err.message != 'At least one of the preconditions you specified did not hold.':
                logger.error(err)
        except Exception as err:
            logger.error(err)
            raise
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def stat_object_test(client,bucket_name, object_name):
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        testfile = 'testfile'
        with open(testfile, 'wb') as file_data:
            file_data.write(fake.text().encode('utf-8'))
        file_data.close()
        # Put a file
        client.fput_object(bucket_name,object_name,testfile)
        file_stat = os.stat(testfile)
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def get_object_test(client, bucket_name, object_name): 
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        # Get a full object
        object_data = client.get_object(bucket_name, object_name)
        # Save object stream to file
        with open("newfile", 'wb') as file_data:
            for data in object_data:
                file_data.write(data)
        file_data.close()
        
        file_stat = os.stat("newfile")
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
        os.remove("newfile")
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def get_partial_object_test(client, bucket_name, object_name): 
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        # Get a partial object
        data = client.get_partial_object(bucket_name, object_name, 5, 10)
        with open('my-testfile', 'wb') as file_data:
            for d in data:
                file_data.write(d)
        
        file_data.close()       
        file_stat = os.stat('my-testfile')

        assert file_stat.st_size == 10
        os.remove("my-testfile")
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def get_fobject_test(client, bucket_name, object_name): 
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
         # Get a full object locally.
        client.fget_object(bucket_name, object_name, "newfile_f")
        
        file_stat = os.stat("newfile_f")
        stat = client.stat_object(bucket_name,object_name)
        assert stat.size == file_stat.st_size
        os.remove("newfile_f")
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def list_objects_test(client, bucket_name): 
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        # List all object paths in bucket.
        objects = client.list_objects(bucket_name, recursive=True)
        for obj in objects:
            _, _, _, _, _, _ = obj.bucket_name, obj.object_name, \
                               obj.last_modified, \
                               obj.etag, obj.size, \
                               obj.content_type
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def list_objects_v2_test(client, bucket_name): 
    try:
        found = client.bucket_exists(bucket_name)
        assert found == True
        # List all object paths in bucket.
        objects = client.list_objects_v2(bucket_name, recursive=True)
        for obj in objects:
            _, _, _, _, _, _ = obj.bucket_name, obj.object_name, \
                               obj.last_modified, \
                               obj.etag, obj.size, \
                               obj.content_type
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def remove_objects_test(client, bucket_name): 
    try: 
        client.make_bucket(bucket_name)
        for i in range(10):
            put_small_object_from_file_test(client,bucket_name,"newobject" + str(i))
        for del_err in client.remove_objects(bucket_name, ["newobject" + str(i) for i in range(10)]):
            logger.error("Deletion Error: {}".format(del_err))
        client.remove_bucket(bucket_name)
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def presigned_get_object_url_test(client,bucket_name, object_name):
    try: 
        presigned_get_object_url = client.presigned_get_object(bucket_name, object_name)
        response = _http.urlopen('GET', presigned_get_object_url)
        if response.status != 200:
            raise ResponseError(response,
                                'GET',
                                bucket_name,
                                object_name).get_exception()
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def presigned_put_object_url_test(client,bucket_name, object_name):
    try: 
        presigned_put_object_url = client.presigned_put_object(bucket_name, object_name)
        value = fake.text().encode('utf-8')
        data = io.BytesIO(value).getvalue()
        response = _http.urlopen('PUT', presigned_put_object_url, body=data)
        if response.status != 200:
            raise ResponseError(response,
                                'PUT',
                                bucket_name,
                                object_name).get_exception()
        object_data = client.get_object(bucket_name, object_name)
        if object_data.read() != value:
            logger.error('Bytes not equal')
    except Exception as err:
        logger.error(err)
        raise

@log_decorate(logger)
def presigned_post_policy_test(client, bucket_name):
    try:
        # Post policy.
        policy = PostPolicy()
        policy.set_bucket_name(bucket_name)
        policy.set_key_startswith('objectPrefix/')

        expires_date = datetime.utcnow()+timedelta(days=10)
        policy.set_expires(expires_date)
        client.presigned_post_policy(policy)
    except Exception as err:
        logger.error(err)
        raise

def init_client():
    client = Minio(S3_ADDRESS,
                   ACCESS_KEY,
                   SECRET_KEY,
                   secure=S3_SECURE)
    global _http
    _http = urllib3.PoolManager(
        cert_reqs='CERT_REQUIRED',
        ca_certs=certifi.where()
    )
# Enable trace
# client.trace_on(sys.stderr)

    return client

def setup(client): 
    logger.info("setting up py client.....")
    bucket_name = generate_random_string().lower()
    object_name = uuid.uuid4().__str__().lower()
    make_bucket_test(client, bucket_name)
    put_object(client, bucket_name, object_name)
    return (bucket_name, object_name)

def run_tests(client):
    try:
        suffixes = ["","-small", "-large", "-fsmall", "-flarge", "-copy", "-copycond"]
        bucket_name, object_name = setup(client)
        make_bucket_test2(client, generate_random_string(65))
   
        list_buckets_test(client)
        put_small_object_from_stream_test(client, bucket_name, object_name + suffixes[1])
        put_large_object_from_stream_test(client, bucket_name, object_name + suffixes[2])
        put_small_object_from_file_test(client, bucket_name, object_name + suffixes[3])
        
        put_large_object_from_file_test(client, bucket_name, object_name + suffixes[4])
        
        copy_object_test(client,bucket_name,object_name + suffixes[5], object_name)
        copy_object_with_conditions_test(client,bucket_name, object_name + suffixes[6], object_name)
        stat_object_test(client, bucket_name, object_name)
        get_object_test(client, bucket_name, object_name)
        get_partial_object_test(client, bucket_name, object_name)
        get_fobject_test(client, bucket_name, object_name)
        presigned_get_object_url_test(client,bucket_name, object_name)
        presigned_put_object_url_test(client,bucket_name, object_name)
        presigned_post_policy_test(client, bucket_name)
        list_objects_test(client,bucket_name)
        list_objects_v2_test(client,bucket_name)
        
        remove_objects_test(client,bucket_name + "rmv")
        teardown(client,bucket_name, object_name, suffixes)
    except Exception as err:
        logger.error("failing tests", err)
        raise
    finally:
        pass

def teardown(client,bucket_name, object_name, suffixes):

    for suffix in suffixes: 
        client.remove_object(bucket_name, object_name + suffix)
    client.remove_bucket(bucket_name)
    return


if __name__ == '__main__':
    client = init_client()
    run_tests(client)
   