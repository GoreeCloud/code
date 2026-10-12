// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

import type {SearchQuery} from '@codemirror/search';
import type {EditorView, Panel, ViewUpdate} from '@codemirror/view';
import {svg} from '../svg.js';

class SearchPanel implements Panel {
  searchField: HTMLInputElement;
  replaceField: HTMLInputElement;
  caseField: HTMLInputElement;
  caseLabel: HTMLLabelElement;
  reField: HTMLInputElement;
  reLabel: HTMLLabelElement;
  wordField: HTMLInputElement;
  wordLabel: HTMLLabelElement;
  dom: HTMLElement;
  query: SearchQuery;
  search: CodeMirrorSearch;

  constructor(readonly codemirrorSearch: CodeMirrorSearch, readonly view: EditorView) {
    this.search = codemirrorSearch;

    const container = view.dom.parentElement;

    const query = (this.query = this.search.getSearchQuery(view.state));
    this.commit = this.commit.bind(this);

    this.searchField = document.createElement('input');
    this.searchField.value = query.search;
    this.searchField.name = 'search';
    const searchText = container.getAttribute('data-search-text');
    this.searchField.placeholder = searchText;
    this.searchField.ariaLabel = searchText;
    this.searchField.classList.add('cm-textfield');
    this.searchField.setAttribute('main-field', 'true');
    this.searchField.addEventListener('keyup', this.commit);
    this.searchField.addEventListener('change', this.commit);

    this.caseField = document.createElement('input');
    this.caseField.checked = query.caseSensitive;
    this.caseField.type = 'checkbox';
    this.caseField.name = 'case_sensitive';
    this.caseField.id = 'search_case_sensitive';
    this.caseField.addEventListener('change', this.commit);
    this.caseField.addEventListener('focus', () => this.updateLabels());
    this.caseField.addEventListener('blur', () => this.updateLabels());

    this.caseLabel = document.createElement('label');
    this.caseLabel.setAttribute('for', 'search_case_sensitive');
    const caseText = container.getAttribute('data-toggle-case-text');
    this.caseLabel.ariaLabel = caseText;
    this.caseLabel.setAttribute('data-tooltip-content', caseText);
    this.caseLabel.textContent = 'aA';

    this.reField = document.createElement('input');
    this.reField.checked = query.regexp;
    this.reField.type = 'checkbox';
    this.reField.name = 'regexp';
    this.reField.id = 'search_regexp';
    this.reField.addEventListener('change', this.commit);
    this.reField.addEventListener('focus', () => this.updateLabels());
    this.reField.addEventListener('blur', () => this.updateLabels());

    this.reLabel = document.createElement('label');
    this.reLabel.setAttribute('for', 'search_regexp');
    const reText = container.getAttribute('data-toggle-regex-text');
    this.reLabel.ariaLabel = reText;
    this.reLabel.setAttribute('data-tooltip-content', reText);
    this.reLabel.textContent = '[.+]';

    this.wordField = document.createElement('input');
    this.wordField.checked = query.wholeWord;
    this.wordField.type = 'checkbox';
    this.wordField.name = 'by_word';
    this.wordField.id = 'search_by_word';
    this.wordField.addEventListener('change', this.commit);
    this.wordField.addEventListener('focus', () => this.updateLabels());
    this.wordField.addEventListener('blur', () => this.updateLabels());

    this.wordLabel = document.createElement('label');
    this.wordLabel.setAttribute('for', 'search_by_word');
    const wholeWordText = container.getAttribute('data-toggle-whole-word-text');
    this.wordLabel.ariaLabel = wholeWordText;
    this.wordLabel.setAttribute('data-tooltip-content', wholeWordText);
    this.wordLabel.textContent = 'W';

    this.updateLabels();

    const searchFieldContainer = document.createElement('span');
    searchFieldContainer.classList.add('search-input-group');
    searchFieldContainer.replaceChildren(this.searchField, this.caseLabel, this.reLabel, this.wordLabel);

    const hiddenInputs = document.createElement('div');
    hiddenInputs.classList.add('search-hidden-inputs');
    hiddenInputs.replaceChildren(this.caseField, this.reField, this.wordField);

    const prevSearch = document.createElement('button');
    prevSearch.classList.add('secondary', 'button');
    prevSearch.type = 'button';
    const findPrevText = container.getAttribute('data-find-prev-text');
    prevSearch.ariaLabel = findPrevText;
    prevSearch.addEventListener('click', () => {
      this.search.findPrevious(view);
    });
    prevSearch.innerHTML = svg('octicon-arrow-up');

    const nextSearch = document.createElement('button');
    nextSearch.classList.add('secondary', 'button');
    nextSearch.type = 'button';
    const findNextText = container.getAttribute('data-find-next-text');
    nextSearch.ariaLabel = findNextText;
    nextSearch.addEventListener('click', () => {
      this.search.findNext(view);
    });
    nextSearch.innerHTML = svg('octicon-arrow-down');

    const searchSection = document.createElement('div');
    searchSection.classList.add('search-section');
    searchSection.replaceChildren(searchFieldContainer, hiddenInputs, prevSearch, nextSearch);

    this.replaceField = document.createElement('input');
    this.replaceField.value = query.replace;
    this.replaceField.name = 'replace';
    const replaceText = container.getAttribute('data-replace-text');
    this.replaceField.placeholder = replaceText;
    this.replaceField.ariaLabel = replaceText;
    this.replaceField.classList.add('cm-textfield');
    this.replaceField.addEventListener('keyup', this.commit);
    this.replaceField.addEventListener('change', this.commit);

    const replaceButton = document.createElement('button');
    replaceButton.classList.add('secondary', 'button');
    replaceButton.type = 'button';
    replaceButton.addEventListener('click', () => {
      this.search.replaceNext(view);
    });
    replaceButton.textContent = replaceText;

    const replaceAllButton = document.createElement('button');
    replaceAllButton.classList.add('secondary', 'button');
    replaceAllButton.type = 'button';
    replaceAllButton.addEventListener('click', () => {
      this.search.replaceAll(view);
    });
    const replaceAllText = container.getAttribute('data-replace-all-text');
    replaceAllButton.textContent = replaceAllText;

    const replaceSection = document.createElement('div');
    replaceSection.classList.add('replace-section');
    replaceSection.replaceChildren(this.replaceField, replaceButton, replaceAllButton);

    this.dom = document.createElement('div');
    this.dom.classList.add('fj-search');
    this.dom.addEventListener('keydown', (e: KeyboardEvent) => this.keydown(e));
    this.dom.replaceChildren(searchSection, replaceSection);
  }

