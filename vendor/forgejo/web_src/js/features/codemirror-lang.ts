import type {LanguageDescription} from '@codemirror/language';

export function languages(codemirrorLanguage: CodeMirrorLanguage): LanguageDescription[] {
  return [
    codemirrorLanguage.LanguageDescription.of({
      name: 'C',
      extensions: ['c', 'h', 'ino'],
      async load() {
        return (await import('@codemirror/lang-cpp')).cpp();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'C3',
      extensions: ['c3'],
      async load() {
        return (await import('@codemirror/lang-cpp')).cpp();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'C++',
      alias: ['cpp'],
      extensions: ['cpp', 'c++', 'cc', 'cxx', 'hpp', 'h++', 'hh', 'hxx'],
      async load() {
        return (await import('@codemirror/lang-cpp')).cpp();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'CSS',
      extensions: ['css'],
      async load() {
        return (await import('@codemirror/lang-css')).css();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Go',
      extensions: ['go'],
      async load() {
        return (await import('@codemirror/lang-go')).go();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'HTML',
      alias: ['xhtml'],
      extensions: ['html', 'htm', 'handlebars', 'hbs'],
      async load() {
        return (await import('@codemirror/lang-html')).html();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Java',
      extensions: ['java'],
      async load() {
        return (await import('@codemirror/lang-java')).java();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'JavaScript',
      alias: ['ecmascript', 'js', 'node'],
      extensions: ['js', 'mjs', 'cjs'],
      async load() {
        return (await import('@codemirror/lang-javascript')).javascript();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'JSON',
      alias: ['json5'],
      extensions: ['json', 'map'],
      async load() {
        return (await import('@codemirror/lang-json')).json();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'JSX',
      extensions: ['jsx'],
      async load() {
        return (await import('@codemirror/lang-javascript')).javascript({jsx: true});
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'LESS',
      extensions: ['less'],
      async load() {
        return (await import('@codemirror/lang-less')).less();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Liquid',
      extensions: ['liquid'],
      async load() {
        return (await import('@codemirror/lang-liquid')).liquid();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Markdown',
      extensions: ['md', 'markdown', 'mkd'],
      async load() {
        return (await import('@codemirror/lang-markdown')).markdown();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'PHP',
      extensions: ['php', 'php3', 'php4', 'php5', 'php7', 'phtml'],
      async load() {
        return (await import('@codemirror/lang-php')).php();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Python',
      extensions: ['BUILD', 'bzl', 'py', 'pyw'],
      filename: /^(BUCK|BUILD)$/,
      async load() {
        return (await import('@codemirror/lang-python')).python();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Rust',
      extensions: ['rs'],
      async load() {
        return (await import('@codemirror/lang-rust')).rust();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'Sass',
      extensions: ['sass'],
      async load() {
        return (await import('@codemirror/lang-sass')).sass({indented: true});
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'SCSS',
      extensions: ['scss'],
      async load() {
        return (await import('@codemirror/lang-sass')).sass();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'TSX',
      extensions: ['tsx'],
      async load() {
        return (await import('@codemirror/lang-javascript')).javascript({jsx: true, typescript: true});
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'TypeScript',
      alias: ['ts'],
      extensions: ['ts', 'mts', 'cts'],
      async load() {
        return (await import('@codemirror/lang-javascript')).javascript({typescript: true});
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'XML',
      alias: ['rss', 'wsdl', 'xsd'],
      extensions: ['xml', 'xsl', 'xsd', 'svg'],
      async load() {
        return (await import('@codemirror/lang-xml')).xml();
      },
    }),
    codemirrorLanguage.LanguageDescription.of({
      name: 'YAML',
      alias: ['yml'],
      extensions: ['yaml', 'yml'],
      async load() {
        return (await import('@codemirror/lang-yaml')).yaml();
      },
    }),
  ];
}
