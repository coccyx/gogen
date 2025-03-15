#!/bin/bash
set -e

# Determine script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Configuration
LAMBDA_DIR="$SCRIPT_DIR/api"
BUILD_DIR="$SCRIPT_DIR/build"
REGION="us-east-1"  # Change this to your AWS region
RUNTIME="python3.13"

# Use environment variable if set, otherwise use the default value
ROLE_ARN=${LAMBDA_ROLE_ARN:-$ROLE_ARN}

# Check if virtual environment exists and activate it
VENV_PATH="$PROJECT_ROOT/.pyvenv"
if [ -d "$VENV_PATH" ]; then
    echo "Activating Python virtual environment..."
    source "$VENV_PATH/bin/activate"
else
    echo "Python virtual environment not found at $VENV_PATH"
    echo "Setting up virtual environment..."
    
    # Check if setup_venv.sh exists and run it
    if [ -f "$PROJECT_ROOT/setup_venv.sh" ]; then
        echo "Running setup_venv.sh..."
        (cd "$PROJECT_ROOT" && bash setup_venv.sh)
        source "$VENV_PATH/bin/activate"
    else
        echo "Creating virtual environment manually..."
        python3 -m venv "$VENV_PATH"
        source "$VENV_PATH/bin/activate"
        pip install --upgrade pip
        
        # Install requirements if they exist
        if [ -f "$SCRIPT_DIR/requirements.txt" ]; then
            echo "Installing requirements from $SCRIPT_DIR/requirements.txt..."
            pip install -r "$SCRIPT_DIR/requirements.txt"
        else
            echo "Installing boto3 and botocore..."
            pip install boto3 botocore
        fi
        
        # Install AWS CLI if needed
        if ! command -v aws &> /dev/null; then
            echo "Installing AWS CLI..."
            pip install awscli
        fi
    fi
fi

# Create build directory if it doesn't exist
mkdir -p $BUILD_DIR

# Function to package and deploy a Lambda function
deploy_lambda() {
    local function_name="Gogen$1"
    local handler_file="${1,,}.py"  # Convert to lowercase
    local handler_name="${1,,}.lambda_handler"
    local zip_file="$BUILD_DIR/${function_name}.zip"
    
    echo "Packaging $function_name..."
    
    # Create a temporary directory for packaging
    local temp_dir=$(mktemp -d)
    
    # Copy the handler file and dependencies
    cp "$LAMBDA_DIR/$handler_file" "$temp_dir/"
    cp "$LAMBDA_DIR/db_utils.py" "$temp_dir/"
    cp "$LAMBDA_DIR/logger.py" "$temp_dir/"
    
    # Copy s3_utils.py if needed by this function
    if [[ "$1" == "Get" || "$1" == "Upsert" ]]; then
        cp "$LAMBDA_DIR/s3_utils.py" "$temp_dir/"
    fi
    
    # Install dependencies into the package
    if [ -f "$SCRIPT_DIR/requirements.txt" ]; then
        echo "Installing dependencies from requirements.txt..."
        pip install -r "$SCRIPT_DIR/requirements.txt" -t "$temp_dir/" --no-cache-dir
    else
        echo "requirements.txt not found, installing boto3 and botocore..."
        pip install boto3 botocore -t "$temp_dir/" --no-cache-dir
    fi
    
    # Create zip file
    echo "Creating zip file: $zip_file"
    (cd "$temp_dir" && zip -r "$zip_file" .)
    
    # Check if Lambda function exists
    echo "Checking if Lambda function $function_name exists..."
    if aws lambda get-function --function-name "$function_name" --region "$REGION" 2>&1 | grep -q "Function not found"; then
        # Create new Lambda function
        echo "Creating new Lambda function: $function_name"
        aws lambda create-function \
            --function-name "$function_name" \
            --runtime "$RUNTIME" \
            --role "$ROLE_ARN" \
            --handler "$handler_name" \
            --zip-file "fileb://$zip_file" \
            --region "$REGION"
    else
        # Update existing Lambda function
        echo "Updating existing Lambda function: $function_name"
        aws lambda update-function-code \
            --function-name "$function_name" \
            --zip-file "fileb://$zip_file" \
            --region "$REGION"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    echo "$function_name deployment complete!"
}

# Validate AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo "AWS CLI is not installed. Installing..."
    pip install awscli
fi

# Validate AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "AWS credentials are not configured. Please run 'aws configure' first."
    exit 1
fi

# Check if ROLE_ARN is set
if [ -z "$ROLE_ARN" ]; then
    echo "Please set the LAMBDA_ROLE_ARN environment variable or the ROLE_ARN variable in this script."
    exit 1
fi

# Deploy each Lambda function
deploy_lambda "Get"
deploy_lambda "List"
deploy_lambda "Search"
deploy_lambda "Upsert"

echo "All Lambda functions deployed successfully!" 