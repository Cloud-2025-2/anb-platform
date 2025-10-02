# AWS Deployment - Quick Start Guide

## Overview
3 EC2 instances architecture with shared NFS storage.

## Architecture Diagram
```
┌─────────────────────────────────────────────────────┐
│              Internet (Users)                        │
└────────────────────┬────────────────────────────────┘
                     │ HTTP/HTTPS
           ┌─────────▼──────────┐
           │   EC2 #1 - WEB     │  (Public Subnet)
           │  ┌──────────────┐  │
           │  │   Nginx :80  │  │  ◄── Entry Point
           │  └──────┬───────┘  │
           │         │          │
           │  ┌──────▼───────┐  │
           │  │ Frontend :80 │  │
           │  └──────────────┘  │
           │  ┌──────────────┐  │
           │  │Backend :8000 │  │
           │  └──┬────────┬──┘  │
           │     │        │     │
           │  ┌──▼────┐ ┌─▼──┐  │
           │  │Postgres│Redis│  │
           │  └────────┘ └────┘  │
           │  ┌──────────────┐  │
           │  │Kafka+Zookeeper│  │  ◄── Message Queue
           │  └──────┬───────┘  │
           └─────────┼──────────┘
                     │ Kafka Topics
           ┌─────────▼──────────┐
           │   EC2 #2 - WORK    │  (Private Subnet)
           │  ┌──────────────┐  │
           │  │Video Worker 1│  │
           │  ├──────────────┤  │
           │  │Video Worker 2│  │  ◄── Scalable
           │  ├──────────────┤  │
           │  │Video Worker 3│  │
           │  └──────────────┘  │
           └─────────┬──────────┘
                     │ NFS Mount
           ┌─────────▼──────────┐
           │   EC2 #3 - NFS     │  (Private Subnet)
           │  ┌──────────────┐  │
           │  │  NFS Server  │  │
           │  │              │  │
           │  │ /exports/    │  │  ◄── Shared Storage
           │  │  anb-storage │  │
           │  │              │  │
           │  │ • videos     │  │
           │  │ • thumbnails │  │
           │  │ • processed  │  │
           │  └──────────────┘  │
           └────────────────────┘
```

## File Structure
```
anb-platform/
├── docker-compose.webserver.yml    # EC2 #1 (Webserver)
├── docker-compose.workers.yml      # EC2 #2 (Workers)
├── .env.webserver.example          # Webserver config template
├── .env.workers.example            # Workers config template
├── nginx/
│   └── nginx.conf                  # Nginx configuration
├── scripts/
│   ├── setup-nfs.sh                # NFS setup script
│   ├── setup-webserver.sh          # Webserver setup script
│   └── setup-workers.sh            # Workers setup script
└── docs/
    └── AWS_DEPLOYMENT.md           # Full documentation
```

## Quick Setup Steps

### 1. Setup NFS Server (EC2 #3)
```bash
# SSH to NFS instance
ssh -i key.pem ec2-user@<NFS_IP>

# Run NFS setup
sudo bash scripts/setup-nfs.sh
# Enter VPC CIDR when prompted (e.g., 10.0.0.0/16)
```

### 2. Setup Webserver (EC2 #1)
```bash
# SSH to webserver
ssh -i key.pem ec2-user@<WEBSERVER_IP>

# Install Docker & Docker Compose
sudo yum update -y
sudo yum install -y docker git nfs-utils
sudo systemctl start docker
sudo usermod -aG docker ec2-user
# Log out and back in

# Clone repository
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform

# Mount NFS
sudo mkdir -p /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage

# Configure environment
cp .env.webserver.example .env.webserver
nano .env.webserver
# Set: POSTGRES_PASSWORD, JWT_SECRET, WEBSERVER_PUBLIC_IP

# Run setup script
bash scripts/setup-webserver.sh
```

### 3. Setup Workers (EC2 #2)
```bash
# SSH to workers
ssh -i key.pem ec2-user@<WORKERS_IP>

# Install Docker & Docker Compose
sudo yum update -y
sudo yum install -y docker git nfs-utils
sudo systemctl start docker
sudo usermod -aG docker ec2-user
# Log out and back in

# Clone repository
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform

# Mount NFS
sudo mkdir -p /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_PRIVATE_IP>:/exports/anb-storage /mnt/nfs/anb-storage

# Configure environment
cp .env.workers.example .env.workers
nano .env.workers
# Set: WEBSERVER_PRIVATE_IP, POSTGRES_PASSWORD

# Run setup script
bash scripts/setup-workers.sh
```

