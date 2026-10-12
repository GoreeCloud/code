import {emojiString, emojiHTML} from './features/emoji.js';

test('emojiString', () => {
  expect(emojiString('+1')).toEqual('👍');
  expect(emojiString('arrow_right')).toEqual('➡️');
  expect(emojiString('european_union')).toEqual('🇪🇺');
  expect(emojiString('eu')).toEqual('🇪🇺');

  expect(emojiString('forgejo')).toEqual(':forgejo:');
  expect(emojiString('frogejo')).toEqual(':frogejo:');
  expect(emojiString('blobnom')).toEqual(':blobnom:');

  expect(emojiString('not-a-emoji')).toEqual(':not-a-emoji:');
});

test('emojiHTML', () => {
  expect(emojiHTML('+1')).toEqual('<span class="emoji" title=":+1:">👍</span>');
  expect(emojiHTML('arrow_right')).toEqual('<span class="emoji" title=":arrow_right:">➡️</span>');
  expect(emojiHTML('european_union')).toEqual('<span class="emoji" title=":european_union:">🇪🇺</span>');
  expect(emojiHTML('eu')).toEqual('<span class="emoji" title=":eu:">🇪🇺</span>');

  expect(emojiHTML('forgejo')).toEqual('<span class="emoji" title=":forgejo:"><img alt=":forgejo:" src="/assets/img/emoji/forgejo.png"></span>');
  expect(emojiHTML('frogejo')).toEqual('<span class="emoji" title=":frogejo:"><img alt=":frogejo:" src="/assets/img/emoji/frogejo.png"></span>');
  expect(emojiHTML('blobnom')).toEqual('<span class="emoji" title=":blobnom:"><img alt=":blobnom:" src="/assets/img/emoji/blobnom.png"></span>');

  expect(emojiHTML('not-a-emoji')).toEqual('<span class="emoji" title=":not-a-emoji:">:not-a-emoji:</span>');
});
