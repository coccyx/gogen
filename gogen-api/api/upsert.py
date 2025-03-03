import json
import http.client
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
    body = json.loads(event['body'])
    headers = { }
    if 'Authorization' not in event['headers']:
        return respond(Exception("Authorization header not present"))
    headers['Authorization'] = event['headers']['Authorization']
    headers['User-Agent'] = 'gogen lambda'
    headers['Content-Length'] = 0
    conn = http.client.HTTPSConnection('api.github.com')
    conn.request("GET", "/user", None, headers)
    response = conn.getresponse()
    if response.status != 200:
        data = response.read().decode('utf-8')
        return respond(Exception(f"Unable to authenticate user to GitHub, status: {response.status}, msg: {response.reason}, data: {data}"))
    validatedbody = { }
    for k, v in body.items():
        if v != "":
            validatedbody[k] = v
    print(f"Item: {validatedbody}")
    table = get_dynamodb_client().Table('gogen')
    response = table.put_item(
        Item=validatedbody
    )
    return respond(None, response) 