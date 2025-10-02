import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: parseInt(__ENV.VUS) || 3,
  duration: __ENV.DURATION || '5m',
};

const BASE_URL = __ENV.BASE_URL || 'http://107.23.232.213:8000';
const FILE_PATH = __ENV.FILE_PATH || './test-video.mp4';
const MAX_POLL_ATTEMPTS = parseInt(__ENV.MAX_POLL_ATTEMPTS) || 20;
const POLL_INTERVAL = parseInt(__ENV.POLL_INTERVAL) || 5;

// Setup function - creates a test user for video upload
export function setup() {
  const timestamp = Date.now();
  const testUser = {
    first_name: 'VideoTest',
    last_name: 'User',
    email: `videotest_${timestamp}@example.com`,
    password1: 'SecurePassword123!',
    password2: 'SecurePassword123!',
    city: 'Bogotá',
    country: 'Colombia',
  };

  // Create test user
  let signupRes = http.post(`${BASE_URL}/api/auth/signup`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (signupRes.status !== 201) {
    console.error('Failed to create test user:', signupRes.body);
    throw new Error('Setup failed: Could not create test user');
  }

  console.log(`✓ Test user created: ${testUser.email}`);

  return {
    email: testUser.email,
    password: testUser.password1,
  };
}

export default function (data) {
  // 1. Login
  let loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    email: data.email,
    password: data.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'Login' },
  });

  check(loginRes, {
    'login success': (r) => r.status === 200,
    'access_token received': (r) => r.json('access_token') !== undefined,
  });

  if (loginRes.status !== 200) {
    console.error('Login failed:', loginRes.body);
    return;
  }

  let token = loginRes.json('access_token');

  // 2. Upload video
  const formData = {
    video_file: http.file(open(FILE_PATH, 'b'), 'test-video.mp4', 'video/mp4'),
    title: `Load Test Video ${Date.now()}`,
  };

  let uploadRes = http.post(`${BASE_URL}/api/videos/upload`, formData, {
    headers: { 
      'Authorization': `Bearer ${token}`,
    },
    tags: { name: 'VideoUpload' },
    timeout: '120s', // Extended timeout for large files
  });

  check(uploadRes, {
    'upload accepted': (r) => r.status === 201,
    'video_id received': (r) => r.json('video_id') !== undefined,
  });

  if (uploadRes.status !== 201) {
    console.error('Upload failed:', uploadRes.body);
    return;
  }

  let videoId = uploadRes.json('video_id');
  let taskId = uploadRes.json('task_id');
  console.log(`Video uploaded: ID=${videoId}, TaskID=${taskId}`);

  // 3. Poll video status until "processed"
  let status = 'uploaded';
  let attempts = 0;

  while (status !== 'processed' && attempts < MAX_POLL_ATTEMPTS) {
    sleep(POLL_INTERVAL);
    
    let pollRes = http.get(`${BASE_URL}/api/videos/${videoId}`, {
      headers: { 
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      tags: { name: 'PollVideoStatus' },
    });

    check(pollRes, {
      'poll success': (r) => r.status === 200,
    });

    if (pollRes.status === 200) {
      status = pollRes.json('status');
      
      if (status === 'processed') {
        console.log(`✓ Video ${videoId} processed successfully after ${attempts + 1} attempts`);
        break;
      } else if (status === 'failed') {
        console.error(`✗ Video ${videoId} processing failed`);
        break;
      }
    }

    attempts++;
  }

  if (attempts >= MAX_POLL_ATTEMPTS) {
    console.warn(`⚠ Video ${videoId} did not complete processing within ${MAX_POLL_ATTEMPTS} attempts`);
  }

  check(status, {
    'video processed': (s) => s === 'processed',
  });
}

export function teardown(data) {
  console.log(`\nVideo upload test completed for user: ${data.email}`);
}
