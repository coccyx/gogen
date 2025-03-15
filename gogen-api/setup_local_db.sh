#!/bin/bash

# Make sure we're in the right directory
cd "$(dirname "$0")"

# Check if virtual environment is activated, if not activate it
if [ -z "$VIRTUAL_ENV" ]; then
    VENV_PATH="/home/clint/local/src/gogen/.pyvenv"
    if [ -d "$VENV_PATH" ]; then
        echo "Activating Python virtual environment..."
        source "$VENV_PATH/bin/activate"
    else
        echo "Warning: Python virtual environment not found at $VENV_PATH"
        echo "Consider creating it with: python3 -m venv $VENV_PATH"
    fi
fi

# Get the schema
echo "Getting table schema..."
python get_schema.py

# Create the table in local DynamoDB using the schema
echo "Creating table in local DynamoDB using schema..."
LOCAL_DYNAMODB=true python create_local_table.py

echo "Local DynamoDB setup complete!" 