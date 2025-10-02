#!/bin/bash
# ANB Platform - NFS Server User Data Script
# Paste this in EC2 User Data when launching NFS instance
# This script runs automatically on first boot

set -e
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting NFS server setup at $(date)"

# Update system
yum update -y

# Install NFS server
yum install -y nfs-utils

# Start and enable NFS
systemctl enable nfs-server
systemctl start nfs-server

# Create export directory structure
mkdir -p /exports/anb-storage
mkdir -p /exports/anb-storage/videos
mkdir -p /exports/anb-storage/thumbnails
mkdir -p /exports/anb-storage/processed
mkdir -p /exports/anb-storage/original

# Set permissions (ec2-user for Amazon Linux)
chown -R ec2-user:ec2-user /exports/anb-storage
chmod -R 755 /exports/anb-storage

# Configure NFS exports for entire VPC
# IMPORTANT: Update 10.0.0.0/16 with your VPC CIDR
VPC_CIDR="10.0.0.0/16"
echo "/exports/anb-storage ${VPC_CIDR}(rw,sync,no_subtree_check,no_root_squash)" > /etc/exports

# Apply exports
exportfs -ra

# Verify
exportfs -v

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

chown ec2-user:ec2-user /home/ec2-user/nfs-info.txt

echo "NFS server setup completed successfully at $(date)"
echo "NFS Server IP: ${PRIVATE_IP}"
echo "Check /home/ec2-user/nfs-info.txt for connection details"
