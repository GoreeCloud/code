import {createApp} from 'vue';

export async function initRepoPullRequestMergeForm() {
  const el = document.getElementById('pull-request-merge-form');
  if (!el) return;

  const {default: PullRequestMergeForm} = await import(/* webpackChunkName: "pull-request-merge-form" */'../components/PullRequestMergeForm.vue');
  const view = createApp(PullRequestMergeForm);
  view.mount(el);
}
