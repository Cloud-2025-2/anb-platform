# Entrega 1: Implementación de una API REST escalable con orquestación de tareas asíncronas para el procesamiento de archivos

## Team

* Ivan Avila - i.avilag@gmail.com
* Ana M. Sánchez - am.sanchezm1@uniandes.edu.co
* David Tobón Molina - d.tobonm2@uniandes.edu.co

## Architecture

- **Backend API**: Go with Gin framework
- **Database**: PostgreSQL with GORM
- **Message Queue**: Apache Kafka with Zookeeper
- **Caching**: Redis for rankings cache
- **Video Processing**: FFmpeg with custom workers
- **Storage**: Local file system
- **Authentication**: JWT tokens

## Quickstart

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (version 20.10 or higher)
- [Docker Compose](https://docs.docker.com/compose/install/) (version 2.0 or higher)

### Running the Project

1. **Clone the repository**
   ```bash
   git clone https://github.com/Cloud-2025-2/anb-platform.git
   cd anb-platform
   ```

   Or download the source code in a _.zip_ file.

2. **Start all services**
   ```bash
   docker-compose up --build -d
   ```

   This starts:
   - PostgreSQL database
   - Redis cache
   - Zookeeper
   - Kafka broker
   - Backend API (port 8000): `http://localhost:8000`
   - Frontend (port 3000): `http://localhost:3000`
   - Video processing workers (2 replicas by default)

   ```bash
   # Scale video processors
   docker-compose up -d --scale video-processor=5
   ```

3. **Check service status**
   ```bash
   docker-compose ps
   ```

4. **View logs (optional)**
   ```bash
   # View all services logs
   docker-compose logs -f
   
   # View specific service logs
   docker-compose logs -f api
   docker-compose logs -f db
   docker-compose logs -f frontend
   ```

### Accessing the Application

- **Frontend**: [TODO](http://localhost:3000/)
- **Check API Health**: http://localhost:8000/health

### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (This will delete all data)
docker-compose down -v
```


### health-check

```
http://localhost:8000/api/health
```

## API Endpoints

### Authentication
- `POST /api/auth/signup` - User registration
- `POST /api/auth/login` - User login, returns JWT token which expires in 1 hour

### Video Management (JWT Required)
- `POST /api/videos/upload` - Upload video for processing
- `GET /api/videos` - List user's videos
- `GET /api/videos/{id}` - Get video details
- `DELETE /api/videos/{id}` - Delete video (if not published)
- `PUT /api/videos/{id}/publish` - Publish video for voting
- `POST /api/public/videos/{id}/vote` - Vote for video

### Public Endpoints
- `GET /api/public/videos` - List published videos
- `GET /api/public/rankings` - Get player rankings

*Check the OpenAPI or Postman docs for details.*

### OpenAPI Documentation

```
http://localhost:8000/swagger/index.html
```

### API Tests with Postman

```
npm install
npm run test
```

To generate an *.html* test report:

```
npm run test:report
```

Note: `Vote for Video` test takes around 20 seconds to run while newman waits for the video to be processed and made public.

## Video Processing

### Video Status Lifecycle

A video goes through several states from upload to being publicly available for voting. This lifecycle is managed automatically by the system.

- `uploaded`: The initial status when a video is successfully uploaded by a user. The video is pending for processing.
- `processing`: The video has been picked up by a worker and is actively being processed.
- `processed`: The video has been successfully processed (trimmed, transcoded, watermarked, etc.) and is ready to be published by the user.
- `published`: The user has published the video, making it visible to the public for voting.
- `failed`: The video processing failed after multiple retry attempts. The task is moved to the Dead Letter Queue (DLQ) for manual inspection.


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

### Processing Steps

The video processing pipeline is a sequence of automated steps executed by our workers after a video is uploaded. Each step is designed to standardize the content for the platform.

1.  **Upload & Enqueue**: A user uploads a video via the API. The video is saved to a temporary location, it's status is updated to `uploaded`, and a processing task with the video's metadata is sent to the `video-processing` Kafka topic.

2.  **Dequeue & Process**: A worker from the consumer group picks up the task from the Kafka topic. The video status is updated to `processing`.

3.  **Transformations**: The worker executes an FFmpeg pipeline with the following transformations:
    *   **Trim**: The video is trimmed to a maximum duration of **30 seconds**.
    *   **Resolution & Aspect Ratio**: The video is resized to **720p** (1280x720) and set to a **16:9** aspect ratio. Black bars are added if necessary to maintain the original aspect ratio.
    *   **Audio Removal**: The audio track is removed from the video.
    *   **Watermark**: The ANB logo (`logo.png`) is added as a watermark to the bottom-right corner of the video.
    *   **Concatenation**: An **intro** (`intro.mp4`) and **outro** (`outro.mp4`) sequence are added to the beginning and end of the main video. These sequences are appended at the en dof the pipeline since they already have the required resolution and aspect ratio.

4.  **Store**: The final, processed video is saved to the designated persistent storage.

5.  **Update**: The video's record in the database is updated with the new status (`processed`) and the URL to the processed file. If the processing fails, the retry mechanism is initiated.


### Retry Mechanism

- **Max Retries**: 3 attempts
- **Backoff**: Exponential (1s, 2s, 4s)
- **DLQ**: Failed videos sent to Dead Letter Queue
- **Monitoring**: All processing events logged

### Performance Tuning

- **Kafka Partitions**: Increase for higher throughput
- **Consumer Groups**: Scale workers horizontally
- **Redis Cache**: Tune TTL based on traffic patterns
- **FFmpeg**: Optimize encoding settings for quality/speed balance

## Security

- JWT authentication for protected endpoints (expiry time of 1 hour)
- Input validation on all API endpoints
- File type validation for uploads

## Caching Strategy

To ensure high performance and reduce the load on the database, a caching strategy is implemented for the player rankings endpoint (`/api/public/rankings`), which is expected to receive high traffic.

-   **Technology**: Redis is used as the caching layer due to its speed and efficiency.

-   **What is Cached**: The results of the rankings query are cached. The cache key is dynamically generated based on the query parameters (`limit` and `city`), ensuring that different filtered views of the rankings are cached separately. For example, a request for the top 10 players in Bogotá will have a different cache key than a request for the top 20 players overall.

-   **Cache TTL (Time-to-Live)**: Each cache entry is set with a configurable Time-to-Live (TTL), 3 minutes by deafault. This ensures that the data remains fresh and is automatically removed from the cache after a certain period, preventing stale data from being served.

-   **Cache Invalidation**: The cache is proactively invalidated to ensure users see up-to-date rankings. The `InvalidateAll` function is called whenever a significant event occurs (e.g., a new vote is cast), which deletes all cached ranking data and forces the system to fetch fresh results from the database on the next request.

-   **Flow**:
    1.  A request is made to the rankings endpoint.
    2.  The system first checks Redis for a cached result using a key generated from the request parameters.
    3.  **Cache Hit**: If the data is found in the cache, it is returned directly to the user.
    4.  **Cache Miss**: If the data is not in the cache, the system queries the PostgreSQL database, stores the result in Redis with the defined TTL, and then returns the data to the user.


# SonarQube
https://sonarcloud.io/project/overview?id=Cloud-2025-2_anb-platform
