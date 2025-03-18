#!/bin/bash
set -e

# Determine script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

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
    S3_BUCKET="gogen-artifacts"
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

# Get the appropriate role ARN based on environment
get_role_arn() {
    local env=$1
    local role_name
    
    if [ "$env" = "prod" ]; then
        role_name="gogen_lambda"
    else
        role_name="gogen_lambda_staging"
    fi
    
    # Get the role ARN
    role_arn=$(aws iam get-role --role-name "$role_name" --query 'Role.Arn' --output text)
    if [ -z "$role_arn" ]; then
        echo "Failed to get ARN for role: $role_name" >&2
        exit 1
    fi
    echo "$role_arn"
}

# Get the role ARN
ROLE_ARN=$(get_role_arn "$ENVIRONMENT")
echo "Using role ARN: $ROLE_ARN"

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
        fi
        
        # Install AWS SAM CLI if needed
        if ! command -v sam &> /dev/null; then
            echo "Installing AWS SAM CLI..."
            pip install aws-sam-cli
        fi
    fi
fi

# Validate AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "AWS credentials are not configured. Please run 'aws configure' first."
    exit 1
fi

# Get or create ACM certificate ARN
get_certificate_arn() {
    # Check if certificate exists for *.gogen.io
    CERT_ARN=$(aws acm list-certificates --query "CertificateSummaryList[?DomainName=='*.gogen.io'].CertificateArn" --output text)
    
    if [ -z "$CERT_ARN" ]; then
        echo "No certificate found for *.gogen.io"
        exit 1
    fi
    
    echo $CERT_ARN
}

# Get certificate ARN
CERT_ARN=$(get_certificate_arn)
if [ -z "$CERT_ARN" ]; then
    echo "Failed to get certificate ARN"
    exit 1
fi

echo "Using certificate ARN: $CERT_ARN"

# Build the SAM application
echo "Building SAM application..."
sam build --use-container

# Deploy the SAM application
echo "Deploying SAM application for $ENVIRONMENT environment..."
sam deploy \
    --stack-name "gogen-api-${ENVIRONMENT}" \
    --s3-bucket "$S3_BUCKET" \
    --parameter-overrides \
        Environment=$ENVIRONMENT \
        LambdaRoleArn=$ROLE_ARN \
        CertificateArn=$CERT_ARN \
    --capabilities CAPABILITY_IAM \
    --no-confirm-changeset \
    --no-fail-on-empty-changeset

echo "Deployment completed successfully for $ENVIRONMENT environment!" 