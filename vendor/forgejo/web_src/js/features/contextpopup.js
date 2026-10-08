import {createApp} from 'vue';
import {parseIssueHref} from '../utils.js';
import {createTippy} from '../modules/tippy.js';

export function initContextPopups() {
  const refIssues = document.querySelectorAll('.ref-issue');
  attachRefIssueContextPopup(refIssues);
}

export async function attachRefIssueContextPopup(refIssues) {
  for (const refIssue of refIssues) {
    if (refIssue.classList.contains('ref-external-issue')) {
      return;
    }

    const {owner, repo, index} = parseIssueHref(refIssue.getAttribute('href'));
    if (!owner) return;

    const el = document.createElement('div');
    refIssue.parentNode.insertBefore(el, refIssue.nextSibling);

    const {default: ContextPopup} = await import(/* webpackChunkName: "context-popup" */'../components/ContextPopup.vue');
    const view = createApp(ContextPopup);

    try {
      view.mount(el);
    } catch (err) {
      console.error(err);
      el.textContent = 'ContextPopup failed to load';
    }

    createTippy(refIssue, {
      content: el,
      placement: 'top-start',
      interactive: true,
      role: 'dialog',
      interactiveBorder: 5,
      onShow: () => {
        el.firstChild.dispatchEvent(new CustomEvent('ce-load-context-popup', {detail: {owner, repo, index}}));
      },
    });
  }
}
