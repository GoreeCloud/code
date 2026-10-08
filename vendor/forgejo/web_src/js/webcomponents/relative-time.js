import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc.js';
const {pageData} = window.config;

dayjs.extend(utc);

export const HALF_MINUTE = 30 * 1000;
export const ONE_MINUTE = 60 * 1000;
export const ONE_HOUR = 60 * ONE_MINUTE;
export const ONE_DAY = 24 * ONE_HOUR;

const ABSOLUTE_DATETIME_FORMAT = new Intl.DateTimeFormat(navigator.language, {
  year: 'numeric',
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  timeZoneName: 'short',
});
const FALLBACK_DATETIME_FORMAT = new Intl.RelativeTimeFormat(navigator.language, {style: 'long'});

/**
 * A list of plural rules for all languages.
 * `plural_rules.go` defines the index for each of the 14 known plural rules.
 *
 * `pageData.PLURAL_RULE_LANG` is the index of the plural rule for the current language.
 * `pageData.PLURAL_RULE_FALLBACK` is the index of the plural rule for the default language,
 * to be used when a string is not translated in the current language.
 *
 * Each plural rule is a function that maps an amount `n` to the appropriate plural form index.
 * Which index means which rule is specific for each language and also defined in `plural_rules.go`.
 * The actual strings are in `pageData.PLURALSTRINGS_LANG` and `pageData.PLURALSTRINGS_FALLBACK`
 * respectively, which is an array indexed by the plural form index.
 *
 * Links to the language plural rule and form definitions:
 * https://codeberg.org/forgejo/forgejo/src/branch/forgejo/modules/translation/plural_rules.go
 * https://www.unicode.org/cldr/charts/46/supplemental/language_plural_rules.html
 * https://translate.codeberg.org/languages/$LANGUAGE_CODE/#information
 * https://github.com/WeblateOrg/language-data/blob/main/languages.csv
 */
const PLURAL_RULES = [
  // [ 0] Common 2-form, e.g. English, German
  function (n) { return n !== 1 ? 1 : 0 },

  // [ 1] Bengali 2-form
  function (n) { return n > 1 ? 1 : 0 },

  // [ 2] Icelandic 2-form
  function (n) { return n % 10 !== 1 || n % 100 === 11 ? 1 : 0 },

  // [ 3] Filipino 2-form
  function (n) { return n !== 1 && n !== 2 && n !== 3 && (n % 10 === 4 || n % 10 === 6 || n % 10 === 9) ? 1 : 0 },

  // [ 4] One form
  function (_) { return 0 },

  // [ 5] Czech 3-form
  function (n) { return (n === 1) ? 0 : (n >= 2 && n <= 4) ? 1 : 2 },

  // [ 6] Russian 3-form
  function (n) { return n % 10 === 1 && n % 100 !== 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20) ? 1 : 2 },

  // [ 7] Polish 3-form
  function (n) { return n === 1 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20) ? 1 : 2 },

  // [ 8] Latvian 3-form
  function (n) { return (n % 10 === 0 || n % 100 >= 11 && n % 100 <= 19) ? 0 : ((n % 10 === 1 && n % 100 !== 11) ? 1 : 2) },

  // [ 9] Lithunian 3-form
  function (n) { return (n % 10 === 1 && (n % 100 < 11 || n % 100 > 19)) ? 0 : ((n % 10 >= 2 && n % 10 <= 9 && (n % 100 < 11 || n % 100 > 19)) ? 1 : 2) },

  // [10] French 3-form
  function (n) { return (n === 0 || n === 1) ? 0 : ((n !== 0 && n % 1000000 === 0) ? 1 : 2) },

  // [11] Catalan 3-form
  function (n) { return (n === 1) ? 0 : ((n !== 0 && n % 1000000 === 0) ? 1 : 2) },

  // [12] Slovenian 4-form
  function (n) { return n % 100 === 1 ? 0 : n % 100 === 2 ? 1 : n % 100 === 3 || n % 100 === 4 ? 2 : 3 },

  // [13] Arabic 6-form
  function (n) { return n === 0 ? 0 : n === 1 ? 1 : n === 2 ? 2 : n % 100 >= 3 && n % 100 <= 10 ? 3 : n % 100 >= 11 ? 4 : 5 },
];

/**
 * Look up the correct localized plural form for amount `n` for the string with the translation key `key`.
 * If the current language does not contain a translation for this key, fallback to the browser's formatting.
 */
function GetPluralizedStringOrFallback(key, n, unit) {
  const translation = pageData.PLURALSTRINGS_LANG[key]?.[PLURAL_RULES[pageData.PLURAL_RULE_LANG](n)];
  if (translation) return translation.replace('%d', n);
  return FALLBACK_DATETIME_FORMAT.format(-n, unit);
}

