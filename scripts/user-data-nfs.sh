#!/bin/bash
# ANB Platform - NFS Server User Data Script
# Paste this in EC2 User Data when launching NFS instance
# This script runs automatically on first boot

set -e
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting NFS server setup at $(date)"

# Update system
sudo yum update -y

# Install NFS server
sudo yum install -y nfs-utils

# Start and enable NFS
sudo systemctl enable nfs-server
sudo systemctl start nfs-server

# Create export directory structure
sudo mkdir -p /exports/anb-storage
sudo mkdir -p /exports/anb-storage/videos
sudo mkdir -p /exports/anb-storage/thumbnails
sudo mkdir -p /exports/anb-storage/processed
sudo mkdir -p /exports/anb-storage/original

# Set permissions (ec2-user for Amazon Linux)
sudo chown -R ec2-user:ec2-user /exports/anb-storage
sudo chmod -R 777 /exports/anb-storage

# Configure NFS exports for entire VPC
# IMPORTANT: Update 10.0.0.0/16 with your VPC CIDR
VPC_CIDR="10.0.0.0/16"
echo "/exports/anb-storage ${VPC_CIDR}(rw,sync,no_subtree_check,no_root_squash)" | sudo tee /etc/exports

# Apply exports
sudo exportfs -ra

# Verify
sudo exportfs -v

# Get private IP
PRIVATE_IP=$(hostname -I | awk '{print $1}')

# Create info file
cat > /home/ec2-user/nfs-info.txt <<EOF
NFS Server Setup Complete!
========================
NFS Server IP: ${PRIVATE_IP}
Export Path: /exports/anb-storage
VPC CIDR: ${VPC_CIDR}

Mount command for clients:
sudo mkdir -p /mnt/nfs/anb-storage
sudo mount -t nfs ${PRIVATE_IP}:/exports/anb-storage /mnt/nfs/anb-storage

Persistent mount (add to /etc/fstab):
${PRIVATE_IP}:/exports/anb-storage /mnt/nfs/anb-storage nfs defaults,_netdev 0 0

Setup completed at: $(date)
EOF

sudo chown ec2-user:ec2-user /home/ec2-user/nfs-info.txt

echo "NFS server setup completed successfully at $(date)"
echo "NFS Server IP: ${PRIVATE_IP}"
echo "Check /home/ec2-user/nfs-info.txt for connection details"
