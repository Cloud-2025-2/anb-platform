#!/bin/bash
# Update package lists and install dependencies
apt-get update -y
apt-get install -y git docker.io

# Start and enable the Docker service
systemctl start docker
systemctl enable docker

# Add the 'ubuntu' user to the 'docker' group
usermod -aG docker ubuntu

# Run subsequent commands as the 'ubuntu' user
runuser -l ubuntu -c '
  git clone https://github.com/Cloud-2025-2/anb-platform && \
  cd anb-platform/backend && \
  docker build -t anb-worker:latest -f Dockerfile.worker . && \
  docker run -d --restart always \
    -e POSTGRES_HOST="anb-platform-db.cibipwvslz8o.us-east-1.rds.amazonaws.com" \
    -e POSTGRES_USER="postgres" \
    -e 'POSTGRES_PASSWORD=X|qI3iSdd$(G(B2?6T7jmLQfD?(6' \
    -e POSTGRES_DB="anb_platform" \
    --name worker anb-worker:latest
'