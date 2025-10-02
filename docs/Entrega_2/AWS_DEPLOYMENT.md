# AWS Deployment Guide - ANB Rising Stars Platform

## Architecture Overview

This deployment uses **3 EC2 instances** with a distributed architecture:

1. **EC2 Webserver** - Nginx, Backend API, Frontend, PostgreSQL, Redis, Kafka, Zookeeper
2. **EC2 Workers** - Video processing workers (scalable)
3. **EC2 NFS** - Shared storage for videos and assets

```
┌──────────────────────────────────────────────────────────────┐
│                      AWS VPC (Your Region)                    │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐│
│  │ EC2 #1 - Webserver (Public Subnet)                       ││
│  │ • Nginx (Port 80/443)                                     ││
│  │ • Backend API (Port 8000)                                 ││
│  │ • Frontend (React)                                        ││
│  │ • PostgreSQL (Port 5432)                                  ││
│  │ • Redis (Port 6379)                                       ││
│  │ • Kafka + Zookeeper (Port 9092)                           ││
│  └──────────────────────────────────────────────────────────┘│
│                            ↕                                   │
│  ┌──────────────────────────────────────────────────────────┐│
│  │ EC2 #2 - Workers (Private Subnet)                        ││
│  │ • Video Processor Workers (3-5 replicas)                  ││
│  │ • Connects to: Kafka, PostgreSQL, NFS                     ││
│  └──────────────────────────────────────────────────────────┘│
│                            ↕                                   │
│  ┌──────────────────────────────────────────────────────────┐│
│  │ EC2 #3 - NFS Storage (Private Subnet)                    ││
│  │ • NFS Server                                              ││
│  │ • Video Storage (Original & Processed)                    ││
│  │ • Assets (intro, outro, logo)                             ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

### AWS Resources
- **3 EC2 instances** (recommended: t3.medium or better)
- **Security Groups** configured
- **VPC** with public and private subnets
- **Elastic IP** (optional, for static webserver IP)
- **Domain name** (optional, for production)

### Software Requirements
- Docker Engine (20.10+)
- Docker Compose (2.0+)
- Git
- NFS utilities

---

## Step 1: Security Group Configuration

### Webserver Security Group
```
Inbound Rules:
- Port 80 (HTTP): 0.0.0.0/0
- Port 443 (HTTPS): 0.0.0.0/0
- Port 22 (SSH): Your IP
- Port 5432 (PostgreSQL): Workers Security Group
- Port 9092 (Kafka): Workers Security Group
- Port 6379 (Redis): Workers Security Group

Outbound Rules:
- All traffic: 0.0.0.0/0
```

### Workers Security Group
```
Inbound Rules:
- Port 22 (SSH): Your IP
- Port 2049 (NFS): Allow from NFS Security Group

Outbound Rules:
- All traffic: 0.0.0.0/0
```

### NFS Security Group
```
Inbound Rules:
- Port 22 (SSH): Your IP
- Port 2049 (NFS): Webserver & Workers Security Groups

Outbound Rules:
- All traffic: 0.0.0.0/0
```

---

## Step 2: EC2 Instance Setup

### Launch 3 EC2 Instances

**Webserver Instance:**
- **AMI**: Amazon Linux 2 or Ubuntu 22.04
- **Type**: t3.medium (2 vCPU, 4 GB RAM) minimum
- **Storage**: 30 GB gp3
- **Public IP**: Enable
- **Security Group**: Webserver SG

**Workers Instance:**
- **AMI**: Amazon Linux 2 or Ubuntu 22.04
- **Type**: t3.large (2 vCPU, 8 GB RAM) minimum
- **Storage**: 20 GB gp3
- **Public IP**: Optional (can use NAT Gateway)
- **Security Group**: Workers SG

**NFS Instance:**
- **AMI**: Amazon Linux 2 or Ubuntu 22.04
- **Type**: t3.small (2 vCPU, 2 GB RAM)
- **Storage**: 50-100 GB gp3 (depends on expected video volume)
- **Public IP**: No
- **Security Group**: NFS SG

---

## Step 3: NFS Server Setup (EC2 #3)

### 3.1. SSH into NFS EC2
```bash
ssh -i your-key.pem ec2-user@<NFS_PRIVATE_IP>
```

### 3.2. Install NFS Server
```bash
# Amazon Linux 2
sudo yum update -y
sudo yum install -y nfs-utils

