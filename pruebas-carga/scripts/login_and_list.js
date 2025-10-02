import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: __ENV.VUS || 5,
  duration: __ENV.DURATION || '2m',
};

const BASE_URL = __ENV.BASE_URL || 'http://107.23.232.213:8000';

// Setup function runs once before all VUs start
export function setup() {
  const testUser = {
    first_name: 'LoadTest',
    last_name: 'User',
    email: `loadtest_${Date.now()}@example.com`,
    password1: 'SecurePassword123!',
    password2: 'SecurePassword123!',
    city: 'Bogotá',
    country: 'Colombia',
  };

  // Create test user
  let signupRes = http.post(`${BASE_URL}/api/auth/signup`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(signupRes, {
    'signup successful': (r) => r.status === 201,
  });

  if (signupRes.status !== 201) {
    console.error('Signup failed:', signupRes.body);
    throw new Error('Failed to create test user');
  }

  console.log(`Test user created: ${testUser.email}`);
  
  return {
    email: testUser.email,
    password: testUser.password1,
  };
}

export default function (data) {
  // Login with the created user
  let loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    email: data.email,
    password: data.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login success': (r) => r.status === 200,
    'access_token received': (r) => r.json('access_token') !== undefined && r.json('access_token') !== '',
    'token_type is bearer': (r) => r.json('token_type') === 'Bearer',
  });

  if (loginRes.status !== 200) {
    console.error('Login failed:', loginRes.body);
    return;
  }

  let token = loginRes.json('access_token');

  // List user's videos
  let videosRes = http.get(`${BASE_URL}/api/videos`, {
    headers: { 
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  check(videosRes, {
    'list videos success': (r) => r.status === 200,
    'videos is array': (r) => Array.isArray(r.json()),
  });

  sleep(1);
}
