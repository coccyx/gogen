import json
import boto3
from botocore.config import Config
from decimal import Decimal
import os

def create_local_table():
    """Create a table in local DynamoDB using the schema from table_schema.json"""
    # Load schema from file
    with open('table_schema.json', 'r') as f:
        schema = json.load(f)

    # Configure local DynamoDB client
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

    # Create table using schema
    try:
        table = dynamodb.create_table(
            TableName=schema['TableName'],
            KeySchema=schema['KeySchema'],
            AttributeDefinitions=schema['AttributeDefinitions'],
            ProvisionedThroughput=schema['ProvisionedThroughput']
        )
        
        # Add GSIs if they exist in the schema
        if 'GlobalSecondaryIndexes' in schema:
            # GSIs are added during table creation, so we don't need to do anything here
            pass
            
        # Wait for the table to be created
        table.meta.client.get_waiter('table_exists').wait(TableName=schema['TableName'])
        print('Table created successfully!')
    except Exception as e:
        if 'ResourceInUseException' in str(e):
            print('Table already exists!')
        else:
            print(f'Error creating table: {str(e)}')

if __name__ == '__main__':
    create_local_table() 