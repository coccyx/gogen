#!/usr/bin/env python3

import os
import yaml
import json
import markdown
import datetime
from pathlib import Path
from typing import Dict, List, Any
import re

class DateTimeEncoder(json.JSONEncoder):
    """Custom JSON encoder for handling datetime objects."""
    def default(self, obj):
        if isinstance(obj, datetime.datetime):
            return obj.isoformat()
        return super().default(obj)

def clean_markdown(md_content: str) -> str:
    """Convert markdown to plain text while preserving structure."""
    # Remove code blocks but keep their content
    md_content = re.sub(r'```[^\n]*\n', '', md_content)
    md_content = re.sub(r'```', '', md_content)
    
    # Convert headers to plain text with newlines
    md_content = re.sub(r'^#{1,6}\s+(.+)$', r'\1\n', md_content, flags=re.MULTILINE)
    
    return md_content.strip()

def extract_yaml_comments(yaml_content: str) -> Dict[str, str]:
    """Extract inline comments from YAML content."""
    comments = {}
    for line_num, line in enumerate(yaml_content.split('\n'), 1):
        if '#' in line:
            code, comment = line.split('#', 1)
            if code.strip():  # Only store comments for lines with actual code
                comments[line_num] = comment.strip()
    return comments

def read_yaml_file(file_path: str) -> Dict[str, Any]:
    """Read and parse a YAML file with error handling."""
    try:
        with open(file_path, 'r') as f:
            content = f.read()
            # Replace tabs with spaces to avoid YAML parsing errors
            content = content.replace('\t', '  ')
            yaml_content = yaml.safe_load(content)
            comments = extract_yaml_comments(content)
            return {
                'content': yaml_content,
                'raw_content': content,
                'comments': comments
            }
    except Exception as e:
        print(f"Warning: Error reading YAML file {file_path}: {str(e)}")
        # Return partial data even if YAML parsing fails
        return {
            'content': None,
            'raw_content': content if 'content' in locals() else None,
            'comments': extract_yaml_comments(content) if 'content' in locals() else {},
            'error': str(e)
        }

def read_markdown_file(file_path: str) -> str:
    """Read and process a markdown file."""
    try:
        with open(file_path, 'r') as f:
            content = f.read()
            return clean_markdown(content)
    except Exception as e:
        print(f"Error reading markdown file {file_path}: {str(e)}")
        return ""

def get_category_from_path(file_path: str) -> str:
    """Extract category from file path."""
    parts = Path(file_path).parts
    if 'examples' in parts:
        idx = parts.index('examples')
        if len(parts) > idx + 1:
            return parts[idx + 1]
    return "uncategorized"

def process_examples(base_path: str) -> List[Dict[str, Any]]:
    """Process all YAML examples in the codebase."""
    examples = []
    example_dirs = [
        'examples/tutorial',
        'examples/weblog',
        'examples/csv',
        'examples/nixOS',
    ]
    
    for dir_path in example_dirs:
        full_path = os.path.join(base_path, dir_path)
        if not os.path.exists(full_path):
            continue
            
        for root, _, files in os.walk(full_path):
            for file in files:
                if file.endswith(('.yml', '.yaml')):
                    file_path = os.path.join(root, file)
                    yaml_data = read_yaml_file(file_path)
                    
                    if yaml_data['content'] is not None:
                        example = {
                            'name': file,
                            'category': get_category_from_path(file_path),
                            'yaml_content': yaml_data['raw_content'],
                            'parsed_content': yaml_data['content'],
                            'comments': yaml_data['comments'],
                            'file_path': os.path.relpath(file_path, base_path)
                        }
                        examples.append(example)
    
    return examples

def process_documentation(base_path: str) -> Dict[str, str]:
    """Process all relevant documentation files."""
    docs = {}
    doc_files = {
        'reference': 'README/Reference.md',
        'tutorial': 'README/Tutorial.md',
        'examples': 'README/Examples.md'
    }
    
    for doc_type, file_path in doc_files.items():
        full_path = os.path.join(base_path, file_path)
        if os.path.exists(full_path):
            docs[doc_type] = read_markdown_file(full_path)
    
    return docs

def main():
    # Use current directory as base path
    base_path = os.getcwd()
    
    # Process examples and documentation
    examples = process_examples(base_path)
    documentation = process_documentation(base_path)
    
    # Create final structure
    bundle = {
        'examples': examples,
        'documentation': documentation,
        'metadata': {
            'total_examples': len(examples),
            'documentation_sections': list(documentation.keys()),
            'generated_at': datetime.datetime.now()  # No need to call isoformat() here
        }
    }
    
    # Write to file
    output_file = 'gogen_examples_bundle.json'
    with open(output_file, 'w') as f:
        json.dump(bundle, f, indent=2, cls=DateTimeEncoder)
    
    print(f"Successfully bundled {len(examples)} examples and {len(documentation)} documentation sections to {output_file}")
    
    # Print any examples that had parsing errors
    errors = [ex for ex in examples if ex.get('parsed_content') is None]
    if errors:
        print("\nWarning: The following files had YAML parsing errors but were included with raw content:")
        for ex in errors:
            print(f"- {ex['file_path']}")

if __name__ == '__main__':
    main() 