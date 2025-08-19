#!/bin/bash

# Script for pushing Docker images with version tags
# Usage: ./docker-push.sh [version]
# If no version is provided, only pushes 'latest' tag

VERSION=$1

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

if [ -n "$VERSION" ]; then
    echo "Pushing Docker images with version $VERSION..."
    
    # Tag and push main gogen image
    docker tag clintsharp/gogen:latest clintsharp/gogen:$VERSION
    docker push clintsharp/gogen:$VERSION
    docker push clintsharp/gogen:latest
    
    # Tag and push gogen-api image if it exists
    if docker images | grep -q "clintsharp/gogen-api"; then
        docker tag clintsharp/gogen-api:latest clintsharp/gogen-api:$VERSION
        docker push clintsharp/gogen-api:$VERSION
        docker push clintsharp/gogen-api:latest
    fi
else
    echo "Pushing Docker images with 'latest' tag only..."
    docker push clintsharp/gogen:latest
    
    # Push gogen-api if it exists
    if docker images | grep -q "clintsharp/gogen-api"; then
        docker push clintsharp/gogen-api:latest
    fi
fi

echo "Docker push completed successfully!"