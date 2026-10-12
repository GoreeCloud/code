import {createApp} from 'vue';

export async function initDiffCommitSelect() {
  const el = document.getElementById('diff-commit-select');
  if (!el) return;

  const {default: DiffCommitSelector} = await import(/* webpackChunkName: "diff-commit-selector" */'../components/DiffCommitSelector.vue');
  const commitSelect = createApp(DiffCommitSelector);
  commitSelect.mount(el);
}
