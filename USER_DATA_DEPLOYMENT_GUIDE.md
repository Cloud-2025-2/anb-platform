# AWS User Data Deployment Guide

## Overview

This guide shows you how to deploy the ANB Platform using **EC2 User Data scripts** for **fully automated setup**. Just paste the scripts when launching instances and they'll configure themselves automatically.

## 🚀 Zero-Configuration Deployment

### What Gets Automated:
- ✅ System updates and package installation
- ✅ Docker & Docker Compose installation
- ✅ NFS server setup and client mounting
- ✅ Repository cloning
- ✅ Service configuration and startup
- ✅ No manual SSH required!

---

## Prerequisites

1. **AWS VPC** configured with:
   - Public subnet (for webserver)
   - Private subnets (for workers and NFS)
   - Internet Gateway attached
   - NAT Gateway (if workers need internet)

2. **Security Groups** configured (see below)

3. **AWS Account** with EC2 launch permissions

---

## Security Group Configuration

### NFS Security Group
```
Inbound:
- Port 22 (SSH): Your IP - for admin access
- Port 2049 (NFS): Webserver SG + Workers SG

Outbound:
- All traffic: 0.0.0.0/0
```

### Webserver Security Group
```
Inbound:
- Port 22 (SSH): Your IP - for admin access
- Port 80 (HTTP): 0.0.0.0/0 - public access
- Port 443 (HTTPS): 0.0.0.0/0 - public access (future)
- Port 5432 (PostgreSQL): Workers SG
- Port 9092 (Kafka): Workers SG
- Port 6379 (Redis): Workers SG

Outbound:
- All traffic: 0.0.0.0/0
```

### Workers Security Group
```
Inbound:
- Port 22 (SSH): Your IP - for admin access

Outbound:
- All traffic: 0.0.0.0/0
```

---

## Step 1: Launch NFS Instance

### 1.1 EC2 Configuration
- **AMI**: Amazon Linux 2
- **Instance Type**: t3.small (2 vCPU, 2GB RAM)
- **Storage**: 50-100GB gp3 (adjust based on expected video volume)
- **Network**: Place in **private subnet**
- **Security Group**: NFS Security Group
- **Key Pair**: Your SSH key

### 1.2 User Data Script

1. Open `scripts/user-data-nfs.sh`
2. **IMPORTANT**: Update the VPC CIDR on line 32:
   ```bash
   VPC_CIDR="10.0.0.0/16"  # Change to your VPC CIDR
   ```
3. Copy the **entire script**
4. In EC2 launch wizard, scroll to **Advanced Details** → **User data**
5. Paste the script
6. Launch instance

