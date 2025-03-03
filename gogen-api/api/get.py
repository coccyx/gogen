import json
import decimal
from boto3.dynamodb.conditions import Key, Attr
from db_utils import get_dynamodb_client

print('Loading function')


def decimal_default(obj):
    if isinstance(obj, decimal.Decimal):
        return float(obj)
    raise TypeError


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': str(err) if err else json.dumps(res, default=decimal_default),
        'headers': {
            'Content-Type': 'application/json',
        },
    }


def lambda_handler(event, context):
    print(f"Received event: {json.dumps(event, indent=2)}")
    q = event['pathParameters']['proxy']
    print(f"Query: {q}")

    table = get_dynamodb_client().Table('gogen')
    response = table.get_item(Key={"gogen": q})

    if 'Item' not in response:
        return {
            'statusCode': '404',
            'body': f'Could not find Gogen: {q}',
        }
    if 'gogen' not in response["Item"]:
        return {
            'statusCode': '404',
            'body': f'Could not find Gogen: {q}',
        }
    return respond(None, response)
