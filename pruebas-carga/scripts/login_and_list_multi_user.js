import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

export let options = {
  vus: __ENV.VUS || 10,
  duration: __ENV.DURATION || '2m',
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate should be less than 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://107.23.232.213:8000';

// Create multiple test users for concurrent load testing
export function setup() {
  const numUsers = parseInt(__ENV.VUS) || 10;
  const users = [];
  
  console.log(`Creating ${numUsers} test users...`);
  
  for (let i = 0; i < numUsers; i++) {
    const timestamp = Date.now();
    const testUser = {
      first_name: `LoadTest${i}`,
      last_name: 'User',
      email: `loadtest_${timestamp}_${i}@example.com`,
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
      console.error(`✗ Failed to create user ${i + 1}: ${signupRes.body}`);
    }
    
    sleep(0.1); // Small delay to avoid overwhelming the server during setup
  }
  
  console.log(`Setup complete: ${users.length} users created`);
  return { users };
}

export default function (data) {
  // Each VU uses a different user based on its ID
  const userIndex = __VU % data.users.length;
  const user = data.users[userIndex];
  
  // Login
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
    sleep(1);
    return;
  }

  let token = loginRes.json('access_token');

  // List user's videos
  let videosRes = http.get(`${BASE_URL}/api/videos`, {
    headers: { 
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'ListVideos' },
  });

  check(videosRes, {
    'list videos success': (r) => r.status === 200,
    'videos is array': (r) => Array.isArray(r.json()),
  });

  // Get public videos
  let publicVideosRes = http.get(`${BASE_URL}/api/public/videos`, {
    tags: { name: 'ListPublicVideos' },
  });

  check(publicVideosRes, {
    'public videos success': (r) => r.status === 200,
    'public videos is array': (r) => Array.isArray(r.json()),
  });

  // Get rankings
  let rankingsRes = http.get(`${BASE_URL}/api/public/rankings`, {
    tags: { name: 'GetRankings' },
  });

  check(rankingsRes, {
    'rankings success': (r) => r.status === 200,
    'rankings is array': (r) => Array.isArray(r.json()),
  });

  sleep(1);
}

// Teardown function runs once after all VUs complete
export function teardown(data) {
  console.log(`\nLoad test completed with ${data.users.length} users`);
}
