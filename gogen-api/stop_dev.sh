#!/bin/bash

cd "$(dirname "$0")"
docker compose down
docker ps | grep ecr | awk '{print $1}' | xargs docker kill
docker ps | grep dynamodb | awk '{print $1}' | xargs docker kill
docker ps | grep minio | awk '{print $1}' | xargs docker kill

