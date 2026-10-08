import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const fromRoot = path => readFile(new URL('../' + path, import.meta.url), 'utf8');

test('development API routes appear in the platform contract and manual', async () => {
  const contract = await fromRoot('goreecloud.platform.yaml');
  const manual = await fromRoot('docs/USER-MANUAL.md');
  for (const suffix of ['', '/branches', '/issues', '/pulls']) {
    const route = '/api/v1/repositories/{owner}/{name}' + suffix;
    assert(contract.includes(route), 'missing platform contract route: ' + route);
    assert(manual.includes(suffix || '/api/v1/repositories'), 'manual must document route: ' + route);
  }
  assert.match(contract, /lifecycle: development/);
  assert.match(contract, /qualification_state: not-run/);
});

test('source attribution and Glaze acceptance boundaries stay explicit', async () => {
  const manifest = await fromRoot('vendor/forgejo/UPSTREAM-SNAPSHOT.md');
  const upstream = await fromRoot('docs/UPSTREAM.md');
  const readme = await fromRoot('README.md');
  const glide = await fromRoot('docs/GLAZE-ADOPTION.md');
  assert.match(manifest, /v15\.0\.9/);
  assert.match(upstream, /19b9b9d216bbfb501c18514bd1a8c980246ca3f7/);
  assert.match(readme, /Development/);
  assert.match(glide, /does not consume.*Glaze V1.7/i);
});
