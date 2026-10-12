import {matchEmoji, matchMention} from '../../utils/match.js';
import {emojiHTML, emojiString} from '../emoji.js';
const {customEmojis} = window.config;

export function initTextExpander(expander) {
  expander?.addEventListener('text-expander-change', ({detail: {key, provide, text}}) => {
    if (key === ':') {
      const matches = matchEmoji(text);
      if (!matches.length) return provide({matched: false});

      const ul = document.createElement('ul');
      ul.classList.add('suggestions');
      for (const name of matches) {
        const li = document.createElement('li');
        li.setAttribute('id', `combobox-emoji-${name}`);
        li.setAttribute('role', 'option');
        li.setAttribute('data-value', emojiString(name));
        if (customEmojis.has(name)) {
          li.style.gap = '0.25rem';
          li.innerHTML = emojiHTML(name);
          li.append(name);
        } else {
          li.textContent = `${emojiString(name)} ${name}`;
        }
        ul.append(li);
      }

      provide({matched: true, fragment: ul});
    } else if (key === '@') {
      const matches = matchMention(text);
      if (!matches.length) return provide({matched: false});

      const ul = document.createElement('ul');
      ul.classList.add('suggestions');
      for (const {value, name, fullname, avatar} of matches) {
        const li = document.createElement('li');
        li.setAttribute('id', `combobox-user-${name}`);
        li.setAttribute('role', 'option');
        li.setAttribute('data-value', `${key}${value}`);

        const img = document.createElement('img');
        img.setAttribute('aria-hidden', 'true');
        img.src = avatar;
        li.append(img);

        const nameSpan = document.createElement('span');
        nameSpan.textContent = name;
        li.append(nameSpan);

        if (fullname && fullname.toLowerCase() !== name) {
          const fullnameSpan = document.createElement('span');
          fullnameSpan.classList.add('fullname');
          fullnameSpan.textContent = fullname;
          li.append(fullnameSpan);
        }

        ul.append(li);
      }

      provide({matched: true, fragment: ul});
    }
  });
  expander?.addEventListener('text-expander-value', ({detail}) => {
    if (detail?.item) {
      // add a space after @mentions as it's likely the user wants one
      const suffix = detail.key === '@' ? ' ' : '';
      detail.value = `${detail.item.getAttribute('data-value')}${suffix}`;
    }
  });
}
