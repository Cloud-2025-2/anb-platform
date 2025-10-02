# Load Testing Scripts for ANB Platform

This directory contains k6 load testing scripts for the ANB Rising Stars platform.

## Prerequisites

Install k6:
```bash
# Windows (using Chocolatey)
choco install k6

# macOS (using Homebrew)
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

## Available Scripts

### 1. `login_and_list.js` - Basic Authentication Load Test

Tests user login and video listing with a single test user.

**Features:**
- Creates one test user in setup
- Tests login endpoint
- Tests listing user's videos
- Validates JWT token structure

**Usage:**
```bash
# Basic run
k6 run scripts/login_and_list.js

# Custom configuration
k6 run --vus 10 --duration 30s scripts/login_and_list.js

# With environment variables
k6 run -e BASE_URL=http://107.23.232.213:8000 -e VUS=5 -e DURATION=1m scripts/login_and_list.js
```

### 2. `login_and_list_multi_user.js` - Multi-User Load Test

Creates multiple test users for more realistic concurrent load testing.

**Features:**
- Creates multiple test users (one per VU)
- Each VU uses a different user
- Tests multiple endpoints: login, list videos, public videos, rankings
- Includes performance thresholds
- Better simulation of real-world usage

**Usage:**
```bash
# Basic run (10 users, 2 minutes)
k6 run scripts/login_and_list_multi_user.js

# Custom load (50 concurrent users, 5 minutes)
k6 run -e VUS=50 -e DURATION=5m scripts/login_and_list_multi_user.js

# Heavy load test
k6 run -e VUS=100 -e DURATION=10m scripts/login_and_list_multi_user.js
```

### 3. `upload_and_poll.js` - Video Upload & Processing Test

Tests video upload and monitors processing status until completion.

**Features:**
- Creates test user for video uploads
- Uploads video file to server
- Polls video status until "processed"
- Measures processing time
- Extended timeout support for large files

**Usage:**
```bash
# Basic run (requires test video file)
k6 run -e FILE_PATH=./test-video.mp4 scripts/upload_and_poll.js

# Custom configuration
k6 run -e VUS=3 -e DURATION=5m -e FILE_PATH=./video.mp4 scripts/upload_and_poll.js

# With custom polling settings
k6 run -e MAX_POLL_ATTEMPTS=30 -e POLL_INTERVAL=3 scripts/upload_and_poll.js
```

**Environment Variables:**
- `BASE_URL`: API base URL (default: `http://107.23.232.213:8000`)
- `FILE_PATH`: Path to test video file (default: `./test-video.mp4`)
- `VUS`: Number of virtual users (default: `3`)
- `DURATION`: Test duration (default: `5m`)
- `MAX_POLL_ATTEMPTS`: Maximum polling attempts (default: `20`)
- `POLL_INTERVAL`: Seconds between status checks (default: `5`)

### 4. `batch_processing.js` - Batch Video Processing Test

Tests concurrent batch upload and processing of multiple videos per user.

**Features:**
- Creates multiple test users
- Each user uploads multiple videos in batch
- Monitors all videos until processed
- Tracks individual video processing times
- Custom metrics for video processing performance
- Performance thresholds (95% < 3 minutes)

**Usage:**
```bash
# Basic run (2 users, 3 videos each)
k6 run -e FILE_PATH=./test-video.mp4 scripts/batch_processing.js

# Heavy batch load (5 users, 5 videos each)
k6 run -e VUS=5 -e BATCH_SIZE=5 -e FILE_PATH=./video.mp4 scripts/batch_processing.js

# Extended test with custom polling
k6 run -e VUS=3 -e BATCH_SIZE=4 -e MAX_POLL_ATTEMPTS=100 -e POLL_INTERVAL=5 scripts/batch_processing.js
```

**Environment Variables:**
- `BASE_URL`: API base URL (default: `http://107.23.232.213:8000`)
- `FILE_PATH`: Path to test video file (default: `./test-video.mp4`)
- `VUS`: Number of virtual users (default: `2`)
- `DURATION`: Test duration (default: `10m`)
- `BATCH_SIZE`: Videos to upload per iteration (default: `3`)
- `MAX_POLL_ATTEMPTS`: Maximum polling attempts (default: `60`)
- `POLL_INTERVAL`: Seconds between status checks (default: `10`)

## Environment Variables Summary

| Variable | Default | Description | Used In |
|----------|---------|-------------|---------|
| `BASE_URL` | `http://107.23.232.213:8000` | API base URL | All scripts |
| `VUS` | `5`/`10`/`3`/`2` | Number of virtual users | All scripts |
| `DURATION` | `2m`/`5m`/`10m` | Test duration | All scripts |
| `FILE_PATH` | `./test-video.mp4` | Path to test video file | Upload scripts |
| `MAX_POLL_ATTEMPTS` | `20`/`60` | Maximum polling attempts | Upload scripts |
| `POLL_INTERVAL` | `5`/`10` | Seconds between polls | Upload scripts |
| `BATCH_SIZE` | `3` | Videos per batch | Batch script |

