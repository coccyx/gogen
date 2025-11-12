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

# Parse command line arguments
ENVIRONMENT="prod"  # Default to production
while getopts "e:" opt; do
  case $opt in
    e) ENVIRONMENT="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

# Validate environment
if [[ "$ENVIRONMENT" != "prod" && "$ENVIRONMENT" != "staging" ]]; then
    echo "Invalid environment: $ENVIRONMENT. Must be 'prod' or 'staging'"
    exit 1
fi

# Set the S3 bucket name based on environment
if [ "$ENVIRONMENT" = "prod" ]; then
    S3_BUCKET="gogen-artifacts-prod"
else
    S3_BUCKET="gogen-artifacts-staging"
fi

# Ensure the S3 bucket exists
ensure_s3_bucket() {
    local bucket=$1
    if ! aws s3api head-bucket --bucket "$bucket" 2>/dev/null; then
        echo "S3 bucket $bucket does not exist or is not accessible"
        exit 1
    fi
    echo "Using S3 bucket: $bucket for deployment artifacts"
}

# Check S3 bucket
ensure_s3_bucket "$S3_BUCKET"

# Expect ROLE_ARN to be set as an environment variable
if [ -z "$ROLE_ARN" ]; then
    echo "Error: ROLE_ARN environment variable is not set." >&2
    exit 1
fi
echo "Using role ARN from environment: $ROLE_ARN"

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
    cp "$LAMBDA_DIR/cors_utils.py" "$temp_dir/"
    
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

# Validate AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "AWS credentials are not configured. Please run 'aws configure' first."
    exit 1
fi

# Get or create ACM certificate ARN
get_certificate_arn() {
    local domain="*.gogen.io"
    
    # Get certificate from us-east-1 (required for API Gateway)
    local cert_arn
    cert_arn=$(aws acm list-certificates --region us-east-1 --query "CertificateSummaryList[?DomainName=='$domain'].CertificateArn" --output text)
    
    if [ -z "$cert_arn" ]; then
        echo "Error: No certificate found for $domain in us-east-1"
        exit 1
    fi
    
    echo "$cert_arn"
}

# Get certificate ARN
echo "Checking for SSL certificate..."
CERT_ARN=$(get_certificate_arn)
if [ -z "$CERT_ARN" ]; then
    echo "Failed to get certificate ARN"
    exit 1
fi

echo "Using certificate ARN: $CERT_ARN"

# Make sure we are in the right directory for SAM commands
cd "$SCRIPT_DIR"

# Build the SAM application
echo "Building SAM application..."
echo "Ensuring clean build directory..."
rm -rfv .aws-sam/build # Ensure clean build directory (Verbose)
echo "Build directory cleanup attempted."
sam build --use-container

# Debug: Print the built template contents
echo "--- Contents of built template ---"
cat .aws-sam/build/template.yaml || echo "Built template not found!"
echo "--- End of built template ---"

# Deploy the SAM application
echo "Deploying SAM application for $ENVIRONMENT environment..."
echo "Using parameters:"
echo "  Environment: ${ENVIRONMENT}"
echo "  LambdaRoleArn: ${ROLE_ARN}"
echo "  CertificateArn: ${CERT_ARN}"
echo "  ProdTableName=gogen"
echo "  StagingTableName=gogen-staging"

# Print the exact parameters being used
echo "Parameter overrides:"
echo "  Environment=${ENVIRONMENT}"
echo "  LambdaRoleArn=${ROLE_ARN}"
echo "  CertificateArn=${CERT_ARN}"
echo "  ProdTableName=gogen"
echo "  StagingTableName=gogen-staging"

sam deploy \
    --stack-name "gogen-api-${ENVIRONMENT}" \
    --s3-bucket "$S3_BUCKET" \
    --parameter-overrides \
        ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
        ParameterKey=LambdaRoleArn,ParameterValue=${ROLE_ARN} \
        ParameterKey=CertificateArn,ParameterValue=${CERT_ARN} \
        ParameterKey=ProdTableName,ParameterValue=gogen \
        ParameterKey=StagingTableName,ParameterValue=gogen-staging \
    --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
    --no-confirm-changeset \
    --no-fail-on-empty-changeset

echo "Deployment completed successfully for $ENVIRONMENT environment!" 
