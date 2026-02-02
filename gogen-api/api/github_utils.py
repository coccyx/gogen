import os
import json
import http.client
from logger import setup_logger

logger = setup_logger(__name__)


def validate_github_token(token):
    """
    Validate the GitHub token by making a request to GitHub's API.

    Args:
        token: GitHub authorization header value (e.g., "token xxx" or "Bearer xxx")

    Returns:
        tuple: (is_valid: bool, error_message: str or None)
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


def get_github_user(token):
    """
    Get GitHub user information using the provided token.

    Args:
        token: GitHub authorization header value (e.g., "token xxx" or "Bearer xxx")

    Returns:
        tuple: (user_info: dict or None, error_message: str or None)
    """
    headers = {
        'Authorization': token,
        'User-Agent': 'gogen lambda',
        'Accept': 'application/json'
    }

    logger.debug("Fetching GitHub user info")
    conn = http.client.HTTPSConnection('api.github.com')
    conn.request("GET", "/user", None, headers)
    response = conn.getresponse()
    data = response.read().decode('utf-8')

    if response.status != 200:
        logger.error(f"Failed to get GitHub user. Status: {response.status}, Reason: {response.reason}")
        logger.debug(f"GitHub API response: {data}")
        return None, f"Failed to get GitHub user info, status: {response.status}, msg: {response.reason}"

    try:
        user_info = json.loads(data)
        logger.info(f"Successfully fetched GitHub user: {user_info.get('login')}")
        return user_info, None
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse GitHub user response: {str(e)}")
        return None, "Failed to parse GitHub user response"


def exchange_code_for_token(code):
    """
    Exchange an OAuth authorization code for an access token.

    Args:
        code: OAuth authorization code from GitHub

    Returns:
        tuple: (token_data: dict or None, error_message: str or None)
    """
    client_id = os.environ.get('GITHUB_OAUTH_CLIENT_ID')
    client_secret = os.environ.get('GITHUB_OAUTH_CLIENT_SECRET')

    if not client_id or not client_secret:
        logger.error("GitHub OAuth credentials not configured")
        return None, "GitHub OAuth credentials not configured"

    headers = {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        'User-Agent': 'gogen lambda'
    }

    body = json.dumps({
        'client_id': client_id,
        'client_secret': client_secret,
        'code': code
    })

    logger.debug("Exchanging OAuth code for access token")
    conn = http.client.HTTPSConnection('github.com')
    conn.request("POST", "/login/oauth/access_token", body, headers)
    response = conn.getresponse()
    data = response.read().decode('utf-8')

    if response.status != 200:
        logger.error(f"OAuth token exchange failed. Status: {response.status}, Reason: {response.reason}")
        logger.debug(f"GitHub OAuth response: {data}")
        return None, f"OAuth token exchange failed, status: {response.status}"

    try:
        token_data = json.loads(data)

        if 'error' in token_data:
            error_desc = token_data.get('error_description', token_data.get('error'))
            logger.error(f"OAuth error: {error_desc}")
            return None, f"OAuth error: {error_desc}"

        if 'access_token' not in token_data:
            logger.error("No access_token in OAuth response")
            return None, "No access_token in OAuth response"

        logger.info("Successfully exchanged OAuth code for access token")
        return token_data, None
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse OAuth response: {str(e)}")
        return None, "Failed to parse OAuth response"
