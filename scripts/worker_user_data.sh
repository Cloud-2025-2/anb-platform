#!/bin/bash
# Update package lists and install dependencies
apt-get update -y
apt-get install -y git docker.io

# Start and enable the Docker service
systemctl start docker
systemctl enable docker

# Add the 'ubuntu' user to the 'docker' group
usermod -aG docker ubuntu

# Clone the application repository into the ubuntu user's home directory
cd /home/ubuntu
git clone https://github.com/Cloud-2025-2/anb-platform

# Navigate to the backend directory
cd anb-platform/backend

# Build the Docker image for the worker
docker build -t anb-worker:latest -f Dockerfile.worker .

# Run the worker container
docker run -d --restart always --name worker anb-worker:latest