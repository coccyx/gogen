#!/bin/bash

# Get the script's directory and the root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
VENV_DIR="$ROOT_DIR/.pyvenv"

# Remove existing virtual environment if it exists
rm -rf "$VENV_DIR"

# Create new virtual environment
python3 -m venv "$VENV_DIR"

# Activate virtual environment
source "$VENV_DIR/bin/activate"

# Upgrade pip
pip install --upgrade pip

# Install requirements from llmcontext directory
pip install -r "$SCRIPT_DIR/requirements.txt"

echo "Virtual environment setup complete in $VENV_DIR" 