import os
import boto3
from botocore.config import Config
from logger import setup_logger

logger = setup_logger(__name__)

def get_dynamodb_client():
    """
    Get a DynamoDB client - uses local endpoint if running locally
    """
    if os.environ.get('AWS_SAM_LOCAL'):
        # Use the container name as hostname when running in SAM Lambda
        logger.info("Configuring DynamoDB client for local development")
        config = Config(
            connect_timeout=5,
            read_timeout=5,
            retries={'max_attempts': 2},
            max_pool_connections=10,
            tcp_keepalive=True
        )
        logger.debug(f"Using config: {config}")
        
        client = boto3.resource('dynamodb', 
                            endpoint_url='http://dynamodb-local:8000',
                            region_name='us-east-1',
                            aws_access_key_id='DUMMYIDEXAMPLE',
                            aws_secret_access_key='DUMMYEXAMPLEKEY',
                            config=config)
        # Test the connection
        try:
            logger.info("Testing DynamoDB connection...")
            tables = client.meta.client.list_tables()
            logger.info(f"Connection successful. Available tables: {tables['TableNames']}")
        except Exception as e:
            logger.error(f"Failed to connect to local DynamoDB: {str(e)}")
            raise
        return client
    return boto3.resource('dynamodb') 