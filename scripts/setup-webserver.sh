#!/bin/bash
# ANB Platform - Webserver Setup Script
# Run this script on EC2 Webserver instance after basic setup

set -e

echo "========================================="
echo "ANB Platform - Webserver Setup"
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
if [ ! -f .env.webserver ]; then
    echo -e "${RED}Error: .env.webserver not found${NC}"
    echo "Please copy .env.webserver.example to .env.webserver and configure it"
    exit 1
fi

# Load environment variables
source .env.webserver

# Validate required variables
if [ -z "$POSTGRES_PASSWORD" ] || [ "$POSTGRES_PASSWORD" == "your_secure_password_here_change_this" ]; then
    echo -e "${RED}Error: POSTGRES_PASSWORD not configured in .env.webserver${NC}"
    exit 1
fi

if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" == "your_super_secure_jwt_secret_minimum_32_chars_change_this_in_production" ]; then
    echo -e "${RED}Error: JWT_SECRET not configured in .env.webserver${NC}"
    exit 1
fi

if [ -z "$WEBSERVER_PUBLIC_IP" ] || [ "$WEBSERVER_PUBLIC_IP" == "your-ec2-public-ip-or-domain.com" ]; then
    echo -e "${RED}Error: WEBSERVER_PUBLIC_IP not configured in .env.webserver${NC}"
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

# Copy assets to NFS
echo "Copying assets to NFS storage..."
if [ -d "backend/assets" ]; then
    sudo cp -r backend/assets/* /mnt/nfs/anb-storage/ 2>/dev/null || true
    echo -e "${GREEN}✓ Assets copied to NFS${NC}"
else
    echo -e "${YELLOW}Warning: backend/assets directory not found${NC}"
fi

# Update Kafka advertised listeners in docker-compose
echo "Updating Kafka configuration..."
sed -i.bak "s/<WEBSERVER_PUBLIC_IP>/${WEBSERVER_PUBLIC_IP}/g" docker-compose.webserver.yml
echo -e "${GREEN}✓ Kafka configuration updated${NC}"

# Build and start services
echo "Building and starting Docker services..."
echo "This may take several minutes..."

docker-compose -f docker-compose.webserver.yml up --build -d

echo ""
echo "Waiting for services to start..."
sleep 30

# Check service health
echo ""
echo "Checking service health..."

# Check Postgres
if docker exec anb-postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is running${NC}"
else
    echo -e "${RED}✗ PostgreSQL is not responding${NC}"
fi

# Check Redis
if docker exec anb-redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is running${NC}"
else
    echo -e "${RED}✗ Redis is not responding${NC}"
fi

# Check Backend
if curl -f http://localhost:8000/api/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Backend API is running${NC}"
else
    echo -e "${YELLOW}⚠ Backend API is not responding yet (may still be starting)${NC}"
fi

# Check Kafka topics
echo ""
echo "Creating Kafka topics..."
sleep 5
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing --partitions 3 --replication-factor 1 --if-not-exists || true
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-retry --partitions 3 --replication-factor 1 --if-not-exists || true
docker exec anb-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-dlq --partitions 1 --replication-factor 1 --if-not-exists || true
echo -e "${GREEN}✓ Kafka topics created${NC}"

echo ""
echo "========================================="
echo -e "${GREEN}Setup completed!${NC}"
echo "========================================="
echo ""
echo "Access your application:"
echo "  Frontend:  http://${WEBSERVER_PUBLIC_IP}"
echo "  API:       http://${WEBSERVER_PUBLIC_IP}/api/health"
echo "  Swagger:   http://${WEBSERVER_PUBLIC_IP}/swagger/index.html"
echo ""
echo "View logs:"
echo "  docker-compose -f docker-compose.webserver.yml logs -f"
echo ""
echo "Next steps:"
echo "  1. Setup workers on EC2 Workers instance"
echo "  2. Test video upload functionality"
echo "  3. Configure SSL/TLS for production"
echo ""
