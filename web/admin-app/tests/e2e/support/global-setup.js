import dns from 'node:dns/promises';
import http from 'node:http';
import https from 'node:https';

const DEFAULT_ORIGIN = 'http://127.0.0.1:8090';
const READY_PATH = '/api/v1/health';
const MAX_ATTEMPTS = Number(process.env.DEVHUB_E2E_READY_ATTEMPTS || 40);
const DELAY_MS = Number(process.env.DEVHUB_E2E_READY_DELAY_MS || 1000);

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function requestHealth(url) {
  const client = url.protocol === 'https:' ? https : http;
  return new Promise((resolve, reject) => {
    const req = client.get(
      url,
      {
        timeout: 2000,
        headers: {
          Accept: 'application/json',
          'User-Agent': 'DevHub-Admin-E2E-Ready/1.0',
        },
      },
      (res) => {
        res.resume();
        res.on('end', () => {
          if (res.statusCode && res.statusCode >= 200 && res.statusCode < 500) {
            resolve(res.statusCode);
            return;
          }
          reject(new Error(`HTTP ${res.statusCode || 'unknown'}`));
        });
      },
    );
    req.on('timeout', () => {
      req.destroy(new Error('health probe timeout'));
    });
    req.on('error', reject);
  });
}

async function probe(origin) {
  const url = new URL(READY_PATH, origin);
  await dns.lookup(url.hostname);
  return requestHealth(url);
}

export default async function globalSetup() {
  const origin = process.env.DEVHUB_E2E_ORIGIN || DEFAULT_ORIGIN;
  let lastError = null;

  console.log(`[admin-e2e] baseURL=${origin}`);
  console.log(`[admin-e2e] waiting for ${new URL(READY_PATH, origin).toString()}`);

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt += 1) {
    try {
      const status = await probe(origin);
      console.log(`[admin-e2e] DevHub ready: HTTP ${status}`);
      return;
    } catch (error) {
      lastError = error;
      const message = error?.code ? `${error.code}: ${error.message}` : error?.message || String(error);
      if (attempt === 1 || attempt === MAX_ATTEMPTS || attempt % 5 === 0) {
        console.log(`[admin-e2e] DevHub not ready (${attempt}/${MAX_ATTEMPTS}): ${message}`);
      }
      await sleep(DELAY_MS);
    }
  }

  throw new Error(`DevHub E2E origin is not ready: ${origin}; last_error=${lastError?.message || lastError}`);
}
