export function renderExternal() {
  const giteaExternalRender = document.querySelector('iframe.external-render');
  if (!giteaExternalRender) return;

  giteaExternalRender.contentWindow.postMessage({requestOffsetHeight: true}, '*');

  const eventListener = (event) => {
    if (event.source !== giteaExternalRender.contentWindow) return;
    const height = Number(event.data?.frameHeight);
    if (!height) return;
    giteaExternalRender.height = height;
    giteaExternalRender.style.overflow = 'hidden';
    window.removeEventListener('message', eventListener);
  };
  window.addEventListener('message', eventListener);
}
