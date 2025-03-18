import os
import json
import decimal
from typing import Any, Dict, Optional, Union

def decimal_default(obj):
    """Helper function to convert Decimal to float for JSON serialization."""
    if isinstance(obj, decimal.Decimal):
        return float(obj)
    raise TypeError(f"Object of type {type(obj)} is not JSON serializable")

def get_cors_headers() -> Dict[str, str]:
    """Get the CORS headers based on the environment."""
    env = os.environ.get('ENVIRONMENT', 'dev')
    origin = 'https://gogen.io' if env == 'prod' else 'https://staging.gogen.io'
    
    return {
        'Access-Control-Allow-Origin': origin,
        'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type,Authorization,X-Requested-With',
        'Access-Control-Allow-Credentials': 'true',
        'Content-Type': 'application/json'
    }

def cors_response(status_code: Union[int, str], body: Optional[Any] = None, additional_headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
    """
    Create a response with CORS headers.
    
    Args:
        status_code: HTTP status code
        body: Response body (will be JSON serialized if not a string)
        additional_headers: Additional headers to include in the response
        
    Returns:
        Dict containing the complete response with CORS headers
    """
    headers = get_cors_headers()
    if additional_headers:
        headers.update(additional_headers)
        
    if isinstance(body, str):
        response_body = body
    else:
        response_body = json.dumps(body, default=decimal_default) if body is not None else ''
        
    return {
        'statusCode': str(status_code),
        'headers': headers,
        'body': response_body
    } 