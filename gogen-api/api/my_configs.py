import json
from boto3.dynamodb.conditions import Attr
from db_utils import get_dynamodb_client, get_table_name
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
    List configurations owned by the authenticated user.

    Validates GitHub token, gets username, then scans DynamoDB for matching configs.
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

        logger.info(f"Fetching configurations for user: {username}")

        # Scan DynamoDB for configurations owned by this user
        table = get_dynamodb_client().Table(get_table_name())

        # Filter for configs where owner matches the username
        filter_expression = Attr('owner').eq(username)

        items = []
        scan_kwargs = {
            'FilterExpression': filter_expression
        }

        # Handle pagination
        done = False
        start_key = None
        while not done:
            if start_key:
                scan_kwargs['ExclusiveStartKey'] = start_key

            response = table.scan(**scan_kwargs)
            items.extend(response.get('Items', []))

            start_key = response.get('LastEvaluatedKey')
            done = start_key is None

        logger.info(f"Found {len(items)} configurations for user {username}")

        # Return the list of configurations
        return respond(None, {
            'Items': items,
            'Count': len(items),
            'owner': username
        })

    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(str(e))
