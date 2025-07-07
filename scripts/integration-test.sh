#!/bin/bash
set -e

echo "Starting Sendria with Docker Compose..."
docker-compose up -d

echo "Waiting for Sendria to be ready..."
for i in {1..30}; do
    if curl -f http://localhost:1080/api/messages/ > /dev/null 2>&1; then
        echo "Sendria is ready!"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 2
done

echo "Running integration tests..."
export SENDRIA_URL=http://localhost:1080
export SENDRIA_SMTP_HOST=localhost:1025

go test -tags=integration -v -timeout 5m ./...

echo "Stopping Sendria..."
docker-compose down

echo "Integration tests completed!"