export function initTab(parentEl: Element) {
  if (!parentEl) {
    return;
  }

  // Keep track of which tab is active for this element.
  let activeTabPath = parentEl.querySelector('.item.active')?.getAttribute('data-tab');
  if (!activeTabPath) {
    return;
  }

  for (const el of parentEl.querySelectorAll('.item')) {
    el.addEventListener('click', (ev) => {
      // There's no data-tab attribute we can't do anything, ignore.
      const tabPath = el.getAttribute('data-tab');
      if (!tabPath) {
        return;
      }

      // The item is already active, ignore.
      if (el.classList.contains('active')) {
        return;
      }

      // Make the current item active and the previous item inactive.
      parentEl.querySelector('.item.active').classList.remove('active');
      document.querySelector(`.tab.active[data-tab=${activeTabPath}]`).classList.remove('active');
      el.classList.add('active');
      document.querySelector(`.tab[data-tab=${tabPath}]`).classList.add('active');
      activeTabPath = tabPath;

      // Not really sure if this is useful, it is kept from how Fomantic did it.
      ev.preventDefault();
    }, {passive: false});
  }
}
