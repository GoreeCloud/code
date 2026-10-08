import {createApp} from 'vue';

export async function initDashboardRepoList() {
  const el = document.getElementById('dashboard-repo-list');
  if (!el) {
    return;
  }

  const {default: DashboardRepoList} = await import(/* webpackChunkName: "dashboard-repo-list" */'../components/DashboardRepoList.vue');
  const dashboardList = createApp(DashboardRepoList);
  dashboardList.mount(el);
}
