// k6 smoke test for FireLine API
// Run: k6 run scripts/load/k6_smoke.js
// Requires: FireLine server running on localhost:8080

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const healthLatency = new Trend('health_latency');
const signupLatency = new Trend('signup_latency');
const loginLatency = new Trend('login_latency');

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // ramp up to 10 VUs
    { duration: '1m', target: 10 },    // stay at 10
    { duration: '30s', target: 50 },   // ramp up to 50
    { duration: '1m', target: 50 },    // stay at 50
    { duration: '30s', target: 0 },    // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],   // 95th percentile < 500ms
    errors: ['rate<0.05'],              // error rate < 5%
    health_latency: ['p(99)<100'],      // health check < 100ms p99
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Health check
  const healthRes = http.get(`${BASE_URL}/health/live`);
  healthLatency.add(healthRes.timings.duration);
  check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health body ok': (r) => r.json().status === 'ok',
  }) || errorRate.add(1);

  // Readiness check
  const readyRes = http.get(`${BASE_URL}/health/ready`);
  check(readyRes, {
    'ready status 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  // Auth flow: signup + login
  const uniqueId = `${__VU}-${__ITER}-${Date.now()}`;
  const signupPayload = JSON.stringify({
    org_name: `Load Test Org ${uniqueId}`,
    org_slug: `load-test-${uniqueId}`,
    email: `load-${uniqueId}@test.fireline.io`,
    password: 'LoadTest123!@#Strong',
    display_name: `Load User ${uniqueId}`,
  });

  const signupRes = http.post(`${BASE_URL}/api/v1/auth/signup`, signupPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  signupLatency.add(signupRes.timings.duration);

  const signupOk = check(signupRes, {
    'signup status 201': (r) => r.status === 201,
    'signup has tokens': (r) => {
      const body = r.json();
      return body.access_token && body.refresh_token;
    },
  });
  if (!signupOk) {
    errorRate.add(1);
    return;
  }

  // Login with the created user
  const loginPayload = JSON.stringify({
    email: `load-${uniqueId}@test.fireline.io`,
    password: 'LoadTest123!@#Strong',
  });

  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  loginLatency.add(loginRes.timings.duration);

  check(loginRes, {
    'login status 200': (r) => r.status === 200,
    'login has access_token': (r) => r.json().access_token !== undefined,
  }) || errorRate.add(1);

  // Token refresh
  if (signupRes.status === 201) {
    const refreshPayload = JSON.stringify({
      refresh_token: signupRes.json().refresh_token,
    });
    const refreshRes = http.post(`${BASE_URL}/api/v1/auth/refresh`, refreshPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    check(refreshRes, {
      'refresh status 200': (r) => r.status === 200,
    }) || errorRate.add(1);
  }

  sleep(1);
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: '  ', enableColors: true }),
  };
}

function textSummary(data) {
  const metrics = data.metrics;
  return `
=== FireLine Load Test Summary ===
  HTTP Request Duration (p95): ${metrics.http_req_duration?.values?.['p(95)']?.toFixed(2) || 'N/A'}ms
  Health Check Latency (p99):  ${metrics.health_latency?.values?.['p(99)']?.toFixed(2) || 'N/A'}ms
  Error Rate:                  ${((metrics.errors?.values?.rate || 0) * 100).toFixed(2)}%
  Total Requests:              ${metrics.http_reqs?.values?.count || 0}
  VUs (max):                   ${metrics.vus?.values?.max || 0}
`;
}
