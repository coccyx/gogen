#!/bin/bash 
docker compose up -d
sleep 5
. ./setup_local_db.sh
sam local start-api --host 0.0.0.0 --port 4000 --warm-containers EAGER --docker-network lambda-local
