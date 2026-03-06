import json
from db_utils import get_dynamodb_client, get_table_name
from s3_utils import upload_config
from cors_utils import cors_response
from auth_utils import get_authenticated_username
from logger import setup_logger

logger = setup_logger(__name__)
logger.info('Loading function')


def respond(err, res=None):
    if err:
        return cors_response(400, str(err))
    return cors_response(200, res)


def lambda_handler(event, context):
    # Handle OPTIONS requests for CORS
    if event.get('httpMethod') == 'OPTIONS':
        return cors_response(200, {'message': 'OK'})
        
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
            
        username, auth_error = get_authenticated_username(event)
        if auth_error:
            return auth_error
            
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
            
            if 'name' in validated_body:
                validated_body['owner'] = username
                s3_path = f"{username}/{validated_body['name']}.yml"
                
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

                # Set the primary key (gogen = owner/name)
                validated_body['gogen'] = f"{username}/{validated_body['name']}"

                # Remove gistID if present (for migration)
                validated_body.pop('gistID', None)
            else:
                logger.error("Name missing in request body")
                return respond("Name is a required field")
        else:
            logger.warning("No config found in request body")
            
        logger.info(f"Processing upsert for item: {validated_body}")
        
        # Store in DynamoDB
        table = get_dynamodb_client().Table(get_table_name())
        logger.debug(f"Attempting to upsert item to DynamoDB: {validated_body}")
        
        response = table.put_item(
            Item=validated_body
        )
        
        logger.info(f"Successfully upserted item to DynamoDB")
        return respond(None, response)
        
    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(e) 
