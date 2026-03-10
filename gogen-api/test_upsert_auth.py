import json
import sys
import unittest
from pathlib import Path
from types import ModuleType
from unittest.mock import Mock, patch

sys.path.insert(0, str(Path(__file__).resolve().parent / 'api'))

boto3_module = ModuleType('boto3')
boto3_module.resource = Mock()
sys.modules.setdefault('boto3', boto3_module)

botocore_module = ModuleType('botocore')
botocore_config_module = ModuleType('botocore.config')


class Config:
    def __init__(self, *args, **kwargs):
        self.args = args
        self.kwargs = kwargs


botocore_config_module.Config = Config
sys.modules.setdefault('botocore', botocore_module)
sys.modules.setdefault('botocore.config', botocore_config_module)

import upsert  # noqa: E402


class UpsertAuthTest(unittest.TestCase):
    @patch('upsert.get_table_name', return_value='gogen')
    @patch('upsert.get_dynamodb_client')
    @patch('upsert.upload_config', return_value=True)
    @patch('upsert.get_authenticated_username', return_value=('actual-user', None))
    def test_upsert_uses_authenticated_username_for_owner(
        self,
        mock_get_authenticated_username,
        mock_upload_config,
        mock_get_dynamodb_client,
        _mock_get_table_name,
    ):
        table = Mock()
        table.put_item.return_value = {'ResponseMetadata': {'HTTPStatusCode': 200}}
        mock_get_dynamodb_client.return_value.Table.return_value = table

        event = {
            'headers': {'authorization': 'token abc123'},
            'body': json.dumps({
                'owner': 'spoofed-user',
                'name': 'sample',
                'description': 'demo',
                'config': 'global: {}'
            })
        }

        response = upsert.lambda_handler(event, None)

        self.assertEqual(response['statusCode'], '200')
        mock_get_authenticated_username.assert_called_once_with(event)
        mock_upload_config.assert_called_once_with('actual-user/sample.yml', 'global: {}')
        table.put_item.assert_called_once()

        stored_item = table.put_item.call_args.kwargs['Item']
        self.assertEqual(stored_item['owner'], 'actual-user')
        self.assertEqual(stored_item['gogen'], 'actual-user/sample')
        self.assertEqual(stored_item['s3Path'], 'actual-user/sample.yml')

    @patch('upsert.get_authenticated_username')
    def test_upsert_returns_auth_error_response(self, mock_get_authenticated_username):
        mock_get_authenticated_username.return_value = (
            None,
            {'statusCode': '401', 'body': json.dumps({'error': 'Authorization header not present'})}
        )

        response = upsert.lambda_handler({'headers': {}, 'body': '{}'}, None)

        self.assertEqual(response['statusCode'], '401')
        self.assertIn('Authorization header not present', response['body'])


if __name__ == '__main__':
    unittest.main()
