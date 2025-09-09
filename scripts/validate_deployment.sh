#!/bin/bash

# ANB Platform Deployment Validation Script
# This script validates that all services are running correctly

echo "🚀 ANB Platform Deployment Validation"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if service is running
check_service() {
    local service_name=$1
    local port=$2
    
    if docker-compose ps | grep -q "$service_name.*Up"; then
        echo -e "${GREEN}✅ $service_name is running${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not running${NC}"
        return 1
    fi
}

# Function to check port connectivity
check_port() {
    local service_name=$1
    local port=$2
    
    if nc -z localhost $port 2>/dev/null; then
        echo -e "${GREEN}✅ $service_name port $port is accessible${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name port $port is not accessible${NC}"
        return 1
    fi
}

# Function to check API endpoint
check_api_endpoint() {
    local endpoint=$1
    local expected_status=$2
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8000$endpoint")
    
    if [ "$response" = "$expected_status" ]; then
        echo -e "${GREEN}✅ API endpoint $endpoint returns $response${NC}"
        return 0
    else
        echo -e "${RED}❌ API endpoint $endpoint returns $response (expected $expected_status)${NC}"
        return 1
    fi
}

# Check Docker Compose services
echo -e "\n📦 Checking Docker Services"
echo "============================"

services=("postgres" "redis" "zookeeper" "kafka" "backend" "video-processor")
all_services_up=true

for service in "${services[@]}"; do
    if ! check_service "$service"; then
        all_services_up=false
    fi
done

# Check port connectivity
echo -e "\n🔌 Checking Port Connectivity"
echo "=============================="

ports=(
    "postgres:5432"
    "redis:6379"
    "kafka:9092"
    "backend:8000"
    "zookeeper:2181"
)

all_ports_accessible=true

for port_info in "${ports[@]}"; do
    IFS=':' read -r service port <<< "$port_info"
    if ! check_port "$service" "$port"; then
        all_ports_accessible=false
    fi
done

# Check API endpoints
echo -e "\n🌐 Checking API Endpoints"
echo "========================="

api_checks=(
    "/api/health:200"
    "/swagger/index.html:200"
    "/api/auth/login:405"  # Method not allowed for GET, but endpoint exists
)

all_apis_working=true

for api_check in "${api_checks[@]}"; do
    IFS=':' read -r endpoint expected_status <<< "$api_check"
    if ! check_api_endpoint "$endpoint" "$expected_status"; then
        all_apis_working=false
    fi
done

# Check Kafka topics
echo -e "\n📨 Checking Kafka Topics"
echo "========================"

kafka_topics=("video-processing" "video-processing-retry" "video-processing-dlq")
kafka_working=true

for topic in "${kafka_topics[@]}"; do
    if docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null | grep -q "$topic"; then
        echo -e "${GREEN}✅ Kafka topic '$topic' exists${NC}"
    else
        echo -e "${YELLOW}⚠️  Kafka topic '$topic' not found (will be created on first use)${NC}"
    fi
done

# Check database connectivity
echo -e "\n💾 Checking Database"
echo "===================="

if docker exec postgres pg_isready -U postgres >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
else
    echo -e "${RED}❌ PostgreSQL is not ready${NC}"
    all_services_up=false
fi

# Check Redis connectivity
echo -e "\n🔴 Checking Redis"
echo "================="

if docker exec redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
    echo -e "${GREEN}✅ Redis is responding${NC}"
else
    echo -e "${RED}❌ Redis is not responding${NC}"
    all_services_up=false
fi

# Check video processing assets
echo -e "\n🎬 Checking Video Assets"
echo "========================"

assets=("backend/assets/logo.png" "backend/assets/intro.mp4" "backend/assets/outro.mp4")
assets_present=true

for asset in "${assets[@]}"; do
    if [ -f "$asset" ]; then
        echo -e "${GREEN}✅ Asset $asset exists${NC}"
    else
        echo -e "${YELLOW}⚠️  Asset $asset not found${NC}"
        assets_present=false
    fi
done

# Final summary
echo -e "\n📋 Deployment Summary"
echo "====================="

if $all_services_up && $all_ports_accessible && $all_apis_working; then
    echo -e "${GREEN}🎉 All systems operational! Deployment is successful.${NC}"
    echo -e "\n📝 Next Steps:"
    echo "1. Upload videos via: POST http://localhost:8000/api/videos/upload"
    echo "2. Monitor workers: docker-compose logs -f video-processor"
    echo "3. View API docs: http://localhost:8000/swagger/index.html"
    exit 0
else
    echo -e "${RED}⚠️  Some issues detected. Please check the logs above.${NC}"
    echo -e "\n🔧 Troubleshooting:"
    echo "1. Check logs: docker-compose logs [service-name]"
    echo "2. Restart services: docker-compose restart"
    echo "3. Rebuild if needed: docker-compose up --build"
    exit 1
fi
