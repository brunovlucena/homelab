import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up to 10 users
    { duration: '5m', target: 10 }, // Stay at 10 users
    { duration: '2m', target: 20 }, // Ramp up to 20 users
    { duration: '5m', target: 20 }, // Stay at 20 users
    { duration: '2m', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': 'k6-load-test',
    },
  };

  // Test health endpoint
  const healthCheck = http.get(`${BASE_URL}/health`, params);
  check(healthCheck, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
  });

  // Test projects endpoint
  const projectsResponse = http.get(`${BASE_URL}/api/v1/projects`, params);
  check(projectsResponse, {
    'projects status is 200': (r) => r.status === 200,
    'projects response time < 300ms': (r) => r.timings.duration < 300,
    'projects returns array': (r) => Array.isArray(JSON.parse(r.body)),
  });

  // Test metrics endpoint
  const metricsResponse = http.get(`${BASE_URL}/metrics`, params);
  check(metricsResponse, {
    'metrics status is 200': (r) => r.status === 200,
    'metrics response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Simulate project view tracking
  const trackData = {
    project_id: 1,
    ip: '127.0.0.1',
    user_agent: 'k6-load-test',
    referrer: 'https://lucena.cloud',
  };

  const trackResponse = http.post(
    `${BASE_URL}/api/v1/analytics/track`,
    JSON.stringify(trackData),
    params
  );
  check(trackResponse, {
    'track status is 200': (r) => r.status === 200,
    'track response time < 400ms': (r) => r.timings.duration < 400,
  });

  // Add some think time between requests
  sleep(1);
}

export function handleSummary(data) {
  return {
    'k6-results.json': JSON.stringify(data),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
} 