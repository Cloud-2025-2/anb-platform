import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: __ENV.VUS || 5,
  duration: __ENV.DURATION || '2m',
};

export default function () {
  // Login
  let loginRes = http.post(`${__ENV.BASE_URL}/api/auth/login`, JSON.stringify({
    email: __ENV.USER_EMAIL,
    password: __ENV.USER_PASS,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login success': (r) => r.status === 200,
    'token received': (r) => r.json('token') !== '',
  });

  let token = loginRes.json('token');

  // List videos
  let videosRes = http.get(`${__ENV.BASE_URL}/api/videos`, {
    headers: { Authorization: `Bearer ${token}` },
  });

  check(videosRes, {
    'list videos 200': (r) => r.status === 200,
  });

  sleep(1);
}
