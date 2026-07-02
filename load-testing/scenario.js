import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { URLSearchParams } from 'https://jslib.k6.io/url/1.0.0/index.js';

export const _options = {
  scenarios: {
    // 1️⃣ Baseline latency & throughput
    baseline_vus_10: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',
      exec: 'baseline',
    },

    // 2️⃣ burst_traffict 
    burst_traffict: {
      executor: 'ramping-arrival-rate',
      startTime: '1m30s',
      timeUnit: '1s',
      stages: [
        { target: 20, duration: '30s' },
        { target: 50, duration: '10s' }, // burst
        { target: 50, duration: '30s' },
        { target: 20, duration: '20s' },
      ],
      preAllocatedVUs: 30,
      maxVUs: 80,
      exec: 'burstTraffic',
    },
  },

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: [
      'p(50)<150',
      'p(95)<200',
      'p(99)<300',
    ],
  },
};

export function headers(TOKEN, IDENTITY) {
  return {
    headers: {
      Authorization: `Bearer ${TOKEN}`,
      'Content-Type': 'application/json',
      'identity-key': IDENTITY,
    },
  };
}

export function params(URL,PARAMS) {
  // buang param kosong
  const cleanParams = Object.fromEntries(
    Object.entries(PARAMS).filter(([_, v]) => v !== "")
  );

  const qs = new URLSearchParams(cleanParams).toString();
  const finalURL = `${URL}?${qs}`;
  return finalURL;
}

// ================= SCENARIOS =================
export function _get_baseline(URL, TOKEN, IDENTITY, PARAMS) {
  const res = http.get(params(URL, PARAMS), headers(TOKEN, IDENTITY));
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(0.2); // simulasi user think time
}

export function _get_burstTraffic(URL, TOKEN, IDENTITY, PARAMS) {
  const res = http.get(params(URL, PARAMS), headers(TOKEN, IDENTITY));
  check(res, { 'status is 200': (r) => r.status === 200 });
}

export function _post_baseline(URL, TOKEN, IDENTITY, PAYLOAD) {
  const body = JSON.stringify(PAYLOAD);
  const res = http.post(URL, body, headers(TOKEN, IDENTITY));
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(0.2); // simulasi user think time
}

export function _post_burstTraffic(URL, TOKEN, IDENTITY, PAYLOAD) {
  const body = JSON.stringify(PAYLOAD);
  const res = http.post(URL, body, headers(TOKEN, IDENTITY));
  check(res, { 'status is 200': (r) => r.status === 200 });
}
