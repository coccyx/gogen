import json
import decimal
import urllib.request
import urllib.error
from boto3.dynamodb.conditions import Key, Attr
from db_utils import get_dynamodb_client
from s3_utils import download_config
from logger import setup_logger

logger = setup_logger(__name__)
logger.info('Loading function')


def decimal_default(obj):
    if isinstance(obj, decimal.Decimal):
        return float(obj)
    raise TypeError


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': str(err) if err else json.dumps(res, default=decimal_default),
        'headers': {
            'Content-Type': 'application/json',
        },
    }


def fetch_gist_content(gist_id):
    try:
        # Use GitHub API to get gist content
        api_url = f'https://api.github.com/gists/{gist_id}'
        logger.info(f"Fetching gist from GitHub API: {api_url}")
        
        with urllib.request.urlopen(api_url) as response:
            gist_data = json.loads(response.read().decode('utf-8'))
            logger.info(f"Successfully fetched gist data. Status: {response.status}")
            
            # Get the first file's content
            if not gist_data.get('files'):
                logger.error("No files found in gist")
                return None
                
            # Get the first file's content
            first_file = next(iter(gist_data['files'].values()))
            content = first_file.get('content')
            
            if not content:
                logger.error("No content found in gist file")
                return None
                
            logger.info(f"Successfully extracted content. Length: {len(content)} bytes")
            return content
                
    except urllib.error.URLError as e:
        logger.error(f"Error fetching gist: {str(e)}")
        if hasattr(e, 'code'):
            logger.error(f"HTTP Error Code: {e.code}")
        if hasattr(e, 'reason'):
            logger.error(f"Error Reason: {e.reason}")
        if hasattr(e, 'headers'):
            logger.debug(f"Error Response Headers: {dict(e.headers)}")
        return None
    except Exception as e:
        logger.error(f"Unexpected error while fetching gist: {str(e)}")
        return None


def lambda_handler(event, context):
    logger.debug(f"Received event: {json.dumps(event, indent=2)}")
    q = event['pathParameters']['proxy']
    logger.debug(f"Query: {q}")

    table = get_dynamodb_client().Table('gogen')
    response = table.get_item(Key={"gogen": q})

    if 'Item' not in response:
        logger.error(f"No item found for query: {q}")
        return {
            'statusCode': '404',
            'body': f'Could not find Gogen: {q}',
        }
    
    item = response['Item']
    if 'gogen' not in item:
        logger.error(f"Item found but missing 'gogen' key for query: {q}")
        return {
            'statusCode': '404',
            'body': f'Could not find Gogen: {q}',
        }

    # Try to fetch the configuration from S3 first
    if 's3Path' in item:
        logger.debug(f"Found s3Path: {item['s3Path']} for query: {q}")
        config_content = download_config(item['s3Path'])
        if config_content:
            item['config'] = config_content
            logger.debug(f"Successfully added config content from S3 for query: {q}")
        else:
            logger.error(f"Failed to fetch config content from S3 for query: {q}")
            return {
                'statusCode': '500',
                'body': f'Failed to fetch configuration from S3 for: {q}',
            }
    # For backward compatibility, try to fetch from GitHub gist if s3Path is not present
    elif 'gistID' in item:
        logger.warning(f"Using legacy gistID: {item['gistID']} for query: {q}. This will be deprecated.")
        config_content = fetch_gist_content(item['gistID'])
        if config_content:
            item['config'] = config_content
            logger.debug(f"Successfully added config content from GitHub gist for query: {q}")
        else:
            logger.error(f"Failed to fetch config content from GitHub gist for query: {q}")
            return {
                'statusCode': '500',
                'body': f'Failed to fetch configuration from GitHub gist for: {q}',
            }
    else:
        logger.error(f"No s3Path or gistID found in item for query: {q}")
        return {
            'statusCode': '500',
            'body': f'Configuration {q} does not have a valid storage location.',
        }
    
    response['Item'] = item
    return respond(None, response)
