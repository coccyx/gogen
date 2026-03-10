from cors_utils import cors_response
from github_utils import get_github_user
from logger import setup_logger

logger = setup_logger(__name__)


def get_header(event, name):
    """Return an HTTP header value using case-insensitive lookup."""
    headers = event.get('headers') or {}
    target = name.lower()

    for key, value in headers.items():
        if key.lower() == target:
            return value

    return None


def get_authenticated_username(event):
    """
    Authenticate the GitHub token from the request and return the username.

    Returns:
        tuple[str | None, dict | None]: (username, error_response)
    """
    auth_header = get_header(event, 'Authorization')
    if not auth_header:
        logger.error("Authorization header not present")
        return None, cors_response(401, {'error': 'Authorization header not present'})

    user_info, error = get_github_user(auth_header)
    if error:
        logger.error(f"Failed to authenticate user: {error}")
        return None, cors_response(401, {'error': error})

    username = user_info.get('login')
    if not username:
        logger.error("Could not get username from GitHub")
        return None, cors_response(401, {'error': 'Could not get username from GitHub'})

    return username, None
