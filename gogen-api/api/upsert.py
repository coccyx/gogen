import json
import http.client
from boto3.dynamodb.conditions import Key, Attr
from db_utils import get_dynamodb_client
from s3_utils import upload_config
from logger import setup_logger

logger = setup_logger(__name__)
logger.info('Loading function')


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': str(err) if err else json.dumps(res),
        'headers': {
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
        },
    }


def validate_github_token(token):
    """
    Validate the GitHub token by making a request to GitHub's API
    """
    headers = {
        'Authorization': token,
        'User-Agent': 'gogen lambda',
        'Content-Length': '0'
    }
    
    logger.debug("Attempting to validate GitHub token")
    conn = http.client.HTTPSConnection('api.github.com')
    conn.request("GET", "/user", None, headers)
    response = conn.getresponse()
    
    if response.status != 200:
        data = response.read().decode('utf-8')
        logger.error(f"GitHub token validation failed. Status: {response.status}, Reason: {response.reason}")
        logger.debug(f"GitHub API response: {data}")
        return False, f"Unable to authenticate user to GitHub, status: {response.status}, msg: {response.reason}"
    
    logger.info("GitHub token validation successful")
    return True, None


def lambda_handler(event, context):
    try:
        logger.debug(f"Received event: {json.dumps(event, indent=2)}")
        
        # Validate request body
        if 'body' not in event:
            logger.error("No request body provided")
            return respond("Request body is required")
            
        try:
            body = json.loads(event['body'])
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in request body: {str(e)}")
            return respond("Invalid JSON in request body")
            
        # Validate GitHub authorization
        if 'headers' not in event or 'Authorization' not in event['headers']:
            logger.error("Authorization header not present")
            return respond("Authorization header not present")
            
        # Validate GitHub token
        is_valid, error_msg = validate_github_token(event['headers']['Authorization'])
        if not is_valid:
            return respond(error_msg)
            
        # Validate and clean request body
        validated_body = {}
        for k, v in body.items():
            if v != "":
                validated_body[k] = v
                
        if not validated_body:
            logger.error("No valid fields in request body")
            return respond("No valid fields in request body")
        
        # Check if config is present in the request
        if 'config' in validated_body:
            config_content = validated_body['config']
            
            # Create S3 path in the format username/sample.yml
            if 'owner' in validated_body and 'name' in validated_body:
                s3_path = f"{validated_body['owner']}/{validated_body['name']}.yml"
                
                # Upload config to S3
                logger.info(f"Uploading config to S3 at path: {s3_path}")
                upload_success = upload_config(s3_path, config_content)
                
                if not upload_success:
                    logger.error(f"Failed to upload config to S3 at path: {s3_path}")
                    return respond("Failed to upload configuration to S3")
                
                # Remove config from DynamoDB item to save space
                # We'll store the S3 path instead
                validated_body.pop('config', None)
                
                # Add S3 path to DynamoDB item
                validated_body['s3Path'] = s3_path
                
                # Remove gistID if present (for migration)
                validated_body.pop('gistID', None)
            else:
                logger.error("Owner or name missing in request body")
                return respond("Owner and name are required fields")
        else:
            logger.warning("No config found in request body")
            
        logger.info(f"Processing upsert for item: {validated_body}")
        
        # Store in DynamoDB
        table = get_dynamodb_client().Table('gogen')
        logger.debug(f"Attempting to upsert item to DynamoDB: {validated_body}")
        
        response = table.put_item(
            Item=validated_body
        )
        
        logger.info(f"Successfully upserted item to DynamoDB")
        return respond(None, response)
        
    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(e) 