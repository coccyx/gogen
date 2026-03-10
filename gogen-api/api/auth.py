import json
from cors_utils import cors_response
from github_utils import exchange_code_for_token, get_github_user
from logger import setup_logger

logger = setup_logger(__name__)
logger.info('Loading function')


def respond(err, res=None):
    if err:
        return cors_response(400, {'error': str(err)})
    return cors_response(200, res)


def lambda_handler(event, context):
    """
    Handle GitHub OAuth code exchange.

    Receives: { "code": "...", "state": "..." }
    Returns: { "access_token": "...", "user": { "login": "...", "avatar_url": "...", "id": ... } }
    """
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

        # Validate required fields
        if 'code' not in body:
            logger.error("Missing 'code' in request body")
            return respond("Missing 'code' in request body")

        code = body['code']
        state = body.get('state')  # State is optional but recommended for CSRF protection

        logger.info(f"Processing OAuth code exchange (state: {state})")

        # Exchange code for access token
        token_data, error = exchange_code_for_token(code)
        if error:
            logger.error(f"Failed to exchange code for token: {error}")
            return respond(error)

        access_token = token_data['access_token']
        token_type = token_data.get('token_type', 'bearer')

        # Get user information
        auth_header = f"{token_type} {access_token}"
        user_info, error = get_github_user(auth_header)
        if error:
            logger.error(f"Failed to get user info: {error}")
            return respond(error)

        # Return token and user info
        response = {
            'access_token': access_token,
            'token_type': token_type,
            'user': {
                'login': user_info.get('login'),
                'avatar_url': user_info.get('avatar_url'),
                'id': user_info.get('id'),
                'name': user_info.get('name'),
                'email': user_info.get('email')
            }
        }

        logger.info(f"OAuth exchange successful for user: {user_info.get('login')}")
        return respond(None, response)

    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(str(e))
