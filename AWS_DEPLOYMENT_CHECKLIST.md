# AWS Deployment Checklist

Use this checklist to ensure all steps are completed for AWS deployment.

## Pre-Deployment

### AWS Account Setup
- [ ] AWS account created and configured
- [ ] AWS CLI installed and configured (optional)
- [ ] VPC created with CIDR (e.g., 10.0.0.0/16)
- [ ] Public subnet created (for webserver)
- [ ] Private subnet created (for workers and NFS)
- [ ] Internet Gateway attached to VPC
- [ ] Route tables configured
- [ ] NAT Gateway created (if workers need internet access)

### Security Groups
- [ ] Webserver security group created
  - [ ] Port 80 (HTTP) - Inbound from 0.0.0.0/0
  - [ ] Port 443 (HTTPS) - Inbound from 0.0.0.0/0
  - [ ] Port 22 (SSH) - Inbound from your IP
  - [ ] Port 5432 (PostgreSQL) - Inbound from Workers SG
  - [ ] Port 9092 (Kafka) - Inbound from Workers SG
  - [ ] Port 6379 (Redis) - Inbound from Workers SG
- [ ] Workers security group created
  - [ ] Port 22 (SSH) - Inbound from your IP
  - [ ] Port 2049 (NFS) - Outbound to NFS SG
- [ ] NFS security group created
  - [ ] Port 22 (SSH) - Inbound from your IP
  - [ ] Port 2049 (NFS) - Inbound from Webserver SG and Workers SG

### EC2 Instances
- [ ] Webserver EC2 launched (t3.medium or better)
  - [ ] AMI: Amazon Linux 2 or Ubuntu 22.04
  - [ ] 30GB storage minimum
  - [ ] Public IP enabled
  - [ ] Key pair assigned
- [ ] Workers EC2 launched (t3.large or better)
  - [ ] AMI: Amazon Linux 2 or Ubuntu 22.04
  - [ ] 20GB storage minimum
  - [ ] Key pair assigned
- [ ] NFS EC2 launched (t3.small or better)
  - [ ] AMI: Amazon Linux 2 or Ubuntu 22.04
  - [ ] 50-100GB storage for videos
  - [ ] Key pair assigned
- [ ] Elastic IP allocated and attached to webserver (optional)

