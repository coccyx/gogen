import json
from boto3.dynamodb.conditions import Key, Attr
from db_utils import get_dynamodb_client
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


def lambda_handler(event, context):
    try:
        logger.debug(f"Received event: {json.dumps(event, indent=2)}")
        
        # Get search query from parameters
        query_params = event.get('queryStringParameters', {})
        q = query_params.get('q')
        
        if not q:
            logger.warning("No search query provided")
            return respond("Search query parameter 'q' is required")
            
        logger.info(f"Processing search query: {q}")
        
        table = get_dynamodb_client().Table('gogen')
        
        # Use pagination to handle large datasets
        items = []
        last_evaluated_key = None
        page_count = 0
        
        while True:
            scan_kwargs = {
                'ProjectionExpression': "gogen, description",
                'FilterExpression': Attr("gogen").contains(q) | Attr("description").contains(q),
                'Limit': 20
            }
            
            if last_evaluated_key:
                scan_kwargs['ExclusiveStartKey'] = last_evaluated_key
            
            logger.debug(f"Scanning DynamoDB table with kwargs: {scan_kwargs}")
            response = table.scan(**scan_kwargs)
            
            page_items = response.get('Items', [])
            items.extend(page_items)
            page_count += 1
            logger.debug(f"Retrieved {len(page_items)} matching items on page {page_count}")
            
            last_evaluated_key = response.get('LastEvaluatedKey')
            if not last_evaluated_key:
                break
            logger.debug(f"More pages available, continuing scan with key: {last_evaluated_key}")
        
        logger.info(f"Successfully retrieved {len(items)} total matching items across {page_count} pages")
        return respond(None, {'Items': items})
        
    except Exception as e:
        logger.error(f"Error in lambda_handler: {str(e)}", exc_info=True)
        return respond(e) 