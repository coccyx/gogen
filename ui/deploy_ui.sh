#!/bin/bash
set -e

# Determine script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

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

# Ensure we're in the UI directory
cd "$SCRIPT_DIR"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "npm is not installed. Please install npm first."
    exit 1
fi

# Validate AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "AWS credentials are not configured. Please run 'aws configure' first."
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
npm install

# Build the application
echo "Building UI for $ENVIRONMENT environment..."
if [ "$ENVIRONMENT" = "prod" ]; then
    npm run build
    BUCKET="gogen.io"
else
    npm run build:staging
    BUCKET="staging.gogen.io"
fi

# Deploy to S3
echo "Deploying to s3://$BUCKET/"
aws s3 sync dist/ "s3://$BUCKET/" --delete

echo "Deployment completed successfully for $ENVIRONMENT environment!" 