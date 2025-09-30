import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: __ENV.VUS || 3,
  duration: __ENV.DURATION || '5m',
};

export default function () {
  // 1. Login
  let loginRes = http.post(`${__ENV.BASE_URL}/api/auth/login`, JSON.stringify({
    email: __ENV.USER_EMAIL,
    password: __ENV.USER_PASS,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login success': (r) => r.status === 200,
  });

  let token = loginRes.json('token');

  // 2. Subida de video
  const filePath = __ENV.FILE_PATH || '/data/video_50mb.mp4';
  const formData = {
    file: http.file(open(filePath, 'b'), 'video.mp4', 'video/mp4'),
  };

  let uploadRes = http.post(`${__ENV.BASE_URL}/api/videos/upload`, formData, {
    headers: { Authorization: `Bearer ${token}` },
  });

  check(uploadRes, {
    'upload accepted': (r) => r.status === 201 || r.status === 202,
  });

  let videoId = uploadRes.json('id');

  // 3. Poll de estado hasta "processed"
  let status = 'uploaded';
  let attempts = 0;

  while (status !== 'processed' && attempts < 20) {
    let pollRes = http.get(`${__ENV.BASE_URL}/api/videos/${videoId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });

    status = pollRes.json('status');
    check(pollRes, {
      'poll success': (r) => r.status === 200,
    });

    if (status === 'processed') {
      break;
    }

    sleep(5); // espera entre polls
    attempts++;
  }
}
