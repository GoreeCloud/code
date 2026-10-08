import { open } from 'node:fs/promises';
import { constants } from 'node:fs';
import { resolve } from 'node:path';

const REPO_NAME = /^[A-Za-z0-9][A-Za-z0-9_.-]{0,99}$/;

/**
 * A restricted development allowlist, not a substitute for user-level Identity/Policy.
 * The file must be owner-controlled; it is loaded for each request so revocations
 * don't depend on restarting the API.
 */
export function parseRepositoryAllowlist(raw) {
  if (typeof raw !== 'string' || raw.length > 16384) throw Error('invalid policy size');
  const value = JSON.parse(raw);
  if (!value || Array.isArray(value) || Object.keys(value).sort().join(',') !== 'repositories' ||
      !Array.isArray(value.repositories) || value.repositories.length < 1 ||
      value.repositories.length > 128) {
    throw Error('invalid allowlist structure');
  }
  const unique = new Set();
  for (const repository of value.repositories) {
    if (typeof repository !== 'string' || repository.length > 181) throw Error('invalid repository');
    const segments = repository.split('/');
    if (segments.length !== 2 || segments.some(p => p === '.' || p === '..' || !REPO_NAME.test(p))) {
      throw Error('invalid repository');
    }
    const canonical = repository.toLowerCase();
    if (unique.has(canonical)) throw Error('duplicate repository');
    unique.add(canonical);
  }
  return unique;
}

export async function readRepositoryAllowlist(filePath) {
  if (typeof filePath !== 'string' || !filePath) throw Error('missing policy path');
  const fd = await open(resolve(filePath), constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const info = await fd.stat();
    if (!info.isFile() || info.size < 20 || info.size > 16384 ||
        (info.mode & 0o077) !== 0 ||
        (typeof process.getuid === 'function' && info.uid !== process.getuid())) {
      throw Error('unsafe allowlist file');
    }
    return parseRepositoryAllowlist(await fd.readFile('utf8'));
  } finally {
    await fd.close();
  }
}

export function repositoryAuthorized(allowlist, owner, name) {
  if (!(allowlist instanceof Set)) throw Error('invalid loaded policy');
  return allowlist.has((owner + '/' + name).toLowerCase());
}
