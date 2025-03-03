import json
from boto3.dynamodb.conditions import Key, Attr
from db_utils import get_dynamodb_client

print('Loading function')


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': str(err) if err else json.dumps(res),
        'headers': {
            'Content-Type': 'application/json',
        },
    }


def lambda_handler(event, context):
    print(f"Received event: {json.dumps(event, indent=2)}")
    q = event['queryStringParameters']['q']
    print(f"Query: {q}")
    table = get_dynamodb_client().Table('gogen')
    response = table.scan(
        ProjectionExpression="gogen, description",
        FilterExpression=Attr("gogen").contains(q) | Attr("description").contains(q)
    )
    return respond(None, response) 