/**
 * Update the displayed text of the given relative-time DOM element with its
 * human-readable, localized relative time string.
 * Returns the recommended interval in milliseconds until the object should be updated again,
 * or null if the object is invalid.
 */
export function DoUpdateRelativeTime(object, now) {
  const absoluteTime = object.getAttribute('datetime');
  if (!absoluteTime) {
    return null;  // Object does not contain a datetime.
  }

  if (!now) now = Date.now();

  const nowJS = dayjs.utc(now);
  const thenJS = dayjs.utc(absoluteTime);

  object.setAttribute('data-tooltip-content', ABSOLUTE_DATETIME_FORMAT.format(thenJS.toDate()));

  if (nowJS.isBefore(thenJS)) {
    // Datetime is in the future.
    object.textContent = pageData.DATETIMESTRINGS.FUTURE;
    return -Math.floor(nowJS.diff(thenJS, 'millisecond'));
  }

  const years = Math.floor(nowJS.diff(thenJS, 'year'));
  if (years >= 1) {
    // Datetime is at least one year ago.
    if (years === 1 && pageData.DATETIMESTRINGS['relativetime.1year']) {
      // Datetime is one year ago.
      object.textContent = pageData.DATETIMESTRINGS['relativetime.1year'];
    } else {
      // Datetime is more than a year ago.
      object.textContent = GetPluralizedStringOrFallback('relativetime.years', years, 'year');
    }
    return ONE_DAY;
  }

  const months = Math.floor(nowJS.diff(thenJS, 'month'));
  if (months >= 1) {
    // Datetime is at least one month but less than a year ago.
    if (months === 1 && pageData.DATETIMESTRINGS['relativetime.1month']) {
      // Datetime is one month ago.
      object.textContent = pageData.DATETIMESTRINGS['relativetime.1month'];
    } else {
      // Datetime is several months ago (but less than a year).
      object.textContent = GetPluralizedStringOrFallback('relativetime.months', months, 'month');
    }
    return ONE_DAY;
  }

  const weeks = Math.floor(nowJS.diff(thenJS, 'week'));
  if (weeks >= 1) {
    // Datetime is at least one week but less than a month ago.
    if (weeks === 1 && pageData.DATETIMESTRINGS['relativetime.1week']) {
      // Datetime is one week ago.
      object.textContent = pageData.DATETIMESTRINGS['relativetime.1week'];
    } else {
      // Datetime is several weeks ago (but less than a month).
      object.textContent = GetPluralizedStringOrFallback('relativetime.weeks', weeks, 'week');
    }
    return ONE_DAY;
  }

  const days = Math.floor(nowJS.diff(thenJS, 'day'));
  if (days >= 1) {
    if (days === 1 && pageData.DATETIMESTRINGS['relativetime.1day']) {
      // Datetime is one day ago.
      object.textContent = pageData.DATETIMESTRINGS['relativetime.1day'];
    } else {
      // Datetime is several days but less than a week ago.
      object.textContent = GetPluralizedStringOrFallback('relativetime.days', days, 'day');
    }
    return ONE_DAY;
  }

  const hours = Math.floor(nowJS.diff(thenJS, 'hour'));
  if (hours >= 1) {
    // Datetime is one or more hours but less than a day ago.
    object.textContent = GetPluralizedStringOrFallback('relativetime.hours', hours, 'hour');
    return ONE_HOUR;
  }

  const minutes = Math.floor(nowJS.diff(thenJS, 'minute'));
  if (minutes >= 1) {
    // Datetime is one or more minutes but less than an hour ago.
    object.textContent = GetPluralizedStringOrFallback('relativetime.mins', minutes, 'minute');
    return ONE_MINUTE;
  }

  // Datetime is very recent.
  object.textContent = pageData.DATETIMESTRINGS.NOW;
  return HALF_MINUTE;
}

window.customElements.define('relative-time', class extends HTMLElement {
  static observedAttributes = ['datetime'];

  alive = false;
  contentSpan = null;

  update = (recurring) => {
    if (!this.alive) return;

    if (!this.shadowRoot) {
      this.attachShadow({mode: 'open'});
      this.contentSpan = document.createElement('span');
      this.contentSpan.setAttribute('part', 'relative-time');
      this.shadowRoot.append(this.contentSpan);
    }

    const next = DoUpdateRelativeTime(this);
    if (recurring && next !== null) setTimeout(() => { this.update(true) }, next);
  };

  connectedCallback() {
    this.alive = true;
    this.update(true);
  }

  disconnectedCallback() {
    this.alive = false;
  }

  attributeChangedCallback(name, oldValue, newValue) {
    if (name === 'datetime' && oldValue !== newValue) this.update(false);
  }

  set textContent(value) {
    if (this.contentSpan) this.contentSpan.textContent = value;
  }
  get textContent() {
    return this.contentSpan?.textContent;
  }
});
