import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, symlink, chmod, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parseRepositoryAllowlist, readRepositoryAllowlist, repositoryAuthorized } from './access-policy.mjs';

test('only bounded, explicit repository lists are valid', () => {
  assert.deepEqual([...parseRepositoryAllowlist('{"repositories":["GoreeCloud/Code"]}')], ['goreecloud/code']);
  for (const invalid of [
    '{}', '{"repositories":[]}', '{"repositories":["a/b","A/B"]}',
    '{"repositories":["../x"]}', '{"repositories":["a/b/c"]}',
    '{"repositories":["a/b"],"wildcard":true}',
    '{"repositories":["a/*"]}', '{"repositories":["a/b?"]}',
  ]) assert.throws(() => parseRepositoryAllowlist(invalid));
});

test('deny unlisted repositories and normalize case', () => {
  const list = parseRepositoryAllowlist('{"repositories":["GoreeCloud/Code"]}');
  assert.equal(repositoryAuthorized(list, 'goreecloud', 'CODE'), true);
  assert.equal(repositoryAuthorized(list, 'other', 'code'), false);
  assert.equal(repositoryAuthorized(list, 'goreecloud', 'secret'), false);
});

test('policy file must be an owner-only regular file, never a symlink', async () => {
  const dir = await mkdtemp(join(tmpdir(), 'code-policy-test-'));
  const policy = join(dir, 'policy.json');
  const shortcut = join(dir, 'link.json');
  try {
    await writeFile(policy, '{"repositories":["GoreeCloud/Code"]}', { mode: 0o600 });
    assert.equal((await readRepositoryAllowlist(policy)).has('goreecloud/code'), true);
    await chmod(policy, 0o644);
    await assert.rejects(() => readRepositoryAllowlist(policy));
    await chmod(policy, 0o600);
    await symlink(policy, shortcut);
    await assert.rejects(() => readRepositoryAllowlist(shortcut));
  } finally {
    await rm(dir, { recursive: true, force: true });
  }
});