### 1.3 Wait for Setup
- Setup takes ~3-5 minutes
- SSH in and check: `cat /home/ec2-user/nfs-info.txt`
- Note the **NFS Server Private IP** (you'll need it for next steps)

---

## Step 2: Launch Webserver Instance

### 2.1 EC2 Configuration
- **AMI**: Amazon Linux 2
- **Instance Type**: t3.medium or better (2 vCPU, 4GB RAM minimum)
- **Storage**: 30GB gp3
- **Network**: Place in **public subnet**
- **Auto-assign Public IP**: Enable
- **Security Group**: Webserver Security Group
- **Key Pair**: Your SSH key

### 2.2 User Data Script

1. Open `scripts/user-data-webserver.sh`
2. **IMPORTANT**: Edit the configuration section (lines 9-12):
   ```bash
   NFS_SERVER_IP="10.0.2.100"  # Replace with NFS private IP from Step 1
   POSTGRES_PASSWORD="YourSecurePassword123!"  # Set strong password
   JWT_SECRET="your-super-secure-jwt-secret-min-32-chars"  # Set secure secret
   WEBSERVER_PUBLIC_IP="ec2-xx-xxx.compute.amazonaws.com"  # Use EC2 public IP or domain
   ```
3. Copy the **entire modified script**
4. In EC2 launch wizard, scroll to **Advanced Details** → **User data**
5. Paste the script
6. Launch instance

### 2.3 Wait for Setup
- Setup takes ~8-12 minutes (Docker builds, services start)
- Monitor progress: Check **EC2 Console → Instance → Actions → Monitor and troubleshoot → Get system log**
- Once complete, SSH in and check: `cat /home/ec2-user/webserver-info.txt`

### 2.4 Verify Webserver
```bash
# Get public IP from AWS Console
PUBLIC_IP=<your-webserver-public-ip>

# Test endpoints
curl http://${PUBLIC_IP}/health
curl http://${PUBLIC_IP}/api/health

# Open in browser
# Frontend: http://${PUBLIC_IP}
# Swagger: http://${PUBLIC_IP}/swagger/index.html
```

---

## Step 3: Launch Workers Instance

### 3.1 EC2 Configuration
- **AMI**: Amazon Linux 2
- **Instance Type**: t3.large or better (2 vCPU, 8GB RAM minimum)
- **Storage**: 20GB gp3
- **Network**: Place in **private subnet** (or public if no NAT)
- **Security Group**: Workers Security Group
- **Key Pair**: Your SSH key

### 3.2 User Data Script

1. Open `scripts/user-data-workers.sh`
2. **IMPORTANT**: Edit the configuration section (lines 9-14):
   ```bash
   NFS_SERVER_IP="10.0.2.100"  # Replace with NFS private IP
   WEBSERVER_PRIVATE_IP="10.0.1.100"  # Replace with webserver PRIVATE IP
   POSTGRES_PASSWORD="YourSecurePassword123!"  # Must match webserver password
   WORKER_REPLICAS="3"  # Number of worker containers
   WORKER_CONCURRENCY="2"  # Tasks per worker
   ```
3. Copy the **entire modified script**
4. In EC2 launch wizard, scroll to **Advanced Details** → **User data**
5. Paste the script
6. Launch instance

### 3.3 Wait for Setup
- Setup takes ~5-8 minutes
- SSH in and check: `cat /home/ec2-user/workers-info.txt`

### 3.4 Verify Workers
```bash
# SSH to workers instance
ssh -i your-key.pem ec2-user@<workers-private-ip>

# Check running workers
docker ps | grep video-processor

# Check logs
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.workers.yml logs -f
```

---

## Verification & Testing

### End-to-End Test

1. **Register a user**:
   ```bash
   curl -X POST http://<WEBSERVER_IP>/api/auth/signup \
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
   ```

2. **Login**:
   ```bash
   TOKEN=$(curl -X POST http://<WEBSERVER_IP>/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123"}' \
     | jq -r '.token')
   ```

3. **Upload a video** (use frontend or API):
   - Frontend: http://<WEBSERVER_IP> → Login → Upload
   - Watch worker logs for processing

4. **Monitor Kafka** (from webserver):
   ```bash
   docker exec anb-kafka kafka-consumer-groups \
     --bootstrap-server localhost:9092 \
     --group video-processors --describe
   ```

---

## Troubleshooting

### Check User Data Logs

**All instances**:
```bash
# View user data execution log
sudo tail -f /var/log/user-data.log

# Or check cloud-init logs
sudo tail -f /var/log/cloud-init-output.log
```

### NFS Issues

**Problem**: Workers/Webserver can't mount NFS
```bash
# On NFS server, verify exports
showmount -e localhost

# Check if NFS is running
systemctl status nfs-server

# Verify security group allows port 2049 from clients
```

**Fix**:
```bash
# Re-export
sudo exportfs -ra

# Restart NFS
sudo systemctl restart nfs-server
```

### Webserver Not Starting

**Problem**: Services not running
```bash
# Check Docker
sudo systemctl status docker

# Check logs
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.webserver.yml logs

# Restart services
docker-compose -f docker-compose.webserver.yml restart
```

### Workers Can't Connect

**Problem**: Workers can't reach Kafka/PostgreSQL
```bash
# Test connectivity
telnet <WEBSERVER_PRIVATE_IP> 5432
telnet <WEBSERVER_PRIVATE_IP> 9092

# Check security groups allow traffic
# Verify WEBSERVER_PRIVATE_IP is correct in user data
```

**Fix**: Update security group rules or edit `.env.workers` and restart

---

## Post-Deployment

### Scale Workers
```bash
# SSH to workers instance
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5
```

### Setup SSL/TLS (Production)
1. Obtain SSL certificate (Let's Encrypt)
2. Copy to webserver: `/home/ec2-user/anb-platform/nginx/ssl/`
3. Uncomment SSL lines in `nginx/nginx.conf`
4. Restart nginx: `docker-compose -f docker-compose.webserver.yml restart nginx`

### Monitor Resources
```bash
# Check disk space
df -h

# Check memory
free -h

# Check Docker stats
docker stats
```

### Backup Database
```bash
# On webserver
docker exec anb-postgres pg_dump -U postgres anb_platform > backup.sql

# Upload to S3 (optional)
aws s3 cp backup.sql s3://your-bucket/backups/
```

---

## Cost Optimization

- **Reserved Instances**: Commit 1-3 years for webserver (40% savings)
- **Spot Instances**: Use for workers (70% savings)
- **Auto-Scaling**: Scale workers based on Kafka lag
- **Schedule**: Stop dev instances overnight

**Estimated Monthly Cost** (us-east-1, on-demand):
- NFS (t3.small): ~$15
- Webserver (t3.medium): ~$30
- Workers (t3.large): ~$60
- **Total**: ~$105/month (before optimizations)

---

## Instance Information Files

After setup, each instance has an info file:

- **NFS**: `/home/ec2-user/nfs-info.txt`
- **Webserver**: `/home/ec2-user/webserver-info.txt`
- **Workers**: `/home/ec2-user/workers-info.txt`

These contain IP addresses, commands, and access information.

---

## Summary

✅ **NFS Server** - Fully automated NFS setup
✅ **Webserver** - Complete stack running (Nginx, Backend, Frontend, DBs, Kafka)
✅ **Workers** - Video processors connected and ready

**Total Deployment Time**: ~20-25 minutes from launch to fully operational

**Manual Steps Required**: Only editing IPs and passwords in user data scripts before pasting

---

## Quick Reference

| Component | Instance Type | Subnet | User Data Script |
|-----------|---------------|---------|------------------|
| NFS Server | t3.small | Private | `user-data-nfs.sh` |
| Webserver | t3.medium+ | Public | `user-data-webserver.sh` |
| Workers | t3.large+ | Private | `user-data-workers.sh` |

**Launch Order**: NFS → Webserver → Workers

**Access**: `http://<WEBSERVER_PUBLIC_IP>`
