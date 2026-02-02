import json
from db_utils import get_dynamodb_client, get_table_name
from s3_utils import delete_config
from cors_utils import cors_response
from github_utils import get_github_user
from logger import setup_logger

logger = setup_logger(__name__)
logger.info('Loading function')


def respond(err, res=None, status_code=400):
    if err:
        return cors_response(status_code, {'error': str(err)})
    return cors_response(200, res)


def lambda_handler(event, context):
    """
    Handle configuration deletion.

    Validates GitHub token, verifies ownership, then deletes from S3 and DynamoDB.
    """
    # Handle OPTIONS requests for CORS
    if event.get('httpMethod') == 'OPTIONS':
        return cors_response(200, {'message': 'OK'})

    try:
        logger.debug(f"Received event: {json.dumps(event, indent=2)}")

        # Validate GitHub authorization
        if 'headers' not in event or 'Authorization' not in event['headers']:
            logger.error("Authorization header not present")
            return respond("Authorization header not present", status_code=401)

        # Get user information from GitHub token
        auth_header = event['headers']['Authorization']
        user_info, error = get_github_user(auth_header)
        if error:
            logger.error(f"Failed to authenticate user: {error}")
            return respond(error, status_code=401)

        username = user_info.get('login')
        if not username:
            logger.error("Could not get username from GitHub")
            return respond("Could not get username from GitHub", status_code=401)

        # Extract config name from path
        path_params = event.get('pathParameters', {})
        proxy_path = path_params.get('proxy', '')

        if not proxy_path:
            logger.error("No configuration path provided")
            return respond("No configuration path provided")

        # Parse owner and config name from path (format: owner/configname)
        path_parts = proxy_path.split('/')
        if len(path_parts) < 2:
            logger.error(f"Invalid path format: {proxy_path}")
            return respond("Invalid path format. Expected: owner/configname")

        owner = path_parts[0]
        config_name = '/'.join(path_parts[1:])
        full_config_name = f"{owner}/{config_name}"

        logger.info(f"User {username} attempting to delete config: {full_config_name}")

        # Verify ownership
        if owner != username:
            logger.error(f"User {username} attempted to delete config owned by {owner}")
            return respond("You can only delete your own configurations", status_code=403)

        # Delete from S3
        s3_path = f"{full_config_name}.yml"
        logger.info(f"Deleting from S3: {s3_path}")
        s3_deleted = delete_config(s3_path)
        if not s3_deleted:
            logger.warning(f"Failed to delete S3 object: {s3_path} (may not exist)")

        # Delete from DynamoDB
        table = get_dynamodb_client().Table(get_table_name())
        logger.info(f"Deleting from DynamoDB: {full_config_name}")

        try:
            response = table.delete_item(
                Key={
                    'gogen': full_config_name
                },
                ReturnValues='ALL_OLD'
            )

            if 'Attributes' not in response:
                logger.warning(f"Configuration not found in DynamoDB: {full_config_name}")
                return respond("Configuration not found", status_code=404)

            logger.info(f"Successfully deleted configuration: {full_config_name}")
            return respond(None, {
                'message': f"Successfully deleted {full_config_name}",
                'deleted': response.get('Attributes', {})
            })

        except Exception as e:
            logger.error(f"Error deleting from DynamoDB: {str(e)}")
            return respond(f"Error deleting configuration: {str(e)}")

    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(str(e))
