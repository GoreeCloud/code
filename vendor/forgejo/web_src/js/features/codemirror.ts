import {isDarkTheme} from '../utils.js';
import {languages} from './codemirror-lang.ts';
import type {LanguageDescription} from '@codemirror/language';
import type {Compartment} from '@codemirror/state';
import type {EditorView, ViewUpdate} from '@codemirror/view';
import {searchPanel} from './codemirror-search.ts';

// Export editor for customization - https://github.com/go-gitea/gitea/issues/10409
function exportEditor(editor: EditorView) {
  if (!window.codeEditors) window.codeEditors = new Set<EditorView>();
  window.codeEditors.add(editor);
}

export interface EditorOptions {
  indentSize?: number;
  tabSize?: number;
  wordWrap: boolean;
  indentStyle: string;
  onContentChange?: (update: ViewUpdate) => void;
}

export interface CodemirrorEditor {
  codemirrorView: CodeMirrorView;
  codemirrorLanguage: CodeMirrorLanguage;
  codemirrorState: CodeMirrorState;
  codemirrorSearch: CodeMirrorSearch;
  view: EditorView;
  languages: LanguageDescription[];
  compartments: {
    wordWrap: Compartment;
    language: Compartment;
    tabSize: Compartment;
  };
}

export async function createCodemirror(
  textarea: HTMLTextAreaElement,
  filename: string,
  editorOpts: EditorOptions,
): Promise<CodemirrorEditor> {
  const codemirrorView = await import(/* webpackChunkName: "codemirror" */ '@codemirror/view');
  const codemirrorCommands = await import(/* webpackChunkName: "codemirror" */ '@codemirror/commands');
  const codemirrorState = await import(/* webpackChunkName: "codemirror" */ '@codemirror/state');
  const codemirrorSearch = await import(/* webpackChunkName: "codemirror" */ '@codemirror/search');
  const codemirrorLanguage = await import(/* webpackChunkName: "codemirror" */ '@codemirror/language');
  const codemirrorAutocomplete = await import(/* webpackChunkName: "codemirror" */ '@codemirror/autocomplete');
  const {tags: t} = await import(/* webpackChunkName: "codemirror" */ '@lezer/highlight');

  const languageDescriptions = languages(codemirrorLanguage);
  const code = codemirrorLanguage.LanguageDescription.matchFilename(
    languageDescriptions,
    filename,
  );
  const onContentChange = editorOpts.onContentChange || ((update) => {
    if (update.docChanged) {
      textarea.value = update.state.doc.toString();
        // Make jquery-are-you-sure happy.
      textarea.dispatchEvent(new Event('change'));
    }
  });

  const darkTheme = isDarkTheme();

  const theme = codemirrorView.EditorView.theme(
    {
      '&': {
        color: 'var(--color-text)',
        backgroundColor: 'var(--color-code-bg)',
        maxHeight: '90vh',
      },
      '.cm-content, .cm-gutter': {
        minHeight: '200px',
      },
      '.cm-scroller': {
        overflow: 'auto',
      },
      '.cm-content': {
        caretColor: 'var(--color-caret)',
        fontFamily: 'var(--fonts-monospace)',
        fontSize: '14px',
      },
      '.cm-cursor, .cm-dropCursor': {
        borederLeftCursor: 'var(--color-caret)',
      },
      '&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground':
        {
          backgroundColor: 'var(--color-primary-light-3)',
        },
      '.cm-panels': {
        backgroundColor: 'var(--color-body)',
        borderColor: 'var(--color-secondary)',
      },
      '.cm-activeLine, .cm-activeLineGutter': {
        backgroundColor: '#6699ff0b',
      },
      '.cm-gutters': {
        backgroundColor: 'var(--color-code-bg)',
        color: 'var(--color-secondary-dark-6)',
      },
      '.cm-line ::selection, .cm-line::selection': {
        color: 'inherit !important',
      },
      '.cm-searchMatch': {
        backgroundColor: '#72a1ff59',
        outline: `1px solid #ffffff0f`,
      },
      '.cm-tooltip.cm-tooltip-autocomplete > ul > li': {
        padding: '0.5em 0.5em',
      },
    },
    {dark: isDarkTheme()},
  );

  const highlightStyle = codemirrorLanguage.HighlightStyle.define([
    {
      tag: [t.keyword, t.operatorKeyword, t.modifier, t.color, t.constant(t.name), t.standard(t.name), t.standard(t.tagName), t.special(t.brace), t.atom, t.bool, t.special(t.variableName)],
      color: darkTheme ? '#569cd6' : '#0064ff',
    },
    {tag: [t.controlKeyword, t.moduleKeyword], color: darkTheme ? '#c586c0' : '#af00db'},
    {
      tag: [t.name, t.deleted, t.character, t.macroName, t.propertyName, t.variableName, t.labelName, t.definition(t.name)],
      color: darkTheme ? '#9cdcfe' : '#383a42',
    },
    {
      tag: [t.typeName, t.className, t.tagName, t.number, t.changed, t.annotation, t.self, t.namespace],
      color: darkTheme ? '#4ec9b0' : '#267f99',
    },
    {
      tag: [t.function(t.variableName), t.function(t.propertyName)],
      color: darkTheme ? '#dcdcaa' : '#795e26',
    },
    {tag: [t.number], color: darkTheme ? '#b5cea8' : '#098658'},
    {
      tag: [t.operator, t.punctuation, t.separator, t.url, t.escape, t.regexp],
      color: darkTheme ? '#d4d4d4' : '#383a42',
    },
    {tag: [t.regexp], color: darkTheme ? '#d16969' : '#af00db'},
    {
      tag: [t.special(t.string), t.processingInstruction, t.string, t.inserted],
      color: darkTheme ? '#ce9178' : '#a31515',
    },
    {tag: [t.meta, t.comment], color: darkTheme ? '#6a9955' : '#6b6b6b'},
    {tag: t.invalid, color: darkTheme ? '#ff0000' : '#e51400'},
    {tag: t.strong, fontWeight: 'bold'},
    {tag: t.emphasis, fontStyle: 'italic'},
    {tag: t.strikethrough, textDecoration: 'line-through'},
    {tag: t.link, color: darkTheme ? '#6a9955' : '#006ab1', textDecoration: 'underline'},
  ]);

  const container = textarea.parentNode.querySelector('.codemirror-container');

  const wordWrap = new codemirrorState.Compartment();
  const language = new codemirrorState.Compartment();
  const tabSize = new codemirrorState.Compartment();

  const view = new codemirrorView.EditorView({
    doc: textarea.value,
    parent: container,
    extensions: [
      codemirrorView.lineNumbers(),
      codemirrorLanguage.foldGutter(),
      codemirrorView.highlightActiveLineGutter(),
      codemirrorView.highlightSpecialChars(),
      codemirrorView.highlightActiveLine(),
      codemirrorView.drawSelection(),
      codemirrorView.dropCursor(),
      codemirrorSearch.search({createPanel: searchPanel(codemirrorSearch)}),
      codemirrorView.keymap.of([
        ...codemirrorAutocomplete.closeBracketsKeymap,
        ...codemirrorCommands.defaultKeymap,
        ...codemirrorCommands.historyKeymap,
        // If no search panel, then disable the search keymap
        ...(document.getElementById('editor-find') ? codemirrorSearch.searchKeymap : []),
        ...codemirrorLanguage.foldKeymap,
        ...codemirrorAutocomplete.completionKeymap,
        codemirrorCommands.indentWithTab,
      ]),
      codemirrorState.EditorState.allowMultipleSelections.of(true),
      codemirrorLanguage.indentOnInput(),
      codemirrorLanguage.syntaxHighlighting(highlightStyle),
      codemirrorLanguage.bracketMatching(),
      codemirrorLanguage.indentUnit.of(
        editorOpts.indentStyle === 'tab' ?
          '\t' :
          ' '.repeat(editorOpts.indentSize || 2),
      ),
      codemirrorAutocomplete.closeBrackets(),
      codemirrorAutocomplete.autocompletion(),
      codemirrorCommands.history(),
      tabSize.of(
        codemirrorState.EditorState.tabSize.of(editorOpts.tabSize || 4),
      ),
      wordWrap.of(
        editorOpts.wordWrap ? codemirrorView.EditorView.lineWrapping : [],
      ),
      language.of(code ? await code.load() : []),
      codemirrorView.EditorView.updateListener.of(onContentChange),
      theme,
    ],
  });

  exportEditor(view);

  container.querySelector('.editor-loading')?.remove();
  return {
    codemirrorView,
    codemirrorState,
    codemirrorLanguage,
    codemirrorSearch,
    view,
    languages: languageDescriptions,
    compartments: {
      tabSize,
      wordWrap,
      language,
    },
  };
}
