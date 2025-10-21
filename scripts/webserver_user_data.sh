#!/bin/bash
# Update package lists and install dependencies
apt-get update -y
apt-get install -y git docker.io

# Start and enable the Docker service
systemctl start docker
systemctl enable docker

# Add the 'ubuntu' user to the 'docker' group
usermod -aG docker ubuntu

# Run user-specific commands as the 'ubuntu' user
su - ubuntu -c '
  cd /home/ubuntu
  git clone https://github.com/Cloud-2025-2/anb-platform
  cd anb-platform/backend
  docker build -t anb-api:latest -f Dockerfile .
  docker run -d --restart always -p 80:8000 \
    -e CORS_ALLOWED_ORIGINS="http://anb-platform-frontend.s3-website-us-east-1.amazonaws.com" \
    -e POSTGRES_HOST="anb-platform-db.cibipwvslz8o.us-east-1.rds.amazonaws.com" \
    -e POSTGRES_USER="postgres" \
    -e POSTGRES_PASSWORD="anb-platform123" \
    -e POSTGRES_DB="anb_platform" \
    --name api-server anb-api:latest
'