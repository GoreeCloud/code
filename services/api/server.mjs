import { createServer } from 'node:http';
import { open } from 'node:fs/promises';
import { constants } from 'node:fs';
import { createHash, timingSafeEqual } from 'node:crypto';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROUTE = /^\/api\/v1\/repositories\/([A-Za-z0-9][A-Za-z0-9_.-]{0,79})\/([A-Za-z0-9][A-Za-z0-9_.-]{0,99})$/;
const HEADERS = {
  'content-type': 'application/json; charset=utf-8',
  'cache-control': 'no-store',
  'x-content-type-options': 'nosniff',
  'content-security-policy': "default-src 'none'; frame-ancestors 'none'",
  'referrer-policy': 'no-referrer',
};

function reply(res, code, payload) {
  res.writeHead(code, HEADERS);
  res.end(JSON.stringify(payload));
}

export async function readProtectedSecret(path) {
  if (!path || typeof path !== 'string') throw Error('missing secret path');
  // Prevent symlink substitution and inspect the opened file, not a separate path lookup.
  const handle = await open(resolve(path), constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const info = await handle.stat();
    if (!info.isFile() || info.size < 16 || info.size > 4096 ||
        (info.mode & 0o077) !== 0 ||
        (typeof process.getuid === 'function' && info.uid !== process.getuid())) {
      throw Error('unsafe secret file');
    }
    const value = (await handle.readFile('utf8')).trim();
    if (value.length < 16 || /[\r\n]/.test(value)) throw Error('invalid secret');
    return value;
  } finally {
    await handle.close();
  }
}

function safeEqual(a, b) {
  if (typeof a !== 'string' || typeof b !== 'string') return false;
  const digest = value => createHash('sha256').update(value).digest();
  return timingSafeEqual(digest(a), digest(b));
}

export function forgejoOrigin(configured) {
  if (!configured) throw Error('Forgejo origin missing');
  const url = new URL(configured);
  if (url.protocol !== 'https:' || url.username || url.password ||
      url.pathname !== '/' || url.search || url.hash) {
    throw Error('Forgejo URL must be an HTTPS origin with no credentials or path');
  }
  return url.origin;
}

async function boundedJson(response) {
  if (!response.body) throw Error('upstream body missing');
  const reader = response.body.getReader();
  let bytes = 0;
  const parts = [];
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      bytes += value.byteLength;
      if (bytes > 65536) throw Error('provider response too large');
      parts.push(value);
    }
  } finally {
    await reader.cancel().catch(() => {});
  }
  return JSON.parse(new TextDecoder().decode(Buffer.concat(parts)));
}

export function createCodeHandler({
  env = process.env,
  readSecret = readProtectedSecret,
  fetchImpl = fetch,
} = {}) {
  return async function handle(req, res) {
    try {
      if (!req.url || req.url.length > 2048) return reply(res, 404, { error: 'not_found' });
      const url = new URL(req.url, 'http://localhost');
      if (req.method === 'GET' && url.pathname === '/healthz' && !url.search) {
        return reply(res, 200, { service: 'goreecloud-code-api', state: 'development' });
      }
      if (req.method !== 'GET') return reply(res, 405, { error: 'read_only' });
      const match = url.search ? null : url.pathname.match(ROUTE);
      if (!match || match[1] === '..' || match[2] === '..') return reply(res, 404, { error: 'not_found' });
      if (!env.GOREECLOUD_CODE_ACCESS_TOKEN_FILE ||
          !env.GOREECLOUD_CODE_FORGEJO_TOKEN_FILE ||
          !env.GOREECLOUD_CODE_FORGEJO_URL) {
        return reply(res, 503, { error: 'not_configured' });
      }
      const appToken = await readSecret(env.GOREECLOUD_CODE_ACCESS_TOKEN_FILE);
      const authorization = req.headers.authorization || '';
      if (!authorization.startsWith('Bearer ') || !safeEqual(authorization.slice(7), appToken)) {
        return reply(res, 401, { error: 'unauthorized' });
      }
      const origin = forgejoOrigin(env.GOREECLOUD_CODE_FORGEJO_URL);
      const providerToken = await readSecret(env.GOREECLOUD_CODE_FORGEJO_TOKEN_FILE);
      const urlPath = '/api/v1/repos/' + encodeURIComponent(match[1]) + '/' + encodeURIComponent(match[2]);
      let upstream;
      try {
        upstream = await fetchImpl(origin + urlPath, {
          method: 'GET',
          redirect: 'error',
          signal: AbortSignal.timeout(8000),
          headers: { authorization: 'token ' + providerToken, accept: 'application/json' },
        });
      } catch {
        return reply(res, 502, { error: 'provider_unavailable' });
      }
      if (!upstream.ok) return reply(res, 502, { error: 'provider_unavailable' });
      let data;
      try {
        data = await boundedJson(upstream);
      } catch {
        return reply(res, 502, { error: 'provider_invalid_response' });
      }
      if (typeof data?.full_name !== 'string') return reply(res, 502, { error: 'provider_invalid_response' });
      return reply(res, 200, {
        fullName: data.full_name,
        description: typeof data.description === 'string' ? data.description : '',
        private: data.private === true,
        archived: data.archived === true,
        defaultBranch: typeof data.default_branch === 'string' ? data.default_branch : null,
      });
    } catch {
      // Do not leak provider URLs, credentials, stack traces or private response bodies.
      return reply(res, 503, { error: 'service_unavailable' });
    }
  };
}

export function startServer() {
  const port = Number(process.env.GOREECLOUD_CODE_PORT || 8731);
  if (!Number.isInteger(port) || port < 1 || port > 65535) throw Error('invalid port');
  const server = createServer(createCodeHandler());
  server.listen(port, '127.0.0.1');
  return server;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  startServer();
}
