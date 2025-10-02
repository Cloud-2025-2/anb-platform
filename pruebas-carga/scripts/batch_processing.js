import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';

// Custom metrics
const videoProcessingTime = new Trend('video_processing_time');

export let options = {
  vus: parseInt(__ENV.VUS) || 2,
  duration: __ENV.DURATION || '10m',
  thresholds: {
    'video_processing_time': ['p(95)<180000'], // 95% should process within 3 minutes
    'http_req_failed': ['rate<0.05'], // Less than 5% error rate
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://107.23.232.213:8000';
const FILE_PATH = __ENV.FILE_PATH || './test-video.mp4';
const MAX_POLL_ATTEMPTS = parseInt(__ENV.MAX_POLL_ATTEMPTS) || 60;
const POLL_INTERVAL = parseInt(__ENV.POLL_INTERVAL) || 10;
const BATCH_SIZE = parseInt(__ENV.BATCH_SIZE) || 3;

// Setup function - creates multiple test users for batch testing
export function setup() {
  const numUsers = parseInt(__ENV.VUS) || 2;
  const users = [];
  
  console.log(`Creating ${numUsers} test users for batch processing...`);
  
  for (let i = 0; i < numUsers; i++) {
    const timestamp = Date.now();
    const testUser = {
      first_name: `BatchTest${i}`,
      last_name: 'User',
      email: `batchtest_${timestamp}_${i}@example.com`,
      password1: 'SecurePassword123!',
      password2: 'SecurePassword123!',
      city: 'Bogotá',
      country: 'Colombia',
    };

    // Create test user
    let signupRes = http.post(`${BASE_URL}/api/auth/signup`, JSON.stringify(testUser), {
      headers: { 'Content-Type': 'application/json' },
    });

    if (signupRes.status === 201) {
      users.push({
        email: testUser.email,
        password: testUser.password1,
      });
      console.log(`✓ User ${i + 1}/${numUsers} created: ${testUser.email}`);
    } else {
      console.error(`✗ Failed to create user ${i + 1}:`, signupRes.body);
    }
    
    sleep(0.1);
  }
  
  console.log(`Setup complete: ${users.length} users created for batch testing`);
  return { users };
}

export default function (data) {
  // Each VU uses a different user
  const userIndex = __VU % data.users.length;
  const user = data.users[userIndex];

  // 1. Login
  let loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    email: user.email,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'Login' },
  });

  check(loginRes, {
    'login success': (r) => r.status === 200,
    'access_token received': (r) => r.json('access_token') !== undefined,
  });

  if (loginRes.status !== 200) {
    console.error(`Login failed for ${user.email}:`, loginRes.body);
    return;
  }

  let token = loginRes.json('access_token');

  // 2. Upload multiple videos in batch
  let uploadedVideos = [];
  
  for (let i = 0; i < BATCH_SIZE; i++) {
    const formData = {
      video_file: http.file(open(FILE_PATH, 'b'), `batch-video-${i}.mp4`, 'video/mp4'),
      title: `Batch Test Video ${__VU}-${__ITER}-${i} (${Date.now()})`,
    };

    let uploadRes = http.post(`${BASE_URL}/api/videos/upload`, formData, {
      headers: { 
        'Authorization': `Bearer ${token}`,
      },
      tags: { name: 'BatchVideoUpload' },
      timeout: '120s',
    });

    check(uploadRes, {
      'batch upload accepted': (r) => r.status === 201,
      'video_id received': (r) => r.json('video_id') !== undefined,
    });

    if (uploadRes.status === 201) {
      const videoId = uploadRes.json('video_id');
      const taskId = uploadRes.json('task_id');
      uploadedVideos.push({
        videoId: videoId,
        taskId: taskId,
        startTime: Date.now(),
      });
      console.log(`Uploaded video ${i + 1}/${BATCH_SIZE}: ID=${videoId}, TaskID=${taskId}`);
    } else {
      console.error(`Failed to upload video ${i + 1}:`, uploadRes.body);
    }
    
    sleep(1); // Small delay between uploads
  }

  // 3. Poll all uploaded videos until processed
  console.log(`Polling ${uploadedVideos.length} videos...`);
  
  let processedCount = 0;
  let attempts = 0;

  while (processedCount < uploadedVideos.length && attempts < MAX_POLL_ATTEMPTS) {
    sleep(POLL_INTERVAL);
    
    for (let i = 0; i < uploadedVideos.length; i++) {
      const video = uploadedVideos[i];
      
      if (video.status === 'processed' || video.status === 'failed') {
        continue; // Skip already completed videos
      }

      let pollRes = http.get(`${BASE_URL}/api/videos/${video.videoId}`, {
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        tags: { name: 'BatchPollVideoStatus' },
      });

      check(pollRes, {
        'batch poll success': (r) => r.status === 200,
      });

      if (pollRes.status === 200) {
        video.status = pollRes.json('status');
        
        if (video.status === 'processed') {
          const endTime = Date.now();
          const processingTime = endTime - video.startTime;
          videoProcessingTime.add(processingTime);
          
          console.log(`✓ Video ${video.videoId} processed in ${(processingTime / 1000).toFixed(2)}s`);
          processedCount++;
        } else if (video.status === 'failed') {
          console.error(`✗ Video ${video.videoId} processing failed`);
          processedCount++;
        }
      }
    }

    attempts++;
    
    if (processedCount === uploadedVideos.length) {
      console.log(`✓ All ${uploadedVideos.length} videos processed successfully`);
      break;
    }
  }

  if (attempts >= MAX_POLL_ATTEMPTS) {
    const remaining = uploadedVideos.length - processedCount;
    console.warn(`⚠ ${remaining} video(s) did not complete processing within ${MAX_POLL_ATTEMPTS} attempts`);
  }

  check(processedCount, {
    'all batch videos processed': (count) => count === uploadedVideos.length,
  });

  sleep(5); // Delay before next iteration
}

export function teardown(data) {
  console.log(`\nBatch processing test completed`);
  console.log(`Total users: ${data.users.length}`);
  console.log(`Batch size per iteration: ${BATCH_SIZE}`);
}
