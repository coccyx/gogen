import boto3
from botocore.config import Config

def create_table():
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

    table = dynamodb.create_table(
        TableName='gogen',
        KeySchema=[
            {
                'AttributeName': 'gogen',
                'KeyType': 'HASH'
            }
        ],
        AttributeDefinitions=[
            {
                'AttributeName': 'gogen',
                'AttributeType': 'S'
            }
        ],
        ProvisionedThroughput={
            'ReadCapacityUnits': 5,
            'WriteCapacityUnits': 5
        }
    )

    # Wait for the table to be created
    table.meta.client.get_waiter('table_exists').wait(TableName='gogen')
    print("Table created successfully!")

if __name__ == '__main__':
    try:
        create_table()
    except Exception as e:
        if 'ResourceInUseException' in str(e):
            print("Table already exists!")
        else:
            print(f"Error creating table: {str(e)}") 