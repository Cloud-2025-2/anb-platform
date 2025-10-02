#!/bin/bash
# ANB Platform - Webserver User Data Script
# Paste this in EC2 User Data when launching Webserver instance
# This script runs automatically on first boot

set -e
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting webserver setup at $(date)"

# === CONFIGURATION SECTION - EDIT THESE VALUES ===
NFS_SERVER_IP="10.0.2.100"  # REQUIRED: Replace with your NFS server private IP
POSTGRES_PASSWORD="ChangeThisSecurePassword123!"  # REQUIRED: Set a strong password
JWT_SECRET="your-super-secure-jwt-secret-minimum-32-characters-here"  # REQUIRED: Set a secure JWT secret
WEBSERVER_PUBLIC_IP="ec2-xx-xxx-xxx-xxx.compute-1.amazonaws.com"  # REQUIRED: Your EC2 public IP or domain
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

# Clone repository
cd /home/ec2-user
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform

# Set ownership
chown -R ec2-user:ec2-user /home/ec2-user/anb-platform

# Copy assets to NFS storage
if [ -d "backend/assets" ]; then
    cp -r backend/assets/* /mnt/nfs/anb-storage/
    echo "Assets copied to NFS storage"
fi

# Create environment file
cat > .env.webserver <<EOF
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
JWT_SECRET=${JWT_SECRET}
WEBSERVER_PUBLIC_IP=${WEBSERVER_PUBLIC_IP}
EOF

# Update Kafka advertised listeners in docker-compose
sed -i "s/<WEBSERVER_PUBLIC_IP>/${WEBSERVER_PUBLIC_IP}/g" docker-compose.webserver.yml

# Update NFS mount paths in docker-compose (ensure consistency)
sed -i "s|/mnt/nfs/anb-storage|/mnt/nfs/anb-storage|g" docker-compose.webserver.yml

# Load environment variables and start services
cd /home/ec2-user/anb-platform
export $(cat .env.webserver | xargs)

# Build and start services
docker-compose -f docker-compose.webserver.yml up --build -d

# Wait for services to start
echo "Waiting for services to start..."
sleep 45

# Create Kafka topics
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing --partitions 3 --replication-factor 1 --if-not-exists || true
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-retry --partitions 3 --replication-factor 1 --if-not-exists || true
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-dlq --partitions 1 --replication-factor 1 --if-not-exists || true

# Get public IP
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
PRIVATE_IP=$(hostname -I | awk '{print $1}')

# Create info file
cat > /home/ec2-user/webserver-info.txt <<EOF
Webserver Setup Complete!
========================
Public IP: ${PUBLIC_IP}
Private IP: ${PRIVATE_IP}

Access Points:
- Frontend:  http://${PUBLIC_IP}
- API:       http://${PUBLIC_IP}/api/health
- Swagger:   http://${PUBLIC_IP}/swagger/index.html

Database:
- PostgreSQL Port: 5432 (accessible from workers)
- Redis Port: 6379 (accessible from workers)
- Kafka Port: 9092 (accessible from workers)

View logs:
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.webserver.yml logs -f

Check services:
docker ps

Setup completed at: $(date)
EOF

chown ec2-user:ec2-user /home/ec2-user/webserver-info.txt

echo "========================================"
echo "Webserver setup completed at $(date)"
echo "Public IP: ${PUBLIC_IP}"
echo "Frontend: http://${PUBLIC_IP}"
echo "Check /home/ec2-user/webserver-info.txt for details"
echo "========================================"
