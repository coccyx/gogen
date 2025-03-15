#!/bin/bash

# Script to set up the Python virtual environment for the Gogen API project

# Define paths
VENV_PATH=".pyvenv"
PROJECT_ROOT="$(pwd)"
API_DIR="$PROJECT_ROOT/gogen-api"

# Check if virtual environment already exists
if [ -d "$VENV_PATH" ]; then
    echo "Virtual environment already exists at $VENV_PATH"
    echo "To recreate it, delete the directory first: rm -rf $VENV_PATH"
    echo "To activate it: source $VENV_PATH/bin/activate"
    exit 0
fi

# Check Python version
PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
echo "Using Python version: $PYTHON_VERSION"

# Create virtual environment
echo "Creating virtual environment at $VENV_PATH..."
python3 -m venv "$VENV_PATH"

if [ ! -d "$VENV_PATH" ]; then
    echo "Failed to create virtual environment. Please check your Python installation."
    exit 1
fi

# Activate virtual environment
echo "Activating virtual environment..."
source "$VENV_PATH/bin/activate"

# Upgrade pip
echo "Upgrading pip..."
pip install --upgrade pip

# Install requirements if they exist
if [ -f "$API_DIR/requirements.txt" ]; then
    echo "Installing requirements from $API_DIR/requirements.txt..."
    pip install -r "$API_DIR/requirements.txt"
else
    echo "Warning: requirements.txt not found at $API_DIR/requirements.txt"
fi

# Install development requirements if they exist
if [ -f "$API_DIR/requirements-dev.txt" ]; then
    echo "Installing development requirements from $API_DIR/requirements-dev.txt..."
    pip install -r "$API_DIR/requirements-dev.txt"
fi

# Install AWS SAM CLI if not already installed
if ! command -v sam &> /dev/null; then
    echo "AWS SAM CLI not found. Installing..."
    pip install aws-sam-cli
else
    echo "AWS SAM CLI already installed: $(sam --version)"
fi

echo ""
echo "Virtual environment setup complete!"
echo ""
echo "To activate the virtual environment, run:"
echo "  source $VENV_PATH/bin/activate"
echo ""
echo "To start the development environment, run:"
echo "  cd gogen-api"
echo "  ./start_dev.sh"
echo ""

# Add a note about automatically activating the environment
echo "Tip: Add this to your .bashrc or .zshrc to automatically activate"
echo "the environment when entering the project directory:"
echo ""
echo 'function cd() {'
echo '  builtin cd "$@"'
echo '  if [[ -d .pyvenv ]] && [[ -f .pyvenv/bin/activate ]]; then'
echo '    source .pyvenv/bin/activate'
echo '  fi'
echo '}' 