import boto3
from botocore.config import Config
import logging
import sys

# Set up logging to stdout
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s'))
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.addHandler(handler)
# Remove any existing handlers to avoid duplicate logs
logger.propagate = False

def test_s3_connection():
    """
    Test connection to local MinIO S3 server
    """
    logger.info("Testing S3 connection...")
    
    config = Config(
        connect_timeout=10,
        read_timeout=10,
        retries={'max_attempts': 3}
    )
    
    try:
        # Create S3 client
        s3 = boto3.resource('s3',
                          endpoint_url='http://localhost:9000',
                          region_name='us-east-1',
                          aws_access_key_id='minioadmin',
                          aws_secret_access_key='minioadmin',
                          config=config)
        
        # List buckets
        buckets = list(s3.buckets.all())
        logger.info(f"Found buckets: {[b.name for b in buckets]}")
        
        # Test gogen-configs bucket
        bucket = s3.Bucket('gogen-configs')
        
        # Upload a test file
        test_content = "This is a test file for the gogen-configs bucket"
        bucket.put_object(Key='test-config.txt', Body=test_content)
        logger.info("Successfully uploaded test file to gogen-configs bucket")
        
        # List objects in bucket
        logger.info("Objects in gogen-configs bucket:")
        for obj in bucket.objects.all():
            logger.info(f"- {obj.key} (size: {obj.size} bytes)")
        
        # Download the test file
        obj = bucket.Object('test-config.txt')
        content = obj.get()['Body'].read().decode('utf-8')
        logger.info(f"Downloaded content: {content}")
        
        return True
        
    except Exception as e:
        logger.error(f"Error testing S3: {str(e)}", exc_info=True)
        return False

if __name__ == "__main__":
    test_s3_connection() 