# Ubuntu
sudo apt update
sudo apt install -y nfs-kernel-server
```

### 3.3. Create Export Directory
```bash
sudo mkdir -p /exports/anb-storage
sudo mkdir -p /exports/anb-storage/videos
sudo mkdir -p /exports/anb-storage/thumbnails
sudo chown -R ec2-user:ec2-user /exports/anb-storage
sudo chmod -R 755 /exports/anb-storage
```

### 3.4. Configure NFS Exports
```bash
sudo nano /etc/exports
```

Add the following (replace with your VPC CIDR):
```
/exports/anb-storage 10.0.0.0/16(rw,sync,no_subtree_check,no_root_squash)
```

### 3.5. Start NFS Service
```bash
# Amazon Linux 2
sudo systemctl enable nfs-server
sudo systemctl start nfs-server
sudo exportfs -ra

# Ubuntu
sudo systemctl enable nfs-kernel-server
sudo systemctl start nfs-kernel-server
sudo exportfs -ra
```

### 3.6. Verify NFS Exports
```bash
sudo exportfs -v
showmount -e localhost
```

---

## Step 4: Webserver Setup (EC2 #1)

### 4.1. SSH into Webserver EC2
```bash
ssh -i your-key.pem ec2-user@<WEBSERVER_PUBLIC_IP>
```

### 4.2. Install Docker & Docker Compose
```bash
# Amazon Linux 2
sudo yum update -y
sudo yum install -y docker git
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ec2-user

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker --version
docker-compose --version
```

### 4.3. Install NFS Client
```bash
# Amazon Linux 2
sudo yum install -y nfs-utils

# Ubuntu
sudo apt install -y nfs-common
```

### 4.4. Mount NFS Storage
```bash
sudo mkdir -p /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage

# Verify mount
df -h | grep nfs
```

### 4.5. Add to /etc/fstab for Persistence
```bash
sudo nano /etc/fstab
```

Add:
```
<NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage nfs defaults,_netdev 0 0
```

### 4.6. Clone Repository
```bash
cd ~
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform
```

### 4.7. Create Environment File
```bash
nano .env.webserver
```

Add:
```bash
# Database
POSTGRES_PASSWORD=your_secure_password_here

# JWT Secret
JWT_SECRET=your_super_secure_jwt_secret_minimum_32_chars

# Webserver IP (replace with actual public IP or domain)
WEBSERVER_PUBLIC_IP=<YOUR_WEBSERVER_PUBLIC_IP_OR_DOMAIN>

# Worker configuration
WORKER_REPLICAS=3
WORKER_CONCURRENCY=2
```

### 4.8. Update Kafka Configuration
Edit `docker-compose.webserver.yml` and replace `<WEBSERVER_PUBLIC_IP>` with your actual IP:
```bash
nano docker-compose.webserver.yml
# Find: KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092,PLAINTEXT_HOST://<WEBSERVER_PUBLIC_IP>:9092
# Replace <WEBSERVER_PUBLIC_IP> with actual IP
```

### 4.9. Update NFS Mount Paths (if different)
If your NFS mount is not at `/mnt/nfs/anb-storage`, update the volume mounts in `docker-compose.webserver.yml`:
```yaml
volumes:
  - /your/nfs/path:/root/storage
