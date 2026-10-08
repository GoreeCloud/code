import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { once } from 'node:events';
import { createCodeHandler, forgejoOrigin } from './server.mjs';

const APPLICATION_SECRET = 'application-token-for-test-only';
const ENV = {
  GOREECLOUD_CODE_ACCESS_TOKEN_FILE: '/test/access',
  GOREECLOUD_CODE_FORGEJO_TOKEN_FILE: '/test/provider',
  GOREECLOUD_CODE_FORGEJO_URL: 'https://forgejo.example',
  GOREECLOUD_CODE_ALLOWED_REPOSITORIES_FILE: '/test/allowed.json',
};

async function withServer(handler, fn) {
  const server = createServer(handler);
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  try {
    await fn('http://127.0.0.1:' + server.address().port);
  } finally {
    await new Promise(resolve => server.close(resolve));
  }
}

function handler(options = {}) {
  return createCodeHandler({
    env: ENV,
    readSecret: async path => path.endsWith('/access') ? APPLICATION_SECRET : 'test-only-forgejo-provider-secret',
    readAllowlist: async () => new Set(['goreecloud/example']),
    fetchImpl: async () => new Response(JSON.stringify({
      full_name: 'GoreeCloud/example', description: 'Example', private: true,
      archived: false, default_branch: 'main', internal_secret: 'MUST-NOT-LEAK',
    }), { status: 200 }),
    ...options,
  });
}

test('health reports development state', async () => withServer(handler(), async base => {
  const res = await fetch(base + '/healthz');
  assert.equal(res.status, 200);
  assert.equal((await res.json()).state, 'development');
}));

test('unauthenticated repo reads fail closed', async () => withServer(handler(), async base => {
  const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example');
  assert.equal(res.status, 401);
}));

test('authorized reads return a minimal provider-neutral projection', async () => withServer(handler(), async base => {
  const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example', {
    headers: { authorization: 'Bearer ' + APPLICATION_SECRET },
  });
  assert.equal(res.status, 200);
  assert.deepEqual(Object.keys(await res.json()).sort(), [
    'archived', 'defaultBranch', 'description', 'fullName', 'private',
  ]);
  assert.equal(res.headers.get('cache-control'), 'no-store');
}));

test('all mutation attempts are refused before calling the provider', async () => {
  let called = false;
  await withServer(handler({ fetchImpl: async () => { called = true; } }), async base => {
    const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example', { method: 'POST' });
    assert.equal(res.status, 405);
    assert.equal(called, false);
  });
});

test('missing security configuration refuses access', async () => withServer(
  createCodeHandler({ env: {}, fetchImpl: async () => { throw Error('unexpected'); } }),
  async base => assert.equal((await fetch(base + '/api/v1/repositories/GoreeCloud/example')).status, 503),
));

test('provider errors redact upstream details', async () => withServer(
  handler({ fetchImpl: async () => new Response('sensitive upstream', { status: 403 }) }),
  async base => {
    const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example', {
      headers: { authorization: 'Bearer ' + APPLICATION_SECRET },
    });
    assert.equal(res.status, 502);
    assert.deepEqual(await res.json(), { error: 'provider_unavailable' });
  },
));

test('oversized upstream payloads fail closed', async () => withServer(
  handler({ fetchImpl: async () => new Response('x'.repeat(70000), { status: 200 }) }),
  async base => {
    const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example', {
      headers: { authorization: 'Bearer ' + APPLICATION_SECRET },
    });
    assert.equal(res.status, 502);
  },
));

test('path and query abuse are not passed to upstream', async () => withServer(
  handler({ fetchImpl: async () => { throw Error('unexpected'); } }),
  async base => {
    for (const suffix of ['?private=true', '/../../admin', '/x%2Fy']) {
      const res = await fetch(base + '/api/v1/repositories/GoreeCloud/example' + suffix);
      assert.equal(res.status, 404);
    }
  },
));

test('provider origin cannot be downgraded or redirected by configuration', () => {
  for (const url of ['http://forgejo.example', 'https://user:password@forgejo.example',
    'https://forgejo.example/path', 'https://forgejo.example/?redirect=x']) {
    assert.throws(() => forgejoOrigin(url));
  }
});

test('listed repository restriction blocks provider access entirely', async () => {
  let called = false;
  await withServer(handler({
    readAllowlist: async () => new Set(['goreecloud/approved']),
    fetchImpl: async () => { called = true; },
  }), async base => {
    const response = await fetch(base + '/api/v1/repositories/GoreeCloud/example', {
      headers: { authorization: 'Bearer ' + APPLICATION_SECRET },
    });
    assert.equal(response.status, 403);
    assert.equal(called, false);
  });
});

test('provider repository identity must match requested authorized repository', async () => {
  await withServer(handler({ fetchImpl: async () =>
    new Response(JSON.stringify({ full_name: 'OtherOrg/secret', private: true })) }), async base => {
    const response = await fetch(base + '/api/v1/repositories/GoreeCloud/example', {
      headers: { authorization: 'Bearer ' + APPLICATION_SECRET },
    });
    assert.equal(response.status, 502);
    assert.deepEqual(await response.json(), { error: 'provider_invalid_response' });
  });
});
