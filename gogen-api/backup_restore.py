import boto3
import json
from decimal import Decimal
import os
from botocore.config import Config

class DecimalEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Decimal):
            # Convert Decimal to float if it's a whole number, otherwise keep as string
            if obj % 1 == 0:
                return int(obj)
            return float(obj)
        return super(DecimalEncoder, self).default(obj)

def backup_table():
    """Backup DynamoDB table to a JSON file"""
    dynamodb = boto3.resource('dynamodb')
    table = dynamodb.Table('gogen')
    
    items = []
    scan_kwargs = {}
    done = False
    start_key = None
    
    while not done:
        if start_key:
            scan_kwargs['ExclusiveStartKey'] = start_key
        response = table.scan(**scan_kwargs)
        items.extend(response.get('Items', []))
        start_key = response.get('LastEvaluatedKey', None)
        done = start_key is None
    
    with open('table_backup.json', 'w') as f:
        json.dump(items, f, cls=DecimalEncoder, indent=2)
    print(f"Backed up {len(items)} items to table_backup.json")

def restore_table(local=True):
    """Restore DynamoDB table from JSON file"""
    if local:
        config = Config(
            connect_timeout=5,
            read_timeout=5,
            retries={'max_attempts': 3}
        )
        dynamodb = boto3.resource('dynamodb',
                                endpoint_url='http://localhost:8000',
                                region_name='us-east-1',
                                aws_access_key_id='DUMMYIDEXAMPLE',
                                aws_secret_access_key='DUMMYEXAMPLEKEY',
                                config=config)
    else:
        dynamodb = boto3.resource('dynamodb')
    
    table = dynamodb.Table('gogen')
    
    try:
        with open('table_backup.json', 'r') as f:
            items = json.load(f)
        
        with table.batch_writer() as batch:
            for item in items:
                # Convert numeric strings back to Decimal where appropriate
                processed_item = {}
                for key, value in item.items():
                    if isinstance(value, (int, float)):
                        processed_item[key] = Decimal(str(value))
                    else:
                        processed_item[key] = value
                batch.put_item(Item=processed_item)
        
        print(f"Restored {len(items)} items to {'local' if local else 'remote'} table")
    except Exception as e:
        print(f"Error restoring data: {str(e)}")

if __name__ == '__main__':
    # If running locally, make sure the table exists first
    if os.environ.get('LOCAL_DYNAMODB'):
        from create_local_table import create_local_table
        create_local_table()
    
    # By default, backup from remote and restore to local
    backup_table()
    restore_table(local=True) 