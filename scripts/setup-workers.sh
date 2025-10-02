#!/bin/bash
# ANB Platform - Workers Setup Script
# Run this script on EC2 Workers instance after basic setup

set -e

echo "========================================="
echo "ANB Platform - Workers Setup"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
    echo -e "${RED}Please do not run as root${NC}"
    exit 1
fi

# Check for required variables
if [ ! -f .env.workers ]; then
    echo -e "${RED}Error: .env.workers not found${NC}"
    echo "Please copy .env.workers.example to .env.workers and configure it"
    exit 1
fi

# Load environment variables
source .env.workers

# Validate required variables
if [ -z "$WEBSERVER_PRIVATE_IP" ] || [ "$WEBSERVER_PRIVATE_IP" == "10.0.1.100" ]; then
    echo -e "${RED}Error: WEBSERVER_PRIVATE_IP not configured in .env.workers${NC}"
    exit 1
fi

if [ -z "$POSTGRES_PASSWORD" ] || [ "$POSTGRES_PASSWORD" == "your_secure_password_here_change_this" ]; then
    echo -e "${RED}Error: POSTGRES_PASSWORD not configured in .env.workers${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Environment variables validated${NC}"

# Check if NFS is mounted
if ! mountpoint -q /mnt/nfs/anb-storage; then
    echo -e "${YELLOW}Warning: NFS not mounted at /mnt/nfs/anb-storage${NC}"
    echo "Please mount NFS before continuing:"
    echo "  sudo mkdir -p /mnt/nfs/anb-storage"
    echo "  sudo mount -t nfs <NFS_IP>:/exports/anb-storage /mnt/nfs/anb-storage"
    exit 1
fi

echo -e "${GREEN}✓ NFS storage mounted${NC}"

# Test connectivity to webserver
echo "Testing connectivity to webserver..."

# Test PostgreSQL connection
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${WEBSERVER_PRIVATE_IP}/5432" 2>/dev/null; then
    echo -e "${GREEN}✓ PostgreSQL port (5432) is accessible${NC}"
else
    echo -e "${RED}✗ Cannot connect to PostgreSQL on ${WEBSERVER_PRIVATE_IP}:5432${NC}"
    echo "  Please check security group rules"
    exit 1
fi

# Test Kafka connection
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${WEBSERVER_PRIVATE_IP}/9092" 2>/dev/null; then
    echo -e "${GREEN}✓ Kafka port (9092) is accessible${NC}"
else
    echo -e "${RED}✗ Cannot connect to Kafka on ${WEBSERVER_PRIVATE_IP}:9092${NC}"
    echo "  Please check security group rules"
    exit 1
fi

# Update docker-compose with environment variables
echo "Updating worker configuration..."
export WEBSERVER_PRIVATE_IP POSTGRES_PASSWORD WORKER_REPLICAS WORKER_CONCURRENCY

# Build and start workers
echo "Building and starting worker services..."
echo "This may take several minutes..."

docker-compose -f docker-compose.workers.yml up --build -d

echo ""
echo "Waiting for workers to start..."
sleep 20

# Check worker status
echo ""
echo "Checking worker status..."
WORKER_COUNT=$(docker ps --filter "name=video-processor" --format "{{.Names}}" | wc -l)

if [ "$WORKER_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ ${WORKER_COUNT} worker(s) running${NC}"
    docker ps --filter "name=video-processor" --format "table {{.Names}}\t{{.Status}}"
else
    echo -e "${RED}✗ No workers running${NC}"
    echo "Check logs: docker-compose -f docker-compose.workers.yml logs"
    exit 1
fi

echo ""
echo "========================================="
echo -e "${GREEN}Workers setup completed!${NC}"
echo "========================================="
echo ""
echo "Worker information:"
echo "  Replicas:    ${WORKER_REPLICAS:-3}"
echo "  Concurrency: ${WORKER_CONCURRENCY:-2}"
echo "  Group ID:    video-processors"
echo ""
echo "View logs:"
echo "  docker-compose -f docker-compose.workers.yml logs -f"
echo ""
echo "Scale workers:"
echo "  docker-compose -f docker-compose.workers.yml up -d --scale video-processor=5"
echo ""
echo "Monitor Kafka consumer group (from webserver):"
echo "  docker exec anb-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group video-processors --describe"
echo ""
