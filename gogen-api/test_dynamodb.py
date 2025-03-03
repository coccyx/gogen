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

def test_connection():
    logger.info("Testing DynamoDB connection...")
    
    config = Config(
        connect_timeout=10,
        read_timeout=10,
        retries={'max_attempts': 3}
    )
    
    try:
        # Create DynamoDB client
        dynamodb = boto3.resource('dynamodb',
                                endpoint_url='http://localhost:8000',
                                region_name='us-east-1',
                                aws_access_key_id='DUMMYIDEXAMPLE',
                                aws_secret_access_key='DUMMYEXAMPLEKEY',
                                config=config)
        
        # List tables
        tables = list(dynamodb.tables.all())
        logger.info(f"Found tables: {[t.name for t in tables]}")
        
        # Test gogen table
        table = dynamodb.Table('gogen')
        
        # Count items
        scan_result = table.scan(Select='COUNT')
        item_count = scan_result['Count']
        logger.info(f"Found {item_count} items in gogen table")
        
        # Get a sample of items
        if item_count > 0:
            sample = table.scan(Limit=3)
            logger.info("Sample items:")
            for item in sample['Items']:
                logger.info(item)
                
        return True
        
    except Exception as e:
        logger.error(f"Error testing DynamoDB: {str(e)}", exc_info=True)
        return False

if __name__ == '__main__':
    test_connection() 