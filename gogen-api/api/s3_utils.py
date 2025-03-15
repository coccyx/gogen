import os
import boto3
from botocore.config import Config
from logger import setup_logger

logger = setup_logger(__name__)

def get_s3_client():
    """
    Get an S3 client - uses local endpoint if running locally
    """
    if os.environ.get('AWS_SAM_LOCAL'):
        # Use the container name as hostname when running in SAM Lambda
        logger.info("Configuring S3 client for local development")
        config = Config(
            connect_timeout=5,
            read_timeout=5,
            retries={'max_attempts': 2},
            max_pool_connections=10,
            tcp_keepalive=True
        )
        logger.debug(f"Using config: {config}")
        
        client = boto3.resource('s3', 
                            endpoint_url='http://minio:9000',  # Use container name in docker network
                            region_name='us-east-1',
                            aws_access_key_id='minioadmin',
                            aws_secret_access_key='minioadmin',
                            config=config)
        # Test the connection
        try:
            logger.info("Testing S3 connection...")
            buckets = list(client.buckets.all())
            logger.info(f"Connection successful. Available buckets: {[b.name for b in buckets]}")
        except Exception as e:
            logger.error(f"Failed to connect to local S3: {str(e)}")
            raise
        return client
    return boto3.resource('s3')

def get_config_bucket():
    """
    Get the gogen-configs bucket
    """
    s3 = get_s3_client()
    return s3.Bucket('gogen-configs')

def upload_config(config_name, content):
    """
    Upload a config file to the gogen-configs bucket
    
    Args:
        config_name (str): Name of the config file
        content (str): Content of the config file
    
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        bucket = get_config_bucket()
        bucket.put_object(Key=config_name, Body=content)
        logger.info(f"Successfully uploaded config {config_name} to gogen-configs bucket")
        return True
    except Exception as e:
        logger.error(f"Error uploading config {config_name}: {str(e)}")
        return False

def download_config(config_name):
    """
    Download a config file from the gogen-configs bucket
    
    Args:
        config_name (str): Name of the config file
    
    Returns:
        str: Content of the config file, or None if not found
    """
    try:
        bucket = get_config_bucket()
        obj = bucket.Object(config_name)
        content = obj.get()['Body'].read().decode('utf-8')
        logger.info(f"Successfully downloaded config {config_name} from gogen-configs bucket")
        return content
    except Exception as e:
        logger.error(f"Error downloading config {config_name}: {str(e)}")
        return None

def list_configs():
    """
    List all configs in the gogen-configs bucket
    
    Returns:
        list: List of config names
    """
    try:
        bucket = get_config_bucket()
        configs = [obj.key for obj in bucket.objects.all()]
        logger.info(f"Found {len(configs)} configs in gogen-configs bucket")
        return configs
    except Exception as e:
        logger.error(f"Error listing configs: {str(e)}")
        return []

def delete_config(config_name):
    """
    Delete a config file from the gogen-configs bucket
    
    Args:
        config_name (str): Name of the config file
    
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        bucket = get_config_bucket()
        bucket.Object(config_name).delete()
        logger.info(f"Successfully deleted config {config_name} from gogen-configs bucket")
        return True
    except Exception as e:
        logger.error(f"Error deleting config {config_name}: {str(e)}")
        return False 