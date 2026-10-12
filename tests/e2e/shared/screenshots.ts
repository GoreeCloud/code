import {expect, type Page, type Locator} from '@playwright/test';

// returns element that should be covered before taking the screenshot
async function masks(page: Page) : Promise<Locator[]> {
  return [
    page.locator('.ui.avatar'),
    page.locator('.sha'),
    page.locator('#repo_migrating'),
    // update order of recently created repos is not fully deterministic
    page.locator('.flex-item-main').filter({hasText: 'relative time in repo'}),
    page.locator('#activity-feed'),
    page.locator('#user-heatmap'),
    // dynamic IDs in fixed-size inputs
    page.locator('input[value*="dyn-id-"]'),
  ];
}

// replaces elements on the page that cause flakiness
async function screenshot_prepare(page: Page) {
  await page.waitForLoadState('domcontentloaded');
  // Version string is dynamic
  await page.locator('footer .left-links').evaluate((node) => node.innerHTML = 'MOCK');

  // replace timestamps in repos to mask them later down
  await page.locator('.flex-item-body > relative-time').filter({hasText: /now|minute/}).evaluateAll((nodes) => {
    for (const node of nodes) node.outerHTML = 'relative time in repo';
  });
  // other time elements
  await page.locator('relative-time').evaluateAll((nodes) => {
    for (const node of nodes) node.outerHTML = 'time element';
  });
  await page.locator('absolute-date').evaluateAll((nodes) => {
    for (const node of nodes) node.outerHTML = 'time element';
  });

  // dynamically generated UUIDs
  await page.getByText('dyn-id-').evaluateAll((nodes) => {
    for (const node of nodes) node.innerHTML = node.innerHTML.replaceAll(/dyn-id-[a-f0-9-]+/g, 'dynamic-id');
  });
  // repeat above, work around https://github.com/microsoft/playwright/issues/34152
  await page.getByText('dyn-id-').evaluateAll((nodes) => {
    for (const node of nodes) node.innerHTML = node.innerHTML.replaceAll(/dyn-id-[a-f0-9-]+/g, 'dynamic-id');
  });

  // attachment IDs in text areas, required for issue-comment-dropzone.
  // playwright does not (yet?) support filtering for content in input elements, see https://github.com/microsoft/playwright/issues/36166
  await page.locator('textarea.markdown-text-editor').evaluateAll((nodes: HTMLTextAreaElement[]) => {
    for (const node of nodes) node.value = node.value.replaceAll(/attachments\/[a-f0-9-]+/g, '/attachments/c1ee9740-dad3-4747-b489-f6fb2e3dfcec');
  });

  // dynamically created test users
  await page.getByText('e2e-test-').evaluateAll((nodes) => {
    for (const node of nodes) node.innerHTML = node.innerHTML.replaceAll(/e2e-test-[0-9-]+/g, 'e2e-test-user');
  });
}

export async function screenshot(page: Page, locator?: Locator, margin = 0) {
  // Optionally include visual testing
  if (process.env.VISUAL_TEST) {
    await screenshot_prepare(page);
    if (locator === undefined) {
      await screenshot_full(page);
    } else {
      await screenshot_selective(page, locator, margin);
    }
  }
}

async function screenshot_selective(page: Page, locator: Locator, margin: number) {
  const clip = await locator.boundingBox();
  clip.x = Math.max(clip.x - margin, 0);
  clip.y = Math.max(clip.y - margin, 0);
  clip.width += margin * 2;
  clip.height += margin * 2;
  await expect(page).toHaveScreenshot({
    fullPage: true,
    timeout: 20000,
    clip,
    mask: await masks(page),
  });
}

async function screenshot_full(page: Page) {
  await expect(page).toHaveScreenshot({
    fullPage: true,
    timeout: 20000,
    mask: await masks(page),
  });
}
