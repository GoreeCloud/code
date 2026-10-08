import test from 'node:test';
import assert from 'node:assert/strict';
import { projectRepository, projectCollection } from './provider-projections.mjs';

test('repository identity is bound to request and extra fields are excluded', () => {
  const mapped = projectRepository({ full_name: 'GoreeCloud/Code', private: true,
    description: 'test', admin_token: 'not for clients' }, 'goreecloud', 'code');
  assert.equal(mapped.fullName, 'GoreeCloud/Code');
  assert.equal('admin_token' in mapped, false);
  assert.throws(() => projectRepository({ full_name: 'Other/private' }, 'goreecloud', 'code'));
});

test('collection projection is bounded and strongly typed', () => {
  const items = Array.from({ length: 29 }, (_, i) => ({
    number: i + 1, title: 'Example', state: 'open', body: 'private', user: { login: 'tester' },
  }));
  const result = projectCollection('issues', items);
  assert.equal(result.count, 20);
  assert.equal('body' in result.items[0], false);
  assert.throws(() => projectCollection('issues', {items: []}));
  assert.throws(() => projectCollection('issues', [{number: 'not-a-number', title: 'Bad'}]));
  assert.throws(() => projectCollection('invalid', [{}]));
});
