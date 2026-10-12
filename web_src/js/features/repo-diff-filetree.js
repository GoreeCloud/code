import {createApp} from 'vue';

export async function initDiffFileTree() {
  const el = document.getElementById('diff-file-tree');
  if (!el) return;

  const {default: DiffFileTree} = await import(/* webpackChunkName: "diff-file-tree" */'../components/DiffFileTree.vue');
  const fileTreeView = createApp(DiffFileTree);
  fileTreeView.mount(el);
}
