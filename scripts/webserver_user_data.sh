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

# Build the Docker image for the API
docker build -t anb-api:latest -f Dockerfile .

# Run the API container, mapping port 80 on the host to 8000 in the container
docker run -d --restart always -p 80:8000 --name api-server anb-api:latest