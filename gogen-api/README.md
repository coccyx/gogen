# Gogen API Local Development Environment

This document describes how to set up and use the local development environment for the Gogen API.

## Overview

The local development environment includes:

- **DynamoDB Local**: A local version of AWS DynamoDB for data storage
- **MinIO**: A local S3-compatible object storage server for the `gogen-configs` bucket
- **SAM Local**: AWS Serverless Application Model for local Lambda function development

This setup allows you to develop and test the entire Gogen API stack without needing access to AWS services.

## Prerequisites

- Docker and Docker Compose
- AWS SAM CLI
- Python 3.13 or compatible version
- AWS CLI (optional, for advanced testing)

## Python Virtual Environment Setup

The project uses a Python virtual environment located at `.pyvenv` in the root directory. This keeps dependencies isolated and ensures consistent development across different machines.

### Creating the Virtual Environment

If the `.pyvenv` directory doesn't exist, create it with the following commands:

```bash
# Navigate to the project root
cd /home/clint/local/src/gogen

# Create the virtual environment
python3 -m venv .pyvenv

# Activate the virtual environment
source .pyvenv/bin/activate

# Install required packages
cd gogen-api
pip install -r requirements.txt
pip install -r requirements-dev.txt  # If it exists
```

### Using the Virtual Environment with SAM

To use the virtual environment with SAM:

1. Activate the virtual environment:
   ```bash
   source /home/clint/local/src/gogen/.pyvenv/bin/activate
   ```

2. Configure SAM to use the virtual environment by adding this to your `samconfig.toml` or when running SAM commands:
   ```bash
   sam build --use-container
   # or
   sam local invoke --env-vars env.json
   ```

3. For development, you can add this to your `.bashrc` or `.zshrc` to automatically activate the environment when entering the project directory:
   ```bash
   function cd() {
     builtin cd "$@"
     if [[ -d .pyvenv ]] && [[ -f .pyvenv/bin/activate ]]; then
       source .pyvenv/bin/activate
     fi
   }
   ```

### Updating Dependencies

When dependencies change, update your virtual environment:

```bash
source /home/clint/local/src/gogen/.pyvenv/bin/activate
pip install -r gogen-api/requirements.txt
```

## Starting the Development Environment

The easiest way to start the entire development environment is to use the provided script:

```bash
cd gogen-api
./start_dev.sh
```

This script will:
1. Start the Docker containers (DynamoDB and MinIO) using Docker Compose
2. Set up the local DynamoDB with the correct schema
3. Build the SAM application
4. Start the SAM local API on port 4000
5. Clean up resources when you exit (Ctrl+C)

## Manual Setup (if needed)

If you prefer to start services individually:

1. Start the Docker containers:
   ```bash
   docker-compose up -d
   ```

2. Set up the local DynamoDB:
   ```bash
   ./setup_local_db.sh
   ```

3. Build and start the SAM application:
   ```bash
   # Make sure your virtual environment is activated
   source /home/clint/local/src/gogen/.pyvenv/bin/activate
   
   # Build and start SAM
   sam build
   sam local start-api --host 0.0.0.0 --port 4000 --warm-containers EAGER --docker-network lambda-local
   ```

## Testing the Environment

### Testing DynamoDB

A test script is provided to verify the DynamoDB connection:

```bash
# Activate virtual environment if not already active
source /home/clint/local/src/gogen/.pyvenv/bin/activate

python test_dynamodb.py
```

### Testing S3/MinIO

A test script is provided to verify the S3 connection:

```bash
# Activate virtual environment if not already active
source /home/clint/local/src/gogen/.pyvenv/bin/activate

python test_s3.py
```

This script will:
1. Connect to the local MinIO server
2. List available buckets
3. Upload a test file to the `gogen-configs` bucket
4. List objects in the bucket
5. Download and verify the test file

## Accessing Services

### API Endpoints

The SAM local API is available at: http://localhost:4000

Available endpoints:
- GET /v1/get/{gogen} - Get a specific Gogen configuration
- POST /v1/upsert - Create or update a Gogen configuration
- GET /v1/list - List all available Gogen configurations

### DynamoDB Local

DynamoDB Local is available at: http://localhost:8000

### MinIO (S3)

#### Web Console

You can access the MinIO web console at: http://localhost:9001

Login with:
- Username: `minioadmin`
- Password: `minioadmin`

#### API

The MinIO S3 API is available at: http://localhost:9000

## Using Services in Your Code

### DynamoDB

Use the provided utility module for DynamoDB operations:

```python
from api.db_utils import get_dynamodb_client

# Get DynamoDB client
dynamodb = get_dynamodb_client()
table = dynamodb.Table('gogen')

# Perform operations
response = table.get_item(Key={'gogen': 'my-gogen'})
```

### S3/MinIO

Use the provided utility module for S3 operations:

```python
from api.s3_utils import upload_config, download_config, list_configs, delete_config

# Upload a config
upload_config('my-config.json', '{"key": "value"}')

# Download a config
content = download_config('my-config.json')

# List all configs
configs = list_configs()

# Delete a config
delete_config('my-config.json')
```

## AWS CLI with Local Services

### DynamoDB

```bash
aws dynamodb list-tables --endpoint-url http://localhost:8000
aws dynamodb scan --table-name gogen --endpoint-url http://localhost:8000
```

### S3/MinIO

You can use the AWS CLI with MinIO by creating a profile:

```bash
aws configure --profile minio
```

Enter:
- AWS Access Key ID: `minioadmin`
- AWS Secret Access Key: `minioadmin`
- Default region name: `us-east-1`
- Default output format: `json`

Then use the profile with the endpoint URL:

```bash
aws --endpoint-url http://localhost:9000 --profile minio s3 ls
aws --endpoint-url http://localhost:9000 --profile minio s3 ls s3://gogen-configs
```

## Connecting from Docker Containers

When connecting to services from other Docker containers in the same network:
- Use `dynamodb-local:8000` instead of `localhost:8000` for DynamoDB
- Use `minio:9000` instead of `localhost:9000` for S3/MinIO

## Data Persistence

- DynamoDB data is persisted in the container
- MinIO data is persisted in a Docker volume named `minio-data`

This ensures your data is preserved between container restarts.

## Troubleshooting

### General Issues

If you encounter issues with the development environment:

1. Check if all containers are running:
   ```bash
   docker-compose ps
   ```

2. Check the logs:
   ```bash
   docker-compose logs
   ```

3. Restart the services:
   ```bash
   docker-compose restart
   ```

4. If all else fails, recreate the services:
   ```bash
   docker-compose down
   docker-compose up -d
   ./setup_local_db.sh
   ```

### DynamoDB Issues

If DynamoDB is not working correctly:
```bash
docker-compose logs dynamodb-local
```

### S3/MinIO Issues

If MinIO is not working correctly:
```bash
docker-compose logs minio
docker-compose logs createbuckets
```

## Development Guidelines

- Keep components modular and focused on a single responsibility
- Document code with clear comments
- Update SUMMARY.md after completing significant features
- Each AWS Lambda function should be a separate .py file in the `api` directory
- Remember that the codebase is being updated from Python 2.7 to Python 3.13 