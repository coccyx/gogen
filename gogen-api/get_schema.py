import boto3
import json

def get_table_schema():
    dynamodb = boto3.client('dynamodb')
    
    try:
        response = dynamodb.describe_table(TableName='gogen')
        schema = {
            'TableName': response['Table']['TableName'],
            'KeySchema': response['Table']['KeySchema'],
            'AttributeDefinitions': response['Table']['AttributeDefinitions'],
            'ProvisionedThroughput': {
                'ReadCapacityUnits': response['Table']['ProvisionedThroughput']['ReadCapacityUnits'],
                'WriteCapacityUnits': response['Table']['ProvisionedThroughput']['WriteCapacityUnits']
            }
        }
        
        # Add GSIs if they exist
        if 'GlobalSecondaryIndexes' in response['Table']:
            schema['GlobalSecondaryIndexes'] = response['Table']['GlobalSecondaryIndexes']
        
        # Save schema to file
        with open('table_schema.json', 'w') as f:
            json.dump(schema, f, indent=2)
            print("Schema saved to table_schema.json")
            
        return schema
    
    except Exception as e:
        print(f"Error getting schema: {str(e)}")
        return None

if __name__ == '__main__':
    get_table_schema() 