import {vi} from 'vitest';

import {issueTitleHTML, excludeLabel} from './repo-issue-sidebar-list.ts';
import {findWipPrefix} from './repo-issue.js';

vi.mock('./comp/ComboMarkdownEditor.js', () => ({}));
// jQuery is missing
vi.mock('./common-global.js', () => ({}));

test('Convert issue title to html', () => {
  expect(issueTitleHTML('')).toEqual('');
  expect(issueTitleHTML('issue title')).toEqual('issue title');

  const expected_thumbs_up = `<span class="emoji" title=":+1:">👍</span>`;
  expect(issueTitleHTML(':+1:')).toEqual(expected_thumbs_up);
  expect(issueTitleHTML(':invalid emoji:')).toEqual(':invalid emoji:');

  const expected_code_block = `<code class="inline-code-block">code</code>`;
  expect(issueTitleHTML('`code`')).toEqual(expected_code_block);
  expect(issueTitleHTML('`invalid code')).toEqual('`invalid code');
  expect(issueTitleHTML('invalid code`')).toEqual('invalid code`');

  expect(issueTitleHTML('issue title :+1: `code`')).toEqual(`issue title ${expected_thumbs_up} ${expected_code_block}`);
});

const getLabelsParam = () => new URLSearchParams(window.location.search).get('labels');

test('Toggles label exclusion from filters', () => {
  expect(getLabelsParam()).toEqual(null);

  const element = document.createElement('div');
  element.dataset['label-id'] = '1';

  // excludes it
  excludeLabel(element);
  expect(getLabelsParam()).toEqual('-1');

  // since it was excluded above, now it should delete it
  excludeLabel(element);
  expect(getLabelsParam()).toEqual('');

  // if we add it manually it should swap it to an exclusion
  window.location.search = '?labels=1';
  expect(getLabelsParam()).toEqual('1');
  excludeLabel(element);
  expect(getLabelsParam()).toEqual('-1');
});

test('Finds wip prefix in string', () => {
  const wipPrefixes = ['wIp:', '[WIP]'];

  expect(findWipPrefix('[wIP]', wipPrefixes)).toBe('[WIP]');
  expect(findWipPrefix('[wip]', wipPrefixes)).toBe('[WIP]');
  expect(findWipPrefix('[WIP]', wipPrefixes)).toBe('[WIP]');

  expect(findWipPrefix('wIP:', wipPrefixes)).toBe('wIp:');
  expect(findWipPrefix('wip:', wipPrefixes)).toBe('wIp:');
  expect(findWipPrefix('WIP:', wipPrefixes)).toBe('wIp:');

  expect(findWipPrefix('wIP:', [])).toBe(undefined);
  expect(findWipPrefix('wIP:', [])).toBe(undefined);
  expect(findWipPrefix('wip:', [])).toBe(undefined);

  expect(findWipPrefix('wip:', ['[WIP]'])).toBe(undefined);
  expect(findWipPrefix('WIP:', ['[WIP]'])).toBe(undefined);
  expect(findWipPrefix('WIP:', ['[WIP]'])).toBe(undefined);
});
