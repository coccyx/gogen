#!/bin/bash 

# Check if virtual environment exists and activate it
VENV_PATH="/home/clint/local/src/gogen/.pyvenv"
if [ -d "$VENV_PATH" ]; then
    echo "Activating Python virtual environment..."
    source "$VENV_PATH/bin/activate"
else
    echo "Warning: Python virtual environment not found at $VENV_PATH"
    echo "Consider creating it with: python3 -m venv $VENV_PATH"
fi

# Start Docker containers
echo "Starting Docker containers..."
docker compose up -d
sleep 5

# Setup local database
echo "Setting up local database..."
. ./setup_local_db.sh

# Build and start SAM application
echo "Building and starting SAM application..."
sam build
sam local start-api --host 0.0.0.0 --port 4000 --warm-containers EAGER --docker-network lambda-local

# Trap Ctrl+C and call cleanup
cleanup() {
    echo "Cleaning up..."
    docker compose down
    exit 0
}

trap cleanup INT

cleanup