## Security Group Rules

### Webserver SG (EC2 #1)
| Type | Port | Source | Description |
|------|------|--------|-------------|
| HTTP | 80 | 0.0.0.0/0 | Public access |
| HTTPS | 443 | 0.0.0.0/0 | Public access (SSL) |
| PostgreSQL | 5432 | Workers SG | DB access |
| Kafka | 9092 | Workers SG | Kafka access |
| Redis | 6379 | Workers SG | Cache access |
| SSH | 22 | Your IP | Admin access |

### Workers SG (EC2 #2)
| Type | Port | Source | Description |
|------|------|--------|-------------|
| NFS | 2049 | NFS SG | Storage access |
| SSH | 22 | Your IP | Admin access |

### NFS SG (EC2 #3)
| Type | Port | Source | Description |
|------|------|--------|-------------|
| NFS | 2049 | Webserver SG + Workers SG | NFS exports |
| SSH | 22 | Your IP | Admin access |

## Environment Variables

### .env.webserver
```bash
POSTGRES_PASSWORD=SecurePassword123!
JWT_SECRET=your-32-char-minimum-secret-key-here
WEBSERVER_PUBLIC_IP=ec2-xx-xxx-xxx-xxx.compute.amazonaws.com
```

### .env.workers
```bash
WEBSERVER_PRIVATE_IP=10.0.1.100
POSTGRES_PASSWORD=SecurePassword123!
WORKER_REPLICAS=3
WORKER_CONCURRENCY=2
```

## Verification

### Check Webserver
```bash
# Health check
curl http://<WEBSERVER_IP>/api/health

# Check containers
docker ps

# Check logs
docker-compose -f docker-compose.webserver.yml logs -f backend
```

### Check Workers
```bash
# Check containers
docker ps | grep video-processor

# Check logs
docker-compose -f docker-compose.workers.yml logs -f

# Monitor Kafka consumer (from webserver)
docker exec anb-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group video-processors --describe
```

### Check NFS
```bash
# On NFS server
showmount -e localhost
df -h /exports/anb-storage

# On clients
df -h | grep nfs
ls -la /mnt/nfs/anb-storage
```

## Scaling Workers
```bash
# On workers EC2
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5

# Verify
docker ps | grep video-processor | wc -l
```

## Access Points
- **Frontend**: http://\<WEBSERVER_IP\>
- **API**: http://\<WEBSERVER_IP\>/api/health
- **Swagger**: http://\<WEBSERVER_IP\>/swagger/index.html

## Troubleshooting

### Workers can't connect to Kafka
```bash
# Test from workers
telnet <WEBSERVER_PRIVATE_IP> 9092

# Check Kafka config on webserver
docker exec anb-kafka cat /etc/kafka/server.properties | grep advertised
```

### NFS mount fails
```bash
# Check NFS exports
showmount -e <NFS_IP>

# Check network connectivity
ping <NFS_IP>

# Remount
sudo umount /mnt/nfs/anb-storage
sudo mount -t nfs <NFS_IP>:/exports/anb-storage /mnt/nfs/anb-storage
```

### Videos not processing
```bash
# Check Kafka topics
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check consumer lag
docker exec anb-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group video-processors --describe

# Restart workers
docker-compose -f docker-compose.workers.yml restart
```

## Cost Estimate (us-east-1)

| Resource | Type | Monthly Cost |
|----------|------|--------------|
| Webserver | t3.medium | ~$30 |
| Workers | t3.large | ~$60 |
| NFS | t3.small + 50GB EBS | ~$20 |
| Data Transfer | 100GB/month | ~$9 |
| **Total** | | **~$119/month** |

*Use Reserved Instances or Spot for 40-70% savings*

## Next Steps
1. ✅ Setup complete - Test video upload
2. 🔒 Configure SSL/TLS with Let's Encrypt
3. 📊 Setup CloudWatch monitoring
4. 🔐 Move secrets to AWS Secrets Manager
5. 🚀 Configure auto-scaling for workers

## Support
- Full Documentation: [AWS_DEPLOYMENT.md](docs/AWS_DEPLOYMENT.md)
- Repository: https://github.com/Cloud-2025-2/anb-platform
- Issues: Create GitHub issue
