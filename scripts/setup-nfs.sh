#!/bin/bash
# ANB Platform - NFS Server Setup Script
# Run this script on EC2 NFS instance

set -e

echo "========================================="
echo "ANB Platform - NFS Server Setup"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Get VPC CIDR from user
echo "Enter your VPC CIDR (e.g., 10.0.0.0/16):"
read VPC_CIDR

if [ -z "$VPC_CIDR" ]; then
    echo -e "${RED}Error: VPC CIDR is required${NC}"
    exit 1
fi

echo "Using VPC CIDR: $VPC_CIDR"

# Detect OS and install NFS
echo "Installing NFS server..."
if [ -f /etc/os-release ]; then
    . /etc/os-release
    if [[ "$ID" == "amzn" ]] || [[ "$ID_LIKE" == *"rhel"* ]]; then
        # Amazon Linux or RHEL-based
        yum update -y
        yum install -y nfs-utils
        systemctl enable nfs-server
        systemctl start nfs-server
    elif [[ "$ID" == "ubuntu" ]] || [[ "$ID_LIKE" == *"debian"* ]]; then
        # Ubuntu or Debian-based
        apt-get update
        apt-get install -y nfs-kernel-server
        systemctl enable nfs-kernel-server
        systemctl start nfs-kernel-server
    else
        echo -e "${RED}Unsupported OS${NC}"
        exit 1
    fi
else
    echo -e "${RED}Cannot detect OS${NC}"
    exit 1
fi

echo -e "${GREEN}✓ NFS server installed${NC}"

# Create export directory
echo "Creating export directory..."
mkdir -p /exports/anb-storage
mkdir -p /exports/anb-storage/videos
mkdir -p /exports/anb-storage/thumbnails
mkdir -p /exports/anb-storage/processed
mkdir -p /exports/anb-storage/original

# Set permissions
chown -R ec2-user:ec2-user /exports/anb-storage 2>/dev/null || chown -R ubuntu:ubuntu /exports/anb-storage
chmod -R 755 /exports/anb-storage

echo -e "${GREEN}✓ Export directory created${NC}"

# Configure exports
echo "Configuring NFS exports..."
EXPORTS_ENTRY="/exports/anb-storage ${VPC_CIDR}(rw,sync,no_subtree_check,no_root_squash)"

# Backup existing exports
if [ -f /etc/exports ]; then
    cp /etc/exports /etc/exports.backup
fi

# Add or update export
if grep -q "/exports/anb-storage" /etc/exports; then
    sed -i "\|/exports/anb-storage|c\\${EXPORTS_ENTRY}" /etc/exports
else
    echo "$EXPORTS_ENTRY" >> /etc/exports
fi

echo -e "${GREEN}✓ NFS exports configured${NC}"

# Apply exports
echo "Applying NFS exports..."
exportfs -ra

# Verify exports
echo ""
echo "Current NFS exports:"
exportfs -v

# Show mount command for clients
echo ""
echo "========================================="
echo -e "${GREEN}NFS Server setup completed!${NC}"
echo "========================================="
echo ""
echo "NFS Server IP: $(hostname -I | awk '{print $1}')"
echo "Export Path: /exports/anb-storage"
echo "VPC CIDR: $VPC_CIDR"
echo ""
echo "To mount on client machines, run:"
echo "  sudo mkdir -p /mnt/nfs/anb-storage"
echo "  sudo mount -t nfs $(hostname -I | awk '{print $1}'):/exports/anb-storage /mnt/nfs/anb-storage"
echo ""
echo "To make mount persistent, add to /etc/fstab:"
echo "  $(hostname -I | awk '{print $1}'):/exports/anb-storage /mnt/nfs/anb-storage nfs defaults,_netdev 0 0"
echo ""
echo "Directory structure:"
echo "  /exports/anb-storage/videos      - User uploaded videos"
echo "  /exports/anb-storage/thumbnails  - Video thumbnails"
echo "  /exports/anb-storage/processed   - Processed videos"
echo "  /exports/anb-storage/original    - Original videos"
echo ""