## API Endpoints Tested

### Authentication Endpoints
- `POST /api/auth/signup` - User registration
  - Body: `{ first_name, last_name, email, password1, password2, city, country }`
  - Response: `201` with success message

- `POST /api/auth/login` - User login
  - Body: `{ email, password }`
  - Response: `200` with `{ access_token, token_type, expires_in }`

### Protected Endpoints (require JWT)
- `POST /api/videos/upload` - Upload video file
  - Header: `Authorization: Bearer <token>`
  - Body: `multipart/form-data` with fields:
    - `video_file`: Video file (MP4, max 100MB)
    - `title`: Video title
  - Response: `201` with `{ message, task_id, video_id }`

- `GET /api/videos` - List user's videos
  - Header: `Authorization: Bearer <token>`
  - Response: `200` with array of videos

- `GET /api/videos/{video_id}` - Get video details and status
  - Header: `Authorization: Bearer <token>`
  - Response: `200` with video object including `status` field

### Public Endpoints
- `GET /api/public/videos` - List all public videos
- `GET /api/public/rankings` - Get player rankings

## Request Body Structures

### Signup Request
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "password1": "SecurePassword123!",
  "password2": "SecurePassword123!",
  "city": "Bogotá",
  "country": "Colombia"
}
```

### Login Request
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePassword123!"
}
```

### Login Response
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 1440
}
```

## Performance Thresholds

The multi-user script includes the following thresholds:
- **95th percentile response time**: < 500ms
- **Error rate**: < 1%

## Interpreting Results

### Key Metrics

- **http_req_duration**: Response time of HTTP requests
  - `avg`: Average response time
  - `p(95)`: 95th percentile (95% of requests are faster than this)
  - `max`: Maximum response time

- **http_req_failed**: Percentage of failed requests
  - Should be close to 0%

- **http_reqs**: Total number of requests made
  - Higher is better for throughput testing

- **iterations**: Number of complete test iterations
  - Each iteration = full test scenario (login + list videos)

### Example Output
```
✓ login success
✓ access_token received
✓ list videos success

checks.........................: 100.00% ✓ 1500      ✗ 0
data_received..................: 2.1 MB  35 kB/s
data_sent......................: 470 kB  7.8 kB/s
http_req_duration..............: avg=123ms min=45ms med=98ms max=456ms p(95)=287ms
http_req_failed................: 0.00%   ✓ 0        ✗ 500
http_reqs......................: 500     8.33/s
iterations.....................: 100     1.67/s
vus............................: 10      min=10     max=10
```

## Troubleshooting

### Issue: "Failed to create test user"

**Cause**: Server might be down or unreachable

**Solution**: 
```bash
# Test server connectivity
curl http://107.23.232.213:8000/api/health

# Check if BASE_URL is correct
k6 run -e BASE_URL=http://YOUR_SERVER:8000 scripts/login_and_list.js
```

### Issue: "Login failed: 401"

**Cause**: User creation succeeded but password doesn't match

**Solution**: Check that `password1` and `password` are consistent in the script

### Issue: High error rates

**Cause**: Server overload or database connection issues

**Solution**: 
- Reduce VUS count: `-e VUS=5`
- Increase sleep duration between requests
- Check server logs: `sudo docker logs anb-backend`

### Issue: Timeout errors

**Cause**: Server response time too slow

**Solution**:
```bash
# Increase timeout
k6 run --http-debug scripts/login_and_list.js

# Check server resources
docker stats
```

## Best Practices

1. **Start small**: Begin with 5-10 VUs and gradually increase
2. **Monitor server**: Watch CPU, memory, and database connections
3. **Realistic scenarios**: Use multi-user script for production testing
4. **Clean data**: Consider cleaning up test users after load tests
5. **Multiple iterations**: Run tests multiple times for consistency

## Cleanup Test Users

After load testing, you may want to remove test users:

```sql
-- Connect to PostgreSQL
docker exec -it anb-postgres psql -U postgres -d anb_platform

-- Delete test users
DELETE FROM users WHERE email LIKE 'loadtest_%@example.com';

-- Verify
SELECT COUNT(*) FROM users WHERE email LIKE 'loadtest_%@example.com';
```

## Advanced Usage

### Custom Scenarios

Create custom load profiles:

```javascript
export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
};
```

### Output to InfluxDB

```bash
k6 run --out influxdb=http://localhost:8086/k6 scripts/login_and_list_multi_user.js
```

### Generate HTML Report

```bash
k6 run --out json=results.json scripts/login_and_list_multi_user.js
k6-reporter results.json
```

## Support

For issues or questions:
- Check backend logs: `docker logs anb-backend`
- Check k6 documentation: https://k6.io/docs/
- Review API documentation: http://107.23.232.213:8000/swagger/index.html
