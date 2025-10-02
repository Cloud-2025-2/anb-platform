# User Data Scripts - Quick Reference Card

## 📋 Pre-Launch Checklist

- [ ] VPC created with public and private subnets
- [ ] Security groups configured (see below)
- [ ] SSH key pair created
- [ ] User data scripts edited with your values

---

## 🔒 Security Groups Setup

### 1. NFS Security Group
```
Inbound:  Port 22 (SSH) ← Your IP
          Port 2049 (NFS) ← Webserver SG + Workers SG
Outbound: All traffic
```

### 2. Webserver Security Group
```
Inbound:  Port 22 (SSH) ← Your IP
          Port 80 (HTTP) ← 0.0.0.0/0
          Port 443 (HTTPS) ← 0.0.0.0/0
          Port 5432 (PostgreSQL) ← Workers SG
          Port 9092 (Kafka) ← Workers SG
          Port 6379 (Redis) ← Workers SG
Outbound: All traffic
```

### 3. Workers Security Group
```
Inbound:  Port 22 (SSH) ← Your IP
Outbound: All traffic
```

---

## 🚀 Launch Sequence

### Step 1: Launch NFS (Private Subnet)
```
Instance Type: t3.small
Storage: 50-100GB
Script: scripts/user-data-nfs.sh

EDIT BEFORE PASTING:
  Line 32: VPC_CIDR="10.0.0.0/16"  ← Your VPC CIDR

Wait: 3-5 minutes
Note: NFS Private IP for next steps
```

### Step 2: Launch Webserver (Public Subnet)
```
Instance Type: t3.medium
Storage: 30GB
Enable: Auto-assign Public IP
Script: scripts/user-data-webserver.sh

EDIT BEFORE PASTING (lines 9-12):
  NFS_SERVER_IP="10.0.2.100"  ← From Step 1
  POSTGRES_PASSWORD="SecurePass123!"
  JWT_SECRET="min-32-chars-secret"
  WEBSERVER_PUBLIC_IP="your.domain.com"  ← Or EC2 public IP

Wait: 8-12 minutes
Access: http://<PUBLIC_IP>
```

### Step 3: Launch Workers (Private Subnet)
```
Instance Type: t3.large
Storage: 20GB
Script: scripts/user-data-workers.sh

EDIT BEFORE PASTING (lines 9-14):
  NFS_SERVER_IP="10.0.2.100"  ← From Step 1
  WEBSERVER_PRIVATE_IP="10.0.1.100"  ← From Step 2
  POSTGRES_PASSWORD="SecurePass123!"  ← Match Step 2
  WORKER_REPLICAS="3"
  WORKER_CONCURRENCY="2"

Wait: 5-8 minutes
```

---

## ✅ Verification

### NFS Server
```bash
ssh -i key.pem ec2-user@<NFS_IP>
cat /home/ec2-user/nfs-info.txt
showmount -e localhost
```

### Webserver
```bash
# From your machine
curl http://<PUBLIC_IP>/api/health

# SSH to instance
ssh -i key.pem ec2-user@<PUBLIC_IP>
cat /home/ec2-user/webserver-info.txt
docker ps  # Should show 7 containers
```

### Workers
```bash
ssh -i key.pem ec2-user@<WORKERS_IP>
cat /home/ec2-user/workers-info.txt
docker ps | grep video-processor  # Should show 3 workers
```

---

## 🐛 Troubleshooting

### Check User Data Execution
```bash
sudo tail -f /var/log/user-data.log
sudo tail -f /var/log/cloud-init-output.log
```

### Common Issues

**NFS not mounting**
```bash
# On NFS server
sudo systemctl status nfs-server
sudo exportfs -ra

# Check security group allows port 2049
```

**Webserver services not starting**
```bash
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.webserver.yml logs
docker-compose -f docker-compose.webserver.yml restart
```

**Workers can't connect**
```bash
# Test connectivity
telnet <WEBSERVER_PRIVATE_IP> 5432
telnet <WEBSERVER_PRIVATE_IP> 9092

# Check security groups
# Verify WEBSERVER_PRIVATE_IP in user data script
```

---

## 📝 Post-Deployment

### Scale Workers
```bash
cd /home/ec2-user/anb-platform
docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5
```

### Monitor Kafka (from webserver)
```bash
docker exec anb-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group video-processors --describe
```

### Backup Database (from webserver)
```bash
docker exec anb-postgres pg_dump -U postgres anb_platform > backup.sql
```

---

## 🎯 Access Points

After deployment:
- **Frontend**: `http://<WEBSERVER_PUBLIC_IP>`
- **API Health**: `http://<WEBSERVER_PUBLIC_IP>/api/health`
- **Swagger Docs**: `http://<WEBSERVER_PUBLIC_IP>/swagger/index.html`

---

## ⏱️ Timeline

| Step | Time | Status Check |
|------|------|--------------|
| Launch NFS | 3-5 min | `cat nfs-info.txt` |
| Launch Webserver | 8-12 min | `curl http://<IP>/health` |
| Launch Workers | 5-8 min | `docker ps \| grep worker` |
| **Total** | **~20 min** | **Ready to use!** |

---

## 💰 Instance Costs (us-east-1, on-demand)

| Instance | Type | Monthly |
|----------|------|---------|
| NFS | t3.small | ~$15 |
| Webserver | t3.medium | ~$30 |
| Workers | t3.large | ~$60 |
| **Total** | | **~$105** |

*Use Reserved Instances for 40% savings*
*Use Spot for workers for 70% savings*

---

## 📚 Full Documentation

- **Detailed Guide**: [USER_DATA_DEPLOYMENT_GUIDE.md](USER_DATA_DEPLOYMENT_GUIDE.md)
- **Manual Deployment**: [docs/AWS_DEPLOYMENT.md](docs/AWS_DEPLOYMENT.md)
- **Architecture**: [AWS_DEPLOYMENT_QUICK_START.md](AWS_DEPLOYMENT_QUICK_START.md)
