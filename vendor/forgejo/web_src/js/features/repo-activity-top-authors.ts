import {createApp} from 'vue';

export async function initRepoActivityTopAuthorsChart() {
  const el = document.getElementById('repo-activity-top-authors-chart');
  if (!el) {
    return;
  }

  const {default: RepoActivityTopAuthors} = await import(/* webpackChunkName: "repo-activity-top-authors" */'../components/RepoActivityTopAuthors.vue');
  const repoActivityTopAuthors = createApp(RepoActivityTopAuthors, {
    locale: {
      commitActivity: el.getAttribute('data-locale-commit-activity'),
    },
  });
  repoActivityTopAuthors.mount(el);
}
