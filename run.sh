#!/bin/bash

# Build the Docker images for the services defined in the docker-compose.yml
echo "Building Docker images..."
docker compose build

# Start up the services in the background
echo "Starting up services..."
docker compose up -d

# Check if the services are running
echo "Checking if services are up..."
docker compose ps

# Tail the logs of the services to see the output
echo "Tailing logs..."
docker compose logs -f