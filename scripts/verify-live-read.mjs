import assert from 'node:assert/strict';

const endpoint = process.env.CODE_READ_BASE_URL;
const owner = process.env.CODE_READ_OWNER;
const repository = process.env.CODE_READ_REPOSITORY;
const credential = process.env.CODE_READ_BEARER;

if (!endpoint || !owner || !repository || !credential) {
  throw Error('CODE_READ_BASE_URL, CODE_READ_OWNER, CODE_READ_REPOSITORY and CODE_READ_BEARER are required');
}
const base = new URL(endpoint);
assert.equal(base.protocol, 'http:');
assert(['127.0.0.1', 'localhost', '[::1]'].includes(base.hostname));
assert.equal(base.pathname, '/');

const read = async suffix => {
  const response = await fetch(new URL(
    '/api/v1/repositories/' + encodeURIComponent(owner) + '/' +
    encodeURIComponent(repository) + suffix, base
  ), {
    headers: { authorization: 'Bearer ' + credential },
    signal: AbortSignal.timeout(10000),
  });
  assert.equal(response.status, 200);
  assert.equal(response.headers.get('cache-control'), 'no-store');
  return response.json();
};

const repositoryResult = await read('');
assert.equal(repositoryResult.fullName.toLowerCase(),
  (owner + '/' + repository).toLowerCase());
for (const route of ['/branches', '/issues', '/pulls']) {
  const response = await read(route);
  assert(Array.isArray(response.items));
  assert(response.items.length <= 20);
}
console.log('Code read-only live contract smoke check: PASS');