```

### 4.10. Copy Assets to NFS
```bash
sudo cp -r backend/assets/* /mnt/nfs/anb-storage/
```

### 4.11. Start Services
```bash
# Load environment variables
export $(cat .env.webserver | xargs)

# Build and start
docker-compose -f docker-compose.webserver.yml up --build -d

# Check logs
docker-compose -f docker-compose.webserver.yml logs -f
```

### 4.12. Verify Services
```bash
# Check running containers
docker ps

# Check backend health
curl http://localhost:8000/api/health

# Check Kafka topics
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

---

## Step 5: Workers Setup (EC2 #2)

### 5.1. SSH into Workers EC2
```bash
ssh -i your-key.pem ec2-user@<WORKERS_PRIVATE_IP>
```

### 5.2. Install Docker & Docker Compose
```bash
# Same as Step 4.2
sudo yum update -y
sudo yum install -y docker git
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ec2-user

sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 5.3. Install NFS Client
```bash
sudo yum install -y nfs-utils
```

### 5.4. Mount NFS Storage
```bash
sudo mkdir -p /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage

# Add to /etc/fstab
echo "<NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage nfs defaults,_netdev 0 0" | sudo tee -a /etc/fstab
```

### 5.5. Clone Repository
```bash
cd ~
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform
```

### 5.6. Create Environment File
```bash
nano .env.workers
```

Add:
```bash
# Webserver private IP (for internal communication)
WEBSERVER_PRIVATE_IP=<WEBSERVER_PRIVATE_IP>

# Database
POSTGRES_PASSWORD=your_secure_password_here

# Worker scaling
WORKER_REPLICAS=3
WORKER_CONCURRENCY=2
```

### 5.7. Update Workers Configuration
Edit `docker-compose.workers.yml` and replace `${WEBSERVER_PRIVATE_IP}`:
```bash
nano docker-compose.workers.yml
# The environment variables will be loaded from .env.workers
```

### 5.8. Start Workers
```bash
# Load environment variables
export $(cat .env.workers | xargs)

# Build and start
docker-compose -f docker-compose.workers.yml up --build -d

# Check logs
docker-compose -f docker-compose.workers.yml logs -f
```

### 5.9. Scale Workers (Optional)
```bash
# Scale to 5 workers
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5

# Check running workers
docker ps | grep video-processor
```

---

## Step 6: Verification & Testing

### 6.1. Test API Endpoints
```bash
# From your local machine or webserver
WEBSERVER_IP=<YOUR_WEBSERVER_PUBLIC_IP>

# Health check
curl http://${WEBSERVER_IP}/api/health

# Sign up
curl -X POST http://${WEBSERVER_IP}/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "Test",
    "lastName": "User",
    "email": "test@example.com",
    "password": "password123",
    "password2": "password123",
    "city": "Bogotá",
    "country": "Colombia"
  }'

# Login
TOKEN=$(curl -X POST http://${WEBSERVER_IP}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}' \
  | jq -r '.token')

echo "Token: $TOKEN"

# Get public videos
curl http://${WEBSERVER_IP}/api/public/videos
```

### 6.2. Test Video Upload
```bash
# Upload a test video
curl -X POST http://${WEBSERVER_IP}/api/videos/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "video=@path/to/your/video.mp4" \
  -F "title=Test Video"

# Check worker logs for processing
ssh ec2-user@<WORKERS_PRIVATE_IP>
cd anb-platform
docker-compose -f docker-compose.workers.yml logs -f video-processor
```

### 6.3. Access Frontend
Open browser: `http://<WEBSERVER_PUBLIC_IP>`

### 6.4. Check OpenAPI Documentation
Open browser: `http://<WEBSERVER_PUBLIC_IP>/swagger/index.html`

---

## Step 7: Monitoring & Maintenance

### 7.1. Monitor Kafka Topics
```bash
# On webserver
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check consumer groups
docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list

# Describe consumer group
docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group video-processors --describe
```

### 7.2. Check Logs
```bash
# Webserver logs
ssh ec2-user@<WEBSERVER_PUBLIC_IP>
cd anb-platform
docker-compose -f docker-compose.webserver.yml logs -f backend

# Worker logs
ssh ec2-user@<WORKERS_PRIVATE_IP>
cd anb-platform
docker-compose -f docker-compose.workers.yml logs -f

# Nginx logs
docker exec anb-nginx tail -f /var/log/nginx/access.log
docker exec anb-nginx tail -f /var/log/nginx/error.log
```

### 7.3. Database Backup
```bash
# On webserver
docker exec anb-postgres pg_dump -U postgres anb_platform > backup_$(date +%Y%m%d).sql

# Copy to S3 (optional)
aws s3 cp backup_$(date +%Y%m%d).sql s3://your-bucket/backups/
```

### 7.4. NFS Storage Usage
```bash
# On NFS server
df -h /exports/anb-storage
du -sh /exports/anb-storage/*
```

---

## Step 8: Production Considerations

### 8.1. SSL/TLS Configuration
1. Obtain SSL certificate (Let's Encrypt or AWS Certificate Manager)
2. Place certificates in `nginx/ssl/` directory
3. Uncomment SSL configuration in `nginx/nginx.conf`
4. Restart nginx: `docker-compose -f docker-compose.webserver.yml restart nginx`

### 8.2. Domain Configuration
1. Point your domain to Webserver Elastic IP
2. Update `nginx/nginx.conf` with your domain name
3. Update CORS settings in backend `cmd/api/main.go`

### 8.3. Increase Resources
- **Webserver**: Upgrade to t3.large or c5.xlarge for production
- **Workers**: Use c5.2xlarge or higher for video processing
- **NFS**: Increase storage as needed, consider EBS gp3 with higher IOPS

### 8.4. Auto-Scaling (Optional)
- Use AWS Auto Scaling Groups for workers
- Configure Launch Templates with workers docker-compose
- Set scaling policies based on Kafka consumer lag

### 8.5. CloudWatch Monitoring
```bash
# Install CloudWatch agent on all instances
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
sudo rpm -U ./amazon-cloudwatch-agent.rpm
```

### 8.6. Security Hardening
- Change default PostgreSQL password
- Use AWS Secrets Manager for sensitive credentials
- Enable AWS Security Hub
- Configure AWS WAF for Nginx
- Enable VPC Flow Logs

---

## Troubleshooting

### Workers Can't Connect to Kafka
```bash
# Check connectivity from workers
telnet <WEBSERVER_PRIVATE_IP> 9092

# Verify Kafka advertised listeners
docker exec anb-kafka kafka-broker-api-versions --bootstrap-server localhost:9092
```

### NFS Mount Issues
```bash
# Check NFS service status
sudo systemctl status nfs-server

# Verify exports
sudo exportfs -v

# Check network connectivity
ping <NFS_PRIVATE_IP>

# Re-mount
sudo umount /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage
```

### Video Processing Stuck
```bash
# Check worker logs
docker-compose -f docker-compose.workers.yml logs video-processor

# Check Kafka consumer lag
docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group video-processors --describe

# Restart workers
docker-compose -f docker-compose.workers.yml restart
```

### Database Connection Issues
```bash
# Check PostgreSQL is accepting connections
docker exec anb-postgres psql -U postgres -c "SELECT version();"

# Verify security group allows port 5432 from workers
```

---

## Scaling Guide

### Scale Workers Horizontally
```bash
# Increase replicas
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5

# Or modify .env.workers
echo "WORKER_REPLICAS=5" >> .env.workers
docker-compose -f docker-compose.workers.yml up -d
```

### Scale Workers Vertically
- Stop workers
- Change EC2 instance type to larger size
- Start workers

### Add More Worker EC2 Instances
1. Launch new EC2 with same configuration
2. Follow Steps 5.1-5.8
3. Workers will automatically join same consumer group

---

## Cleanup

### Stop All Services
```bash
# Webserver
docker-compose -f docker-compose.webserver.yml down

# Workers
docker-compose -f docker-compose.workers.yml down
```

### Remove Volumes
```bash
docker-compose -f docker-compose.webserver.yml down -v
```

### Terminate EC2 Instances
- Stop/Terminate instances from AWS Console
- Delete associated EBS volumes
- Release Elastic IPs

---

## Cost Optimization

- Use **Reserved Instances** for webserver (1-3 year commitment)
- Use **Spot Instances** for workers (70% savings)
- Enable **EBS gp3** with baseline performance
- Use **S3 Glacier** for old video archives
- Configure **AWS Budgets** and alerts

---

## Support & Resources

- **API Documentation**: `http://<YOUR_IP>/swagger/index.html`
- **Repository**: https://github.com/Cloud-2025-2/anb-platform
- **SonarQube**: https://sonarcloud.io/project/overview?id=Cloud-2025-2_anb-platform

---

## Quick Reference Commands

```bash
# Restart all services on webserver
docker-compose -f docker-compose.webserver.yml restart

# Restart workers
docker-compose -f docker-compose.workers.yml restart

# View logs
docker-compose -f docker-compose.webserver.yml logs -f backend
docker-compose -f docker-compose.workers.yml logs -f

# Scale workers
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5

# Database backup
docker exec anb-postgres pg_dump -U postgres anb_platform > backup.sql

# Check NFS mount
df -h | grep nfs

# Monitor Kafka
docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group video-processors --describe
```
