# ANB Rising Stars Showcase Platform

A scalable REST API platform for basketball talent discovery with asynchronous video processing using Apache Kafka.

## 🏗️ Architecture

- **Backend API**: Go with Gin framework
- **Database**: PostgreSQL with GORM
- **Message Queue**: Apache Kafka with Zookeeper
- **Caching**: Redis for rankings cache
- **Video Processing**: FFmpeg with custom workers
- **Storage**: Local file system (abstracted for cloud migration)
- **Authentication**: JWT tokens

## 🎥 Video Processing Pipeline

Videos undergo comprehensive processing:

1. **Duration**: Cut to maximum 30 seconds
2. **Resolution**: Standardized to 720p (1280x720)
3. **Aspect Ratio**: Adjusted to 16:9 with padding
4. **Audio**: Removed for consistency
5. **Watermark**: ANB logo overlay (bottom-right)
6. **Intro/Outro**: Added to processed videos
7. **Status Tracking**: uploaded → processing → processed → published

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### 1. Clone Repository

```bash
git clone https://github.com/Cloud-2025-2/anb-platform.git
cd anb-platform
```

### 2. Environment Setup

Create `.env` file in the backend directory:

```env
POSTGRES_HOST=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=anb_platform
POSTGRES_PORT=5432
REDIS_ADDR=redis:6379
KAFKA_BROKERS=kafka:29092
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRE_MINUTES=60
APP_PORT=8000
```

### 3. Start Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL database
- Redis cache
- Zookeeper
- Kafka broker
- Backend API (port 8000)
- Video processing workers (2 replicas)

### 4. Verify Setup

```bash
# Check API health
curl http://localhost:8000/api/health

# View API documentation
open http://localhost:8000/swagger/index.html
```

## 📋 API Endpoints

### Authentication
- `POST /api/auth/signup` - User registration
- `POST /api/auth/login` - User login

### Video Management (JWT Required)
- `POST /api/videos/upload` - Upload video for processing
- `GET /api/videos` - List user's videos
- `GET /api/videos/{id}` - Get video details
- `DELETE /api/videos/{id}` - Delete video (if not published)
- `PUT /api/videos/{id}/publish` - Publish video for voting

### Public Endpoints
- `GET /api/public/videos` - List published videos
- `POST /api/public/videos/{id}/vote` - Vote for video (JWT required)
- `GET /api/public/rankings` - Get player rankings (cached)

## 🎬 Video Processing Details

### Kafka Topics

- `video-processing` - Main processing queue
- `video-processing-retry` - Retry queue with exponential backoff
- `video-processing-dlq` - Dead Letter Queue for failed processing

### Processing Steps

1. **Upload**: Video uploaded via API, task sent to Kafka
2. **Queue**: Kafka consumers pick up processing tasks
3. **Process**: FFmpeg pipeline applies all transformations
4. **Store**: Processed video saved to storage
5. **Update**: Database updated with processed video info

### Retry Mechanism

- **Max Retries**: 3 attempts
- **Backoff**: Exponential (1s, 2s, 4s)
- **DLQ**: Failed videos sent to Dead Letter Queue
- **Monitoring**: All processing events logged

## 🏃‍♂️ Development

### Local Development

```bash
# Backend API
cd backend
go run cmd/api/main.go

# Video Workers
go run cmd/worker/main.go
```

### Testing

```bash
# Run tests
go test ./...

# Test video upload
curl -X POST http://localhost:8000/api/videos/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "video_file=@test.mp4" \
  -F "title=Test Video"
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | Database host | localhost |
| `KAFKA_BROKERS` | Kafka broker addresses | localhost:9092 |
| `REDIS_ADDR` | Redis address | localhost:6379 |
| `JWT_SECRET` | JWT signing secret | (required) |
| `APP_PORT` | API server port | 8000 |

### Video Processing Assets

Place required assets in `backend/assets/`:
- `logo.png` - ANB watermark
- `intro.mp4` - Intro video segment
- `outro.mp4` - Outro video segment

## 📊 Monitoring

### Kafka Topics Status

```bash
# List topics
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check consumer groups
docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list
```

### Logs

```bash
# API logs
docker-compose logs -f backend

# Worker logs
docker-compose logs -f video-processor

# Kafka logs
docker-compose logs -f kafka
```

## 🏗️ Production Deployment

### Scaling Workers

```bash
# Scale video processors
docker-compose up -d --scale video-processor=5
```

### Performance Tuning

- **Kafka Partitions**: Increase for higher throughput
- **Consumer Groups**: Scale workers horizontally
- **Redis Cache**: Tune TTL based on traffic patterns
- **FFmpeg**: Optimize encoding settings for quality/speed balance

## 🔒 Security

- JWT authentication for protected endpoints
- Input validation on all API endpoints
- File type validation for uploads
- Rate limiting (recommended for production)

## 📈 Caching Strategy

Rankings endpoint uses Redis caching:
- **TTL**: 3 minutes
- **Invalidation**: On vote events
- **Fallback**: Database query on cache miss

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## 📄 License

MIT License - see LICENSE file for details.

# Integrantes
* Ivan Avila - i.avilag@gmail.com
* Ana M. Sánchez - am.sanchezm1@uniandes.edu.co
* David Tobón Molina - d.tobonm2@uniandes.edu.co

# SonarQube
https://sonarcloud.io/project/overview?id=Cloud-2025-2_anb-platform