  commit() {
    this.updateLabels();
    const query = new this.search.SearchQuery({
      search: this.searchField.value,
      caseSensitive: this.caseField.checked,
      regexp: this.reField.checked,
      wholeWord: this.wordField.checked,
      replace: this.replaceField.value,
    });
    if (!query.eq(this.query)) {
      this.query = query;
      this.view.dispatch({effects: this.search.setSearchQuery.of(query)});
      // Set the new search query and reset the selection
      const anchor = this.view.state.selection.main.anchor;
      this.view.dispatch({
        selection: {anchor},
        effects: this.search.setSearchQuery.of(query),
      });
    }
  }

  keydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && e.target === this.searchField) {
      e.preventDefault();
      if (e.shiftKey) {
        this.search.findPrevious(this.view);
      } else {
        this.search.findNext(this.view);
      }
    } else if (e.key === 'Enter' && e.target === this.replaceField) {
      e.preventDefault();
      this.search.replaceNext(this.view);
    }
  }

  update(update: ViewUpdate) {
    for (const tr of update.transactions) for (const effect of tr.effects) {
      if (effect.is(this.search.setSearchQuery) && !effect.value.eq(this.query)) {
        this.setQuery(effect.value);
      }
    }
  }

  setQuery(query: SearchQuery) {
    this.query = query;
    this.searchField.value = query.search;
    this.replaceField.value = query.replace;
    this.caseField.checked = query.caseSensitive;
    this.reField.checked = query.regexp;
    this.wordField.checked = query.wholeWord;
    this.updateLabels();
  }

  updateLabels() {
    this.caseLabel.classList.toggle('active', this.caseField.checked);
    this.caseLabel.classList.toggle('focused', this.caseField === document.activeElement);
    this.reLabel.classList.toggle('active', this.reField.checked);
    this.reLabel.classList.toggle('focused', this.reField === document.activeElement);
    this.wordLabel.classList.toggle('active', this.wordField.checked);
    this.wordLabel.classList.toggle('focused', this.wordField === document.activeElement);
  }

  mount() {
    this.searchField.select();
  }

  get pos() {
    return 80;
  }

  get top() {
    return true;
  }
}

export function searchPanel(
  codemirrorSearch: CodeMirrorSearch,
): (view: EditorView) => Panel {
  return (view) => {
    return new SearchPanel(codemirrorSearch, view);
  };
}
