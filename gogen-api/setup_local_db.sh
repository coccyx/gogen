#!/bin/bash

# Make sure we're in the right directory
cd "$(dirname "$0")"

# Get the schema
echo "Getting table schema..."
python get_schema.py

# Backup and restore data
echo "Backing up data and restoring to local DynamoDB..."
LOCAL_DYNAMODB=true python backup_restore.py

echo "Local DynamoDB setup complete!" 