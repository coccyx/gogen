#!/bin/bash 

# Determine script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Set GOGEN_HOME environment variable
export GOGEN_HOME="$PROJECT_ROOT"
echo "Setting GOGEN_HOME to $GOGEN_HOME"

# Check if virtual environment exists and activate it
VENV_PATH="$PROJECT_ROOT/.pyvenv"
if [ -d "$VENV_PATH" ]; then
    echo "Activating Python virtual environment..."
    source "$VENV_PATH/bin/activate"
else
    echo "Warning: Python virtual environment not found at $VENV_PATH"
    echo "Consider creating it with: python3 -m venv $VENV_PATH"
fi

# Start Docker containers
echo "Starting Docker containers..."
cd "$SCRIPT_DIR"
docker compose up -d
sleep 5

# Setup local database
echo "Setting up local database..."
. "$SCRIPT_DIR/setup_local_db.sh"

# Function to run test gogen commands
run_test_commands() {
    echo "Waiting for SAM local to start..."
    sleep 10  # Give SAM local some time to start

    echo "Running test gogen commands to validate API..."
    GOGEN_APIURL=http://localhost:4000 gogen -c "$PROJECT_ROOT/examples/tutorial/tutorial1.yml" push tutorial1
    GOGEN_APIURL=http://localhost:4000 gogen -c coccyx/tutorial1 config
    
    echo "Test commands completed."
}

# Build and start SAM application
echo "Building and starting SAM application..."
cd "$SCRIPT_DIR"
sam build

# Start test commands in background
run_test_commands &
TEST_COMMANDS_PID=$!

# Start SAM local in foreground
sam local start-api --host 0.0.0.0 --port 4000 --warm-containers EAGER --docker-network lambda-local

# Trap Ctrl+C and call cleanup
cleanup() {
    echo "Cleaning up..."
    # Kill the test commands process if it's still running
    if ps -p $TEST_COMMANDS_PID > /dev/null; then
        kill $TEST_COMMANDS_PID
    fi
    cd "$SCRIPT_DIR"
    docker compose down
    exit 0
}

trap cleanup INT

cleanup