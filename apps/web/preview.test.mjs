import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const html = await readFile(new URL('./index.html', import.meta.url), 'utf8');
const css = await readFile(new URL('./styles.css', import.meta.url), 'utf8');

test('preview identifies itself as development rather than a live forge', () => {
  assert.match(html, /Development preview/i);
  assert.match(html, /static design preview/i);
  assert.match(html, /not a finished forge/i);
});

test('preview exposes essential navigation and presentation fallback hooks', () => {
  assert.match(html, /lang="en"/);
  assert.match(html, /class="skip-link"/);
  assert.match(html, /id="content"/);
  assert.match(css, /:focus-visible/);
  assert.match(css, /prefers-reduced-motion/);
  assert.match(css, /prefers-contrast/);
  assert.match(css, /forced-colors/);
});
