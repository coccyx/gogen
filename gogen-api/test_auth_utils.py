import sys
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent / 'api'))

import auth_utils  # noqa: E402


class AuthUtilsTest(unittest.TestCase):
    def test_get_header_is_case_insensitive(self):
        event = {
            'headers': {
                'authorization': 'token abc123',
            }
        }

        self.assertEqual(auth_utils.get_header(event, 'Authorization'), 'token abc123')

    @patch('auth_utils.get_github_user')
    def test_get_authenticated_username_returns_username(self, mock_get_github_user):
        mock_get_github_user.return_value = ({'login': 'clint'}, None)
        event = {'headers': {'Authorization': 'token abc123'}}

        username, error = auth_utils.get_authenticated_username(event)

        self.assertEqual(username, 'clint')
        self.assertIsNone(error)
        mock_get_github_user.assert_called_once_with('token abc123')

    @patch('auth_utils.get_github_user')
    def test_get_authenticated_username_returns_401_without_authorization_header(self, mock_get_github_user):
        username, error = auth_utils.get_authenticated_username({'headers': {}})

        self.assertIsNone(username)
        self.assertEqual(error['statusCode'], '401')
        self.assertIn('Authorization header not present', error['body'])
        mock_get_github_user.assert_not_called()


if __name__ == '__main__':
    unittest.main()
