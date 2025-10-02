#!/bin/bash
# ANB Platform - Workers User Data Script
# Paste this in EC2 User Data when launching Workers instance
# This script runs automatically on first boot

set -e
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting workers setup at $(date)"

# === CONFIGURATION SECTION - EDIT THESE VALUES ===
NFS_SERVER_IP="10.0.2.100"  # REQUIRED: Replace with your NFS server private IP
WEBSERVER_PRIVATE_IP="10.0.1.100"  # REQUIRED: Replace with webserver private IP
POSTGRES_PASSWORD="ChangeThisSecurePassword123!"  # REQUIRED: Must match webserver password
WORKER_REPLICAS="3"  # Optional: Number of worker containers (default: 3)
WORKER_CONCURRENCY="2"  # Optional: Concurrent tasks per worker (default: 2)
# === END CONFIGURATION SECTION ===

# Update system
yum update -y

# Install Docker
yum install -y docker git nfs-utils

# Start Docker
systemctl start docker
systemctl enable docker

# Add ec2-user to docker group
usermod -aG docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# Verify installations
docker --version
docker-compose --version

# Create NFS mount point
mkdir -p /mnt/nfs/anb-storage

# Mount NFS
echo "Mounting NFS from ${NFS_SERVER_IP}..."
mount -t nfs ${NFS_SERVER_IP}:/exports/anb-storage /mnt/nfs/anb-storage

# Add to /etc/fstab for persistence
echo "${NFS_SERVER_IP}:/exports/anb-storage /mnt/nfs/anb-storage nfs defaults,_netdev 0 0" >> /etc/fstab

# Verify NFS mount
df -h | grep nfs

# Test connectivity to webserver
echo "Testing connectivity to webserver..."
timeout 10 bash -c "cat < /dev/null > /dev/tcp/${WEBSERVER_PRIVATE_IP}/5432" && echo "PostgreSQL port accessible" || echo "WARNING: Cannot reach PostgreSQL"
timeout 10 bash -c "cat < /dev/null > /dev/tcp/${WEBSERVER_PRIVATE_IP}/9092" && echo "Kafka port accessible" || echo "WARNING: Cannot reach Kafka"

# Clone repository
cd /home/ec2-user
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform

# Set ownership
chown -R ec2-user:ec2-user /home/ec2-user/anb-platform

# Create environment file
cat > .env.workers <<EOF
WEBSERVER_PRIVATE_IP=${WEBSERVER_PRIVATE_IP}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
WORKER_REPLICAS=${WORKER_REPLICAS}
WORKER_CONCURRENCY=${WORKER_CONCURRENCY}
EOF

# Load environment variables and start workers
cd /home/ec2-user/anb-platform
export $(cat .env.workers | xargs)

# Build and start workers
docker-compose -f docker-compose.workers.yml up --build -d --scale video-processor=${WORKER_REPLICAS}

# Wait for workers to start
echo "Waiting for workers to start..."
sleep 30

# Get instance info
PRIVATE_IP=$(hostname -I | awk '{print $1}')
WORKER_COUNT=$(docker ps --filter "name=video-processor" --format "{{.Names}}" | wc -l)

# Create info file
cat > /home/ec2-user/workers-info.txt <<EOF
Workers Setup Complete!
========================
Private IP: ${PRIVATE_IP}
Workers Running: ${WORKER_COUNT}

Configuration:
- Webserver IP: ${WEBSERVER_PRIVATE_IP}
- NFS Server IP: ${NFS_SERVER_IP}
- Worker Replicas: ${WORKER_REPLICAS}
- Concurrency: ${WORKER_CONCURRENCY}
- Consumer Group: video-processors

View logs:
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.workers.yml logs -f

Check workers:
docker ps | grep video-processor

Scale workers:
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5

Monitor Kafka (from webserver):
docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group video-processors --describe

Setup completed at: $(date)
EOF

chown ec2-user:ec2-user /home/ec2-user/workers-info.txt

echo "========================================"
echo "Workers setup completed at $(date)"
echo "Workers running: ${WORKER_COUNT}"
echo "Check /home/ec2-user/workers-info.txt for details"
echo "========================================"
