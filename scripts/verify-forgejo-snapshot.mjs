import assert from 'node:assert/strict';
import { readFile, stat, lstat } from 'node:fs/promises';
import { join } from 'node:path';

const root = new URL('../vendor/forgejo/', import.meta.url);
async function mandatory(relative) {
  const url = new URL(relative, root);
  const info = await lstat(url);
  assert(info.isFile() && info.size > 0, relative + ' must be nonempty regular file');
  return readFile(url, 'utf8');
}

const manifest = await mandatory('UPSTREAM-SNAPSHOT.md');
const license = await mandatory('LICENSE');
const gomod = await mandatory('go.mod');
await mandatory('go.sum');
assert.match(manifest, /Release tag: v15\.0\.9/);
assert.match(manifest, /Resolved commit: [0-9a-f]{40}/);
assert.match(manifest, /Origin: https:\/\/codeberg\.org\/forgejo\/forgejo\.git/);
assert.match(gomod, /^module /m);
assert(license.length > 1000, 'Upstream license must not be truncated');
try {
  await stat(new URL('.git/', root));
  throw Error('Embedded vendor .git is not permitted');
} catch (error) {
  if (error.code !== 'ENOENT') throw error;
}
console.log('Forgejo tracked-source snapshot structure and manifest: PASS');