### Domain & SSL (Production)
- [ ] Domain name purchased and configured
- [ ] DNS records pointing to webserver
- [ ] SSL certificate obtained (Let's Encrypt or ACM)

---

## EC2 #3: NFS Server Setup

- [ ] SSH into NFS instance
- [ ] Update system: `sudo yum update -y`
- [ ] Install NFS server: `sudo yum install -y nfs-utils`
- [ ] Clone repository: `git clone https://github.com/Cloud-2025-2/anb-platform.git`
- [ ] Run NFS setup script: `sudo bash scripts/setup-nfs.sh`
- [ ] Enter VPC CIDR when prompted
- [ ] Verify exports: `showmount -e localhost`
- [ ] Note down NFS private IP: ___________________

---

## EC2 #1: Webserver Setup

### System Setup
- [ ] SSH into webserver instance
- [ ] Update system: `sudo yum update -y`
- [ ] Install Docker: `sudo yum install -y docker`
- [ ] Start Docker: `sudo systemctl start docker && sudo systemctl enable docker`
- [ ] Add user to docker group: `sudo usermod -aG docker ec2-user`
- [ ] Log out and back in
- [ ] Install Docker Compose
- [ ] Install NFS client: `sudo yum install -y nfs-utils`
- [ ] Install Git: `sudo yum install -y git`

### NFS Mount
- [ ] Create mount point: `sudo mkdir -p /mnt/nfs/anb-storage`
- [ ] Mount NFS: `sudo mount -t nfs <NFS_IP>:/exports/anb-storage /mnt/nfs/anb-storage`
- [ ] Verify mount: `df -h | grep nfs`
- [ ] Add to /etc/fstab for persistence
- [ ] Test mount after reboot (optional)

### Application Setup
- [ ] Clone repository: `git clone https://github.com/Cloud-2025-2/anb-platform.git`
- [ ] Navigate to directory: `cd anb-platform`
- [ ] Copy environment template: `cp .env.webserver.example .env.webserver`
- [ ] Edit environment file: `nano .env.webserver`
  - [ ] Set POSTGRES_PASSWORD (strong password)
  - [ ] Set JWT_SECRET (32+ characters)
  - [ ] Set WEBSERVER_PUBLIC_IP (EC2 public IP or domain)
- [ ] Copy assets to NFS: `sudo cp -r backend/assets/* /mnt/nfs/anb-storage/`
- [ ] Run setup script: `bash scripts/setup-webserver.sh`
- [ ] Wait for services to start (2-3 minutes)

### Verification
- [ ] Check containers: `docker ps` (should show 7 containers)
- [ ] Test API health: `curl http://localhost:8000/api/health`
- [ ] Test public access: `curl http://<PUBLIC_IP>/api/health`
- [ ] Check Kafka topics: `docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --list`
- [ ] Access Swagger: `http://<PUBLIC_IP>/swagger/index.html`
- [ ] Access frontend: `http://<PUBLIC_IP>`

---

## EC2 #2: Workers Setup

### System Setup
- [ ] SSH into workers instance
- [ ] Update system: `sudo yum update -y`
- [ ] Install Docker: `sudo yum install -y docker`
- [ ] Start Docker: `sudo systemctl start docker && sudo systemctl enable docker`
- [ ] Add user to docker group: `sudo usermod -aG docker ec2-user`
- [ ] Log out and back in
- [ ] Install Docker Compose
- [ ] Install NFS client: `sudo yum install -y nfs-utils`
- [ ] Install Git: `sudo yum install -y git`

### NFS Mount
- [ ] Create mount point: `sudo mkdir -p /mnt/nfs/anb-storage`
- [ ] Mount NFS: `sudo mount -t nfs <NFS_IP>:/exports/anb-storage /mnt/nfs/anb-storage`
- [ ] Verify mount: `df -h | grep nfs`
- [ ] Add to /etc/fstab for persistence

### Application Setup
- [ ] Clone repository: `git clone https://github.com/Cloud-2025-2/anb-platform.git`
- [ ] Navigate to directory: `cd anb-platform`
- [ ] Copy environment template: `cp .env.workers.example .env.workers`
- [ ] Edit environment file: `nano .env.workers`
  - [ ] Set WEBSERVER_PRIVATE_IP (webserver private IP)
  - [ ] Set POSTGRES_PASSWORD (same as webserver)
  - [ ] Set WORKER_REPLICAS (default: 3)
  - [ ] Set WORKER_CONCURRENCY (default: 2)
- [ ] Run setup script: `bash scripts/setup-workers.sh`
- [ ] Wait for workers to start (1-2 minutes)

### Verification
- [ ] Check containers: `docker ps | grep video-processor`
- [ ] Check logs: `docker-compose -f docker-compose.workers.yml logs -f`
- [ ] Test connectivity to webserver PostgreSQL: `telnet <WEBSERVER_PRIVATE_IP> 5432`
- [ ] Test connectivity to Kafka: `telnet <WEBSERVER_PRIVATE_IP> 9092`

---

## Integration Testing

### User Registration & Login
- [ ] Register new user via API or frontend
- [ ] Login and receive JWT token
- [ ] Verify token works on protected endpoints

### Video Upload & Processing
- [ ] Upload test video via frontend or API
- [ ] Check video status changes to "processing"
- [ ] Monitor worker logs: `docker-compose -f docker-compose.workers.yml logs -f`
- [ ] Wait for processing to complete (status: "processed")
- [ ] Verify processed video is accessible
- [ ] Check thumbnail generation

### Voting & Rankings
- [ ] Publish a processed video
- [ ] Vote on public video
- [ ] Check rankings endpoint
- [ ] Verify vote count updates

### Kafka Monitoring
- [ ] Check consumer group status:
  ```bash
  docker exec anb-kafka kafka-consumer-groups \
    --bootstrap-server localhost:9092 \
    --group video-processors --describe
  ```
- [ ] Verify no lag in message consumption
- [ ] Check DLQ for failed messages

---

## Performance & Monitoring

### Load Testing
- [ ] Run Postman collection tests
- [ ] Upload multiple videos concurrently
- [ ] Monitor resource usage (CPU, memory, disk)
- [ ] Scale workers if needed: `docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5`

### Monitoring Setup
- [ ] Setup CloudWatch agent (optional)
- [ ] Configure CloudWatch alarms for:
  - [ ] High CPU usage
  - [ ] High memory usage
  - [ ] Disk space alerts
  - [ ] Application errors
- [ ] Setup log aggregation (CloudWatch Logs)

---

## Production Readiness

### Security
- [ ] Change default passwords
- [ ] Move secrets to AWS Secrets Manager or Parameter Store
- [ ] Configure SSL/TLS certificates
- [ ] Update nginx.conf for HTTPS
- [ ] Enable HTTPS redirect
- [ ] Configure WAF rules (optional)
- [ ] Enable VPC Flow Logs
- [ ] Review and tighten security group rules
- [ ] Setup AWS Systems Manager Session Manager (passwordless SSH)

### Backup & Recovery
- [ ] Setup PostgreSQL automated backups
- [ ] Configure RDS snapshot schedule (if migrated to RDS)
- [ ] Backup NFS storage to S3: `aws s3 sync /exports/anb-storage s3://your-backup-bucket/`
- [ ] Test restore procedure
- [ ] Document recovery steps

### High Availability (Optional)
- [ ] Setup Application Load Balancer
- [ ] Configure multiple webserver instances
- [ ] Migrate PostgreSQL to RDS
- [ ] Migrate Redis to ElastiCache
- [ ] Migrate Kafka to MSK (Managed Streaming for Kafka)
- [ ] Migrate NFS to EFS (Elastic File System)
- [ ] Setup auto-scaling groups for workers
- [ ] Configure health checks

### Documentation
- [ ] Document instance IPs and credentials
- [ ] Update DNS records
- [ ] Create runbook for common issues
- [ ] Train team on deployment procedures
- [ ] Document scaling procedures
- [ ] Create incident response plan

---

## Post-Deployment

### Monitoring
- [ ] Monitor logs for errors (first 24 hours)
- [ ] Check video processing success rate
- [ ] Monitor Kafka consumer lag
- [ ] Check disk space on NFS
- [ ] Review CloudWatch metrics

### Optimization
- [ ] Analyze video processing times
- [ ] Optimize FFmpeg settings if needed
- [ ] Tune PostgreSQL parameters
- [ ] Adjust worker replicas based on load
- [ ] Configure Redis cache TTL based on usage
- [ ] Enable CloudFront CDN for video delivery (optional)

### Cost Optimization
- [ ] Review AWS Cost Explorer
- [ ] Purchase Reserved Instances for predictable workloads
- [ ] Use Spot Instances for workers
- [ ] Setup budget alerts
- [ ] Configure S3 lifecycle policies for old videos
- [ ] Right-size instances based on actual usage

---

## Troubleshooting Guide

### Common Issues
- [ ] Workers can't connect to Kafka → Check security groups and Kafka advertised listeners
- [ ] NFS mount fails → Check security groups, verify NFS server is running
- [ ] Videos stuck in processing → Check worker logs, restart workers
- [ ] Frontend shows 404 for videos → Verify NFS mount on frontend container
- [ ] High CPU on workers → Scale workers or upgrade instance type
- [ ] Database connection errors → Check PostgreSQL max_connections setting

---

## Rollback Plan

In case of issues:
- [ ] Document current state
- [ ] Stop new deployments
- [ ] Identify root cause
- [ ] Revert to previous configuration if needed:
  ```bash
  git checkout <previous-commit>
  docker-compose -f docker-compose.webserver.yml up --build -d
  docker-compose -f docker-compose.workers.yml up --build -d
  ```
- [ ] Verify functionality
- [ ] Document lessons learned

---

## Sign-Off

Deployment completed by: _____________________

Date: _____________________

Verified by: _____________________

Production URL: _____________________

Notes:
_____________________________________________
_____________________________________________
_____________________________________________
