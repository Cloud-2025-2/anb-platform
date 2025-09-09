# ANB Platform Makefile
# Provides convenient commands for development and deployment

.PHONY: help build up down logs test clean validate monitor performance

# Default target
help:
	@echo "🚀 ANB Platform - Available Commands"
	@echo "===================================="
	@echo "Development:"
	@echo "  make build     - Build all Docker images"
	@echo "  make up        - Start all services"
	@echo "  make down      - Stop all services"
	@echo "  make restart   - Restart all services"
	@echo "  make logs      - View logs from all services"
	@echo ""
	@echo "Testing & Monitoring:"
	@echo "  make validate  - Validate deployment"
	@echo "  make test      - Run Go tests"
	@echo "  make monitor   - Monitor Kafka topics"
	@echo "  make perf      - Run performance tests"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean     - Clean up containers and volumes"
	@echo "  make reset     - Full reset (clean + rebuild)"
	@echo "  make deps      - Update Go dependencies"

# Development commands
build:
	@echo "🔨 Building Docker images..."
	docker-compose build

up:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo "✅ Services started. API available at http://localhost:8000"
	@echo "📚 API Documentation: http://localhost:8000/swagger/index.html"

down:
	@echo "🛑 Stopping all services..."
	docker-compose down

restart: down up

logs:
	@echo "📋 Showing logs from all services..."
	docker-compose logs -f

# Service-specific logs
logs-api:
	docker-compose logs -f backend

logs-worker:
	docker-compose logs -f video-processor

logs-kafka:
	docker-compose logs -f kafka

logs-db:
	docker-compose logs -f postgres

# Testing and validation
validate:
	@echo "🔍 Validating deployment..."
	@chmod +x scripts/validate_deployment.sh
	@./scripts/validate_deployment.sh

test:
	@echo "🧪 Running Go tests..."
	cd backend && go test ./...

test-pipeline:
	@echo "🎬 Testing video processing pipeline..."
	cd backend && go run test_video_pipeline.go

monitor:
	@echo "📊 Starting Kafka monitoring..."
	@chmod +x scripts/monitor_kafka.sh
	@./scripts/monitor_kafka.sh

perf:
	@echo "⚡ Running performance tests..."
	cd scripts && go run performance_test.go

# Database operations
db-migrate:
	@echo "🗄️  Running database migrations..."
	docker-compose exec backend go run cmd/migrate/main.go

db-seed:
	@echo "🌱 Seeding database with test data..."
	docker-compose exec backend go run cmd/seed/main.go

db-reset:
	@echo "🔄 Resetting database..."
	docker-compose down postgres
	docker volume rm anb-platform_postgres_data || true
	docker-compose up -d postgres
	@sleep 5
	@make db-migrate

# Kafka operations
kafka-topics:
	@echo "📝 Listing Kafka topics..."
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-create-topics:
	@echo "📨 Creating Kafka topics..."
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing --partitions 3 --replication-factor 1 || true
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-retry --partitions 3 --replication-factor 1 || true
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --create --topic video-processing-dlq --partitions 1 --replication-factor 1 || true

kafka-reset:
	@echo "🔄 Resetting Kafka..."
	docker-compose down kafka zookeeper
	docker volume rm anb-platform_kafka_data anb-platform_zookeeper_data || true
	docker-compose up -d zookeeper kafka
	@sleep 10
	@make kafka-create-topics

# Maintenance
clean:
	@echo "🧹 Cleaning up containers and volumes..."
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

reset: clean build up
	@echo "🔄 Full reset complete!"
	@sleep 5
	@make validate

deps:
	@echo "📦 Updating Go dependencies..."
	cd backend && go mod tidy && go mod download

# Development helpers
dev-api:
	@echo "🔧 Starting API in development mode..."
	cd backend && go run cmd/api/main.go

dev-worker:
	@echo "👷 Starting worker in development mode..."
	cd backend && go run cmd/worker/main.go

# Production deployment
prod-build:
	@echo "🏭 Building for production..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

prod-up:
	@echo "🚀 Starting production deployment..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Health checks
health:
	@echo "🏥 Checking service health..."
	@curl -f http://localhost:8000/api/health || echo "❌ API health check failed"
	@docker exec postgres pg_isready -U postgres || echo "❌ Database health check failed"
	@docker exec redis redis-cli ping || echo "❌ Redis health check failed"
	@docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1 || echo "❌ Kafka health check failed"

# Scaling
scale-workers:
	@echo "⚡ Scaling video processors to 5 instances..."
	docker-compose up -d --scale video-processor=5

scale-down:
	@echo "📉 Scaling video processors to 2 instances..."
	docker-compose up -d --scale video-processor=2

# Backup and restore
backup-db:
	@echo "💾 Backing up database..."
	docker exec postgres pg_dump -U postgres anb_platform > backup_$(shell date +%Y%m%d_%H%M%S).sql

restore-db:
	@echo "🔄 Restoring database..."
	@read -p "Enter backup file path: " backup_file; \
	docker exec -i postgres psql -U postgres anb_platform < $$backup